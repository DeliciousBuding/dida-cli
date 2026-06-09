use std::sync::Arc;

use http::{HeaderMap, HeaderValue, Method, header};
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

use crate::{
    client::{ApiSurface, JsonClient, bearer_headers, join_url},
    error::{DidaHttpError, Result},
    policy::{RetryPolicy, TimeoutPolicy},
    transport::HttpTransport,
};

pub const MCP_PROTOCOL_VERSION: &str = "2024-11-05";
pub const INIT_PROTOCOL_VERSION: &str = "2025-03-26";

#[derive(Clone, Debug)]
pub struct McpClientConfig {
    pub url: String,
    pub token: String,
    pub session_id: Option<String>,
    pub retry_policy: RetryPolicy,
    pub timeout_policy: TimeoutPolicy,
    pub max_response_bytes: u64,
}

impl McpClientConfig {
    pub fn new(token: impl Into<String>) -> Self {
        Self {
            url: "https://mcp.dida365.com".to_owned(),
            token: token.into(),
            session_id: None,
            retry_policy: RetryPolicy::default(),
            timeout_policy: TimeoutPolicy::new(std::time::Duration::from_secs(60)),
            max_response_bytes: 16 << 20,
        }
    }
}

#[derive(Debug, Clone, Deserialize, PartialEq, Eq)]
pub struct McpTool {
    pub name: String,
    #[serde(default)]
    pub description: Option<String>,
    #[serde(rename = "inputSchema", default)]
    pub input_schema: Option<Value>,
    #[serde(rename = "outputSchema", default)]
    pub output_schema: Option<Value>,
    #[serde(default)]
    pub annotations: Option<Value>,
}

#[derive(Debug, Clone, Deserialize, PartialEq, Eq)]
pub struct McpRpcError {
    pub code: Option<Value>,
    pub message: String,
}

#[derive(Debug, Deserialize)]
struct RpcResponse {
    #[serde(default)]
    result: Option<Value>,
    #[serde(default)]
    error: Option<McpRpcError>,
}

#[derive(Debug, Deserialize)]
struct ToolList {
    #[serde(default)]
    tools: Vec<McpTool>,
}

pub struct McpClient<T> {
    client: JsonClient<T>,
    config: McpClientConfig,
}

impl<T: HttpTransport> McpClient<T> {
    pub fn new(transport: Arc<T>, config: McpClientConfig) -> Self {
        let client = JsonClient::new(transport, ApiSurface::OfficialMcp)
            .with_retry_policy(config.retry_policy.clone())
            .with_timeout_policy(config.timeout_policy)
            .with_max_response_bytes(config.max_response_bytes);
        Self { client, config }
    }

