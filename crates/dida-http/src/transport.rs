use async_trait::async_trait;
use bytes::Bytes;
use http::{HeaderMap, Method, StatusCode, Uri};

use crate::error::Result;

#[derive(Clone, Debug)]
pub struct HttpRequest {
    pub method: Method,
    pub uri: Uri,
    pub headers: HeaderMap,
    pub body: Bytes,
    pub max_response_bytes: Option<u64>,
}

#[derive(Clone, Debug)]
pub struct HttpResponse {
    pub status: StatusCode,
    pub headers: HeaderMap,
    pub body: Bytes,
}

#[async_trait]
pub trait HttpTransport: Send + Sync + 'static {
    async fn execute(&self, request: HttpRequest) -> Result<HttpResponse>;
}

#[cfg(feature = "reqwest-client")]
pub struct ReqwestTransport {
    inner: reqwest::Client,
}

#[cfg(feature = "reqwest-client")]
impl ReqwestTransport {
    pub fn new(inner: reqwest::Client) -> Self {
        Self { inner }
    }
}

#[cfg(feature = "reqwest-client")]
#[async_trait]
impl HttpTransport for ReqwestTransport {
    async fn execute(&self, request: HttpRequest) -> Result<HttpResponse> {
        use crate::error::DidaHttpError;
        use bytes::BytesMut;
        use futures_util::StreamExt;

        let mut builder = self
            .inner
            .request(request.method.clone(), request.uri.to_string())
            .headers(request.headers.clone());
        if !request.body.is_empty() {
            builder = builder.body(request.body);
        }
        let response = builder
            .send()
            .await
            .map_err(|error| DidaHttpError::transport(error.to_string()))?;

        let status = response.status();
        let headers = response.headers().clone();
        let mut body = BytesMut::new();
        let max = request.max_response_bytes.unwrap_or(u64::MAX);
        let mut stream = response.bytes_stream();
        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|error| DidaHttpError::transport(error.to_string()))?;
            if body.len() as u64 + chunk.len() as u64 > max {
                return Err(DidaHttpError::ResponseTooLarge { max_bytes: max });
            }
            body.extend_from_slice(&chunk);
        }
        Ok(HttpResponse {
            status,
            headers,
            body: body.freeze(),
        })
    }
}
