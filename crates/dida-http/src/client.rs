use std::sync::Arc;

use bytes::Bytes;
use http::{HeaderMap, HeaderValue, Method, Uri, header};
use serde::{Serialize, de::DeserializeOwned};
use tokio::time;

use crate::{
    error::{DidaHttpError, Result},
    policy::{RetryDecision, RetryPolicy, TimeoutPolicy},
    transport::{HttpRequest, HttpResponse, HttpTransport},
};

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum ApiSurface {
    WebApi,
    OfficialMcp,
    OpenApi,
    Upgrade,
}

impl ApiSurface {
    pub const fn label(self) -> &'static str {
        match self {
            Self::WebApi => "dida web api",
            Self::OfficialMcp => "official mcp",
            Self::OpenApi => "openapi",
            Self::Upgrade => "upgrade",
        }
    }
}

#[derive(Clone)]
pub struct JsonClient<T> {
    transport: Arc<T>,
    surface: ApiSurface,
    retry_policy: RetryPolicy,
    timeout_policy: TimeoutPolicy,
    max_response_bytes: u64,
}

impl<T: HttpTransport> JsonClient<T> {
    pub fn new(transport: Arc<T>, surface: ApiSurface) -> Self {
        Self {
            transport,
            surface,
            retry_policy: RetryPolicy::default(),
            timeout_policy: TimeoutPolicy::default(),
            max_response_bytes: 16 << 20,
        }
    }

    pub fn with_retry_policy(mut self, policy: RetryPolicy) -> Self {
        self.retry_policy = policy;
        self
    }

    pub fn with_timeout_policy(mut self, policy: TimeoutPolicy) -> Self {
        self.timeout_policy = policy;
        self
    }

    pub fn with_max_response_bytes(mut self, max_response_bytes: u64) -> Self {
        self.max_response_bytes = max_response_bytes;
        self
    }

    pub async fn json<I, O>(
        &self,
        method: Method,
        url: &str,
        headers: HeaderMap,
        body: Option<&I>,
    ) -> Result<O>
    where
        I: Serialize + Sync,
        O: DeserializeOwned,
    {
        let mut headers = headers;
        headers.insert(header::ACCEPT, HeaderValue::from_static("application/json"));
        let body = match body {
            Some(value) => {
                headers.insert(
                    header::CONTENT_TYPE,
                    HeaderValue::from_static("application/json"),
                );
                serde_json::to_vec(value).map_err(DidaHttpError::EncodeJson)?
            }
            None => Vec::new(),
        };
        let response = self
            .request_raw(method, url, headers, body, Some(self.max_response_bytes))
            .await?;
        if response.body.is_empty() {
            serde_json::from_slice(b"null").map_err(DidaHttpError::DecodeJson)
        } else {
            serde_json::from_slice(&response.body).map_err(DidaHttpError::DecodeJson)
        }
    }

    pub async fn empty<I>(
        &self,
        method: Method,
        url: &str,
        headers: HeaderMap,
        body: Option<&I>,
    ) -> Result<()>
    where
        I: Serialize + Sync,
    {
        let _: serde_json::Value = self.json(method, url, headers, body).await?;
        Ok(())
    }

    pub async fn request_raw(
        &self,
        method: Method,
        url: &str,
        headers: HeaderMap,
        body: Vec<u8>,
        max_response_bytes: Option<u64>,
    ) -> Result<HttpResponse> {
        let uri = url
            .parse::<Uri>()
            .map_err(|source| DidaHttpError::InvalidUrl {
                url: url.to_owned(),
                source,
            })?;
        let path = uri
            .path_and_query()
            .map(|value| value.as_str().to_owned())
            .unwrap_or_else(|| "/".to_owned());
        let mut attempt = 0;
        loop {
            let request = HttpRequest {
                method: method.clone(),
                uri: uri.clone(),
                headers: headers.clone(),
                body: Bytes::from(body.clone()),
                max_response_bytes,
            };
            let fut = self.transport.execute(request);
            let response = time::timeout(self.timeout_policy.request_timeout, fut)
                .await
                .map_err(|_| DidaHttpError::Timeout {
                    timeout: self.timeout_policy.request_timeout,
                });
            let response = match response {
                Ok(Ok(response)) => response,
                Ok(Err(error)) => match self.retry_policy.classify_transport_error(attempt) {
                    RetryDecision::RetryAfter(delay) => {
                        attempt += 1;
                        time::sleep(delay).await;
                        let _ = error;
                        continue;
                    }
                    RetryDecision::DoNotRetry => return Err(error),
                },
                Err(error) => return Err(error),
            };

            if response.status.is_success() {
                return Ok(response);
            }
            let status = response.status.as_u16();
            match self.retry_policy.classify_status(status, attempt) {
                RetryDecision::RetryAfter(delay) => {
                    attempt += 1;
                    time::sleep(delay).await;
                }
                RetryDecision::DoNotRetry => {
                    return Err(DidaHttpError::HttpStatus {
                        surface: self.surface.label(),
                        method: method.to_string(),
                        path,
                        status,
                        body: summarize_body(&response.body),
                    });
                }
            }
        }
    }
}

fn summarize_body(body: &[u8]) -> String {
    let text = String::from_utf8_lossy(body).trim().to_owned();
    if text.len() > 500 {
        format!("{}...", &text[..500])
    } else {
        text
    }
}

pub(crate) fn join_url(base: &str, path: &str) -> String {
    let base = base.trim_end_matches('/');
    if path.starts_with("http://") || path.starts_with("https://") {
        return path.to_owned();
    }
    if path.starts_with('/') {
        format!("{base}{path}")
    } else {
        format!("{base}/{path}")
    }
}

pub(crate) fn bearer_headers(token: &str) -> Result<HeaderMap> {
    let mut headers = HeaderMap::new();
    let value = HeaderValue::from_str(&format!("Bearer {token}"))
        .map_err(|error| DidaHttpError::Other(format!("invalid bearer token header: {error}")))?;
    headers.insert(header::AUTHORIZATION, value);
    Ok(headers)
}