    pub async fn initialize(&self, client_name: &str, client_version: &str) -> Result<()> {
        let payload = json!({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": INIT_PROTOCOL_VERSION,
                "clientInfo": {
                    "name": client_name,
                    "version": client_version,
                },
                "capabilities": {}
            }
        });
        self.post_rpc(payload, false).await?;
        self.notify_initialized().await
    }

    pub async fn notify_initialized(&self) -> Result<()> {
        let payload = json!({
            "jsonrpc": "2.0",
            "method": "notifications/initialized",
            "params": {}
        });
        self.post_rpc(payload, true).await.map(|_| ())
    }

    pub async fn tools(&self) -> Result<Vec<McpTool>> {
        let payload = json!({
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        });
        let result = self.post_rpc(payload, true).await?;
        let list: ToolList = serde_json::from_value(result).map_err(DidaHttpError::DecodeJson)?;
        Ok(list.tools)
    }

    pub async fn call_tool<A: Serialize + Sync>(&self, name: &str, arguments: &A) -> Result<Value> {
        let payload = json!({
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": name,
                "arguments": arguments
            }
        });
        let result = self.post_rpc(payload, true).await?;
        Ok(unwrap_tool_result(result))
    }

    async fn post_rpc(&self, payload: Value, include_protocol: bool) -> Result<Value> {
        if self.config.token.trim().is_empty() {
            return Err(DidaHttpError::MissingCredential {
                surface: "official mcp",
            });
        }
        let mut headers = bearer_headers(&self.config.token)?;
        headers.insert(
            header::CONTENT_TYPE,
            HeaderValue::from_static("application/json"),
        );
        headers.insert(
            header::ACCEPT,
            HeaderValue::from_static("application/json, text/event-stream"),
        );
        if include_protocol {
            headers.insert(
                "MCP-Protocol-Version",
                HeaderValue::from_static(MCP_PROTOCOL_VERSION),
            );
        }
        if let Some(session_id) = self.config.session_id.as_deref() {
            headers.insert(
                "Mcp-Session-Id",
                HeaderValue::from_str(session_id).map_err(|error| {
                    DidaHttpError::Other(format!("invalid session id: {error}"))
                })?,
            );
        }
        let response = self
            .client
            .request_raw(
                Method::POST,
                &self.config.url,
                headers,
                serde_json::to_vec(&payload).map_err(DidaHttpError::EncodeJson)?,
                Some(self.config.max_response_bytes),
            )
            .await?;

        if response.body.is_empty() && payload.get("id").is_none() {
            return Ok(Value::Null);
        }
        let rpc = parse_rpc_response(&response.headers, &response.body)?;
        if let Some(error) = rpc.error {
            return Err(DidaHttpError::McpRpc {
                code: error.code,
                message: error.message,
            });
        }
        Ok(rpc.result.unwrap_or(Value::Null))
    }
}

fn parse_rpc_response(headers: &HeaderMap, body: &[u8]) -> Result<RpcResponse> {
    if headers
        .get(header::CONTENT_TYPE)
        .and_then(|value| value.to_str().ok())
        .is_some_and(|value| value.contains("text/event-stream"))
    {
        return parse_sse_response(body);
    }
    serde_json::from_slice(body).map_err(DidaHttpError::DecodeJson)
}

fn parse_sse_response(body: &[u8]) -> Result<RpcResponse> {
    let text = String::from_utf8_lossy(body);
    let mut buffer = Vec::new();
    for line in text.lines() {
        if line.trim().is_empty() {
            if !buffer.is_empty() {
                let payload = buffer.join("\n");
                return serde_json::from_str(&payload).map_err(DidaHttpError::DecodeJson);
            }
            continue;
        }
        if let Some(data) = line.strip_prefix("data:") {
            buffer.push(data.trim().to_owned());
        }
    }
    if !buffer.is_empty() {
        let payload = buffer.join("\n");
        return serde_json::from_str(&payload).map_err(DidaHttpError::DecodeJson);
    }
    Err(DidaHttpError::Other(
        "empty official mcp response".to_owned(),
    ))
}

fn unwrap_tool_result(result: Value) -> Value {
    if let Some(value) = result.get("structuredContent") {
        return unwrap_envelope(value.clone());
    }
    if let Some(content) = result.get("content").and_then(Value::as_array) {
        if content.len() == 1 {
            if let Some(text) = content[0].get("text").and_then(Value::as_str) {
                if let Ok(decoded) = serde_json::from_str::<Value>(text) {
                    return unwrap_envelope(decoded);
                }
                return Value::String(text.to_owned());
            }
        }
    }
    unwrap_envelope(result)
}

fn unwrap_envelope(mut value: Value) -> Value {
    loop {
        let Some(object) = value.as_object() else {
            return value;
        };
        if object.len() != 1 {
            return value;
        }
        if let Some(next) = object.get("result").or_else(|| object.get("results")) {
            value = next.clone();
            continue;
        }
        return value;
    }
}

#[allow(dead_code)]
fn _join_mcp_url(base: &str, path: &str) -> String {
    join_url(base, path)
}
