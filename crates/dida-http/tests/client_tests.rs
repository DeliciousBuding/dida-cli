use std::{
    collections::VecDeque,
    sync::{Arc, Mutex},
    time::Duration,
};

use async_trait::async_trait;
use bytes::Bytes;
use dida_http::{
    DidaHttpError, GitHubAsset, HttpRequest, HttpResponse, HttpTransport, McpClient,
    McpClientConfig, OpenApiClient, OpenApiClientConfig, RetryPolicy, TimeoutPolicy, UpgradeClient,
    UpgradeClientConfig, WebApiClient, WebApiClientConfig, sha256_hex,
    verify_sha256_from_checksums,
};
use http::{HeaderMap, HeaderValue, Method, StatusCode, header};
use serde_json::{Value, json};

#[derive(Default)]
struct MockTransport {
    calls: Mutex<Vec<HttpRequest>>,
    responses: Mutex<VecDeque<HttpResponse>>,
}

impl MockTransport {
    fn with_responses(responses: Vec<HttpResponse>) -> Arc<Self> {
        Arc::new(Self {
            calls: Mutex::new(Vec::new()),
            responses: Mutex::new(responses.into()),
        })
    }

    fn calls(&self) -> Vec<HttpRequest> {
        self.calls.lock().unwrap().clone()
    }
}

#[async_trait]
impl HttpTransport for MockTransport {
    async fn execute(&self, request: HttpRequest) -> dida_http::Result<HttpResponse> {
        self.calls.lock().unwrap().push(request);
        self.responses
            .lock()
            .unwrap()
            .pop_front()
            .ok_or_else(|| DidaHttpError::transport("no mock response queued"))
    }
}

fn json_response(value: Value) -> HttpResponse {
    HttpResponse {
        status: StatusCode::OK,
        headers: HeaderMap::new(),
        body: Bytes::from(serde_json::to_vec(&value).unwrap()),
    }
}

fn response(status: StatusCode, body: impl Into<Bytes>) -> HttpResponse {
    HttpResponse {
        status,
        headers: HeaderMap::new(),
        body: body.into(),
    }
}

