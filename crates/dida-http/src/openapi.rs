use std::sync::Arc;

use http::Method;
use serde::{Serialize, de::DeserializeOwned};

use crate::{
    client::{ApiSurface, JsonClient, bearer_headers, join_url},
    error::{DidaHttpError, Result},
    policy::{RetryPolicy, TimeoutPolicy},
    transport::HttpTransport,
};

#[derive(Clone, Debug)]
pub struct OpenApiClientConfig {
    pub base_url: String,
    pub access_token: String,
    pub retry_policy: RetryPolicy,
    pub timeout_policy: TimeoutPolicy,
    pub max_response_bytes: u64,
}

impl OpenApiClientConfig {
    pub fn new(access_token: impl Into<String>) -> Self {
        Self {
            base_url: "https://api.dida365.com/open/v1".to_owned(),
            access_token: access_token.into(),
            retry_policy: RetryPolicy::default(),
            timeout_policy: TimeoutPolicy::default(),
            max_response_bytes: 16 << 20,
        }
    }
}

pub struct OpenApiClient<T> {
    client: JsonClient<T>,
    config: OpenApiClientConfig,
}

impl<T: HttpTransport> OpenApiClient<T> {
    pub fn new(transport: Arc<T>, config: OpenApiClientConfig) -> Self {
        let client = JsonClient::new(transport, ApiSurface::OpenApi)
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
        if self.config.access_token.trim().is_empty() {
            return Err(DidaHttpError::MissingCredential { surface: "openapi" });
        }
        let url = join_url(&self.config.base_url, path);
        self.client
            .json(
                method,
                &url,
                bearer_headers(&self.config.access_token)?,
                body,
            )
            .await
    }
}
