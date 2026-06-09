//! Async HTTP abstractions for DidaCLI API surfaces.

mod checksum;
mod client;
mod download;
mod error;
mod mcp;
mod openapi;
mod policy;
mod transport;
mod upgrade;
mod webapi;

pub use checksum::{sha256_hex, verify_sha256_from_checksums};
pub use client::{ApiSurface, JsonClient};
pub use download::{DownloadOptions, download_bounded};
pub use error::{DidaHttpError, Result};
pub use mcp::{
    INIT_PROTOCOL_VERSION, MCP_PROTOCOL_VERSION, McpClient, McpClientConfig, McpRpcError, McpTool,
};
pub use openapi::{OpenApiClient, OpenApiClientConfig};
pub use policy::{RetryDecision, RetryPolicy, TimeoutPolicy};
pub use transport::{HttpRequest, HttpResponse, HttpTransport};
pub use upgrade::{GitHubAsset, GitHubRelease, UpgradeClient, UpgradeClientConfig};
pub use webapi::{
    DEFAULT_WEB_API_BASE_V1, DEFAULT_WEB_API_BASE_V2, WebApiClient, WebApiClientConfig,
};
