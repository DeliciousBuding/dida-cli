use std::time::Duration;

use thiserror::Error;

pub type Result<T> = std::result::Result<T, DidaHttpError>;

#[derive(Debug, Error)]
pub enum DidaHttpError {
    #[error("missing credential for {surface}")]
    MissingCredential { surface: &'static str },
    #[error("invalid URL {url}: {source}")]
    InvalidUrl {
        url: String,
        source: http::uri::InvalidUri,
    },
    #[error("build HTTP request: {0}")]
    BuildRequest(#[from] http::Error),
    #[error("transport error: {0}")]
    Transport(String),
    #[error("{surface} {method} {path} returned HTTP {status}: {body}")]
    HttpStatus {
        surface: &'static str,
        method: String,
        path: String,
        status: u16,
        body: String,
    },
    #[error("encode JSON request: {0}")]
    EncodeJson(#[source] serde_json::Error),
    #[error("decode JSON response: {0}")]
    DecodeJson(#[source] serde_json::Error),
    #[error("response exceeded {max_bytes} bytes")]
    ResponseTooLarge { max_bytes: u64 },
    #[error("download exceeded {max_bytes} bytes")]
    DownloadTooLarge { max_bytes: u64 },
    #[error("request timed out after {timeout:?}")]
    Timeout { timeout: Duration },
    #[error("checksum mismatch for {name}: got {actual}, want {expected}")]
    ChecksumMismatch {
        name: String,
        actual: String,
        expected: String,
    },
    #[error("archive {name:?} not found in checksums")]
    ChecksumMissing { name: String },
    #[error("official MCP RPC error {code:?}: {message}")]
    McpRpc {
        code: Option<serde_json::Value>,
        message: String,
    },
    #[error("{0}")]
    Other(String),
}

impl DidaHttpError {
    pub fn transport(error: impl Into<String>) -> Self {
        Self::Transport(error.into())
    }
}
