use std::sync::Arc;

use http::{HeaderMap, HeaderValue, Method, header};
use serde::{Serialize, de::DeserializeOwned};

use crate::{
    client::{ApiSurface, JsonClient, join_url},
    error::{DidaHttpError, Result},
    policy::{RetryPolicy, TimeoutPolicy},
    transport::HttpTransport,
};

pub const DEFAULT_WEB_API_BASE_V2: &str = "https://api.dida365.com/api/v2";
pub const DEFAULT_WEB_API_BASE_V1: &str = "https://api.dida365.com/api/v1";

#[derive(Clone, Debug)]
pub struct WebApiClientConfig {
    pub base_url_v2: String,
    pub base_url_v1: String,
    pub cookie_token: String,
    pub user_agent: String,
    pub device_id: String,
    pub retry_policy: RetryPolicy,
    pub timeout_policy: TimeoutPolicy,
    pub max_response_bytes: u64,
}

impl WebApiClientConfig {
    pub fn new(cookie_token: impl Into<String>) -> Self {
        Self {
            base_url_v2: DEFAULT_WEB_API_BASE_V2.to_owned(),
            base_url_v1: DEFAULT_WEB_API_BASE_V1.to_owned(),
            cookie_token: cookie_token.into(),
            user_agent:
                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0"
                    .to_owned(),
            device_id: "649000000000000000000000".to_owned(),
            retry_policy: RetryPolicy::default(),
            timeout_policy: TimeoutPolicy::default(),
            max_response_bytes: 16 << 20,
        }
    }
}

pub struct WebApiClient<T> {
    client: JsonClient<T>,
    config: WebApiClientConfig,
}

impl<T: HttpTransport> WebApiClient<T> {
    pub fn new(transport: Arc<T>, config: WebApiClientConfig) -> Self {
        let client = JsonClient::new(transport, ApiSurface::WebApi)
            .with_retry_policy(config.retry_policy.clone())
            .with_timeout_policy(config.timeout_policy)
            .with_max_response_bytes(config.max_response_bytes);
        Self { client, config }
    }

    pub async fn json<I, O>(&self, method: Method, path: &str, body: Option<&I>) -> Result<O>
    where
        I: Serialize + Sync,
        O: DeserializeOwned,
    {
        let url = join_url(&self.config.base_url_v2, path);
        self.client.json(method, &url, self.headers()?, body).await
    }

    pub async fn json_v1<I, O>(&self, method: Method, path: &str, body: Option<&I>) -> Result<O>
    where
        I: Serialize + Sync,
        O: DeserializeOwned,
    {
        let url = join_url(&self.config.base_url_v1, path);
        self.client.json(method, &url, self.headers()?, body).await
    }

    fn headers(&self) -> Result<HeaderMap> {
        if self.config.cookie_token.trim().is_empty() {
            return Err(DidaHttpError::MissingCredential { surface: "web api" });
        }
        let mut headers = HeaderMap::new();
        headers.insert(
            header::COOKIE,
            HeaderValue::from_str(&format!("t={}", self.config.cookie_token))
                .map_err(|error| DidaHttpError::Other(format!("invalid cookie header: {error}")))?,
        );
        headers.insert(
            header::USER_AGENT,
            HeaderValue::from_str(&self.config.user_agent).map_err(|error| {
                DidaHttpError::Other(format!("invalid user-agent header: {error}"))
            })?,
        );
        headers.insert(
            "x-device",
            HeaderValue::from_str(&self.device_header()).map_err(|error| {
                DidaHttpError::Other(format!("invalid x-device header: {error}"))
            })?,
        );
        Ok(headers)
    }

    fn device_header(&self) -> String {
        serde_json::json!({
            "platform": "web",
            "os": "OS X",
            "device": "Firefox 95.0",
            "name": "DidaCLI",
            "version": 4531,
            "id": self.config.device_id,
            "channel": "website",
            "campaign": "",
            "websocket": ""
        })
        .to_string()
    }
}