#[tokio::test]
async fn web_api_builds_cookie_device_json_request() {
    let transport = MockTransport::with_responses(vec![json_response(json!({"ok": true}))]);
    let mut config = WebApiClientConfig::new("cookie-token");
    config.base_url_v2 = "https://example.invalid/api/v2".to_owned();
    config.device_id = "device-1".to_owned();
    config.retry_policy = RetryPolicy::none();
    let client = WebApiClient::new(transport.clone(), config);

    let out: Value = client
        .json(Method::POST, "/batch/check/0", Some(&json!({"x": 1})))
        .await
        .unwrap();

    assert_eq!(out, json!({"ok": true}));
    let calls = transport.calls();
    assert_eq!(calls.len(), 1);
    assert_eq!(calls[0].method, Method::POST);
    assert_eq!(
        calls[0].uri.to_string(),
        "https://example.invalid/api/v2/batch/check/0"
    );
    assert_eq!(
        calls[0].headers.get(header::COOKIE).unwrap(),
        &HeaderValue::from_static("t=cookie-token")
    );
    assert!(
        calls[0]
            .headers
            .get("x-device")
            .unwrap()
            .to_str()
            .unwrap()
            .contains("device-1")
    );
    assert_eq!(&calls[0].body[..], br#"{"x":1}"#);
}

#[tokio::test]
async fn openapi_builds_bearer_request() {
    let transport = MockTransport::with_responses(vec![json_response(json!({"id": "p1"}))]);
    let mut config = OpenApiClientConfig::new("access-token");
    config.base_url = "https://open.example.invalid".to_owned();
    config.retry_policy = RetryPolicy::none();
    let client = OpenApiClient::new(transport.clone(), config);

    let out: Value = client
        .json::<(), Value>(Method::GET, "/project/p1", None)
        .await
        .unwrap();

    assert_eq!(out, json!({"id": "p1"}));
    let calls = transport.calls();
    assert_eq!(
        calls[0].uri.to_string(),
        "https://open.example.invalid/project/p1"
    );
    assert_eq!(
        calls[0].headers.get(header::AUTHORIZATION).unwrap(),
        &HeaderValue::from_static("Bearer access-token")
    );
}

#[tokio::test]
async fn json_client_retries_retryable_status() {
    let transport = MockTransport::with_responses(vec![
        response(StatusCode::SERVICE_UNAVAILABLE, "busy"),
        json_response(json!({"ok": true})),
    ]);
    let mut config = OpenApiClientConfig::new("access-token");
    config.base_url = "https://open.example.invalid".to_owned();
    config.retry_policy = RetryPolicy {
        max_retries: 1,
        base_delay: Duration::from_millis(1),
        max_delay: Duration::from_millis(1),
        retry_429: true,
        retry_server_errors: true,
    };
    config.timeout_policy = TimeoutPolicy::new(Duration::from_secs(1));
    let client = OpenApiClient::new(transport.clone(), config);

    let out: Value = client
        .json::<(), Value>(Method::GET, "/project", None)
        .await
        .unwrap();

    assert_eq!(out, json!({"ok": true}));
    assert_eq!(transport.calls().len(), 2);
}

#[tokio::test]
async fn mcp_parses_sse_tool_list() {
    let mut headers = HeaderMap::new();
    headers.insert(
        header::CONTENT_TYPE,
        HeaderValue::from_static("text/event-stream"),
    );
    let transport = MockTransport::with_responses(vec![HttpResponse {
        status: StatusCode::OK,
        headers,
        body: Bytes::from_static(
            br#"event: message
data: {"result":{"tools":[{"name":"tasks.list","description":"List tasks"}]}}

"#,
        ),
    }]);
    let mut config = McpClientConfig::new("mcp-token");
    config.url = "https://mcp.example.invalid".to_owned();
    config.retry_policy = RetryPolicy::none();
    let client = McpClient::new(transport.clone(), config);

    let tools = client.tools().await.unwrap();

    assert_eq!(tools.len(), 1);
    assert_eq!(tools[0].name, "tasks.list");
    let calls = transport.calls();
    assert_eq!(
        calls[0].headers.get("MCP-Protocol-Version").unwrap(),
        &HeaderValue::from_static("2024-11-05")
    );
    assert_eq!(
        calls[0].headers.get(header::ACCEPT).unwrap(),
        &HeaderValue::from_static("application/json, text/event-stream")
    );
}

#[tokio::test]
async fn upgrade_download_rejects_oversized_content_length() {
    let mut headers = HeaderMap::new();
    headers.insert(header::CONTENT_LENGTH, HeaderValue::from_static("11"));
    let transport = MockTransport::with_responses(vec![HttpResponse {
        status: StatusCode::OK,
        headers,
        body: Bytes::from_static(b"small"),
    }]);
    let config = UpgradeClientConfig {
        github_api_base: "https://api.example.invalid".to_owned(),
        repository: "owner/repo".to_owned(),
        artifact_max_bytes: 10,
        artifact_retry_policy: RetryPolicy::none(),
        ..UpgradeClientConfig::default()
    };
    let client = UpgradeClient::new(transport, config);
    let asset = GitHubAsset {
        name: "dida.zip".to_owned(),
        browser_download_url: "https://download.example.invalid/dida.zip".to_owned(),
    };

    let err = client.download_asset(&asset).await.unwrap_err();

    assert!(matches!(
        err,
        DidaHttpError::DownloadTooLarge { max_bytes: 10 }
    ));
}

#[tokio::test]
async fn checksum_verification_matches_named_archive() {
    let data = b"archive";
    let checksums = format!("{}  dida.tar.gz\n", sha256_hex(data));

    verify_sha256_from_checksums(data, &checksums, "dida.tar.gz").unwrap();

    let err = verify_sha256_from_checksums(b"bad", &checksums, "dida.tar.gz").unwrap_err();
    assert!(matches!(err, DidaHttpError::ChecksumMismatch { .. }));
}
