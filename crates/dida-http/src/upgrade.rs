use std::sync::Arc;

use http::{HeaderMap, HeaderValue, Method, header};
use serde::Deserialize;

use crate::{
    checksum::verify_sha256_from_checksums,
    client::{ApiSurface, JsonClient, join_url},
    download::{DownloadOptions, download_bounded},
    error::Result,
    policy::{RetryPolicy, TimeoutPolicy},
    transport::HttpTransport,
};

#[derive(Clone, Debug)]
pub struct UpgradeClientConfig {
    pub github_api_base: String,
    pub repository: String,
    pub user_agent: String,
    pub metadata_retry_policy: RetryPolicy,
    pub metadata_timeout_policy: TimeoutPolicy,
    pub artifact_retry_policy: RetryPolicy,
    pub artifact_timeout_policy: TimeoutPolicy,
    pub metadata_max_response_bytes: u64,
    pub artifact_max_bytes: u64,
}

impl Default for UpgradeClientConfig {
    fn default() -> Self {
        Self {
            github_api_base: "https://api.github.com".to_owned(),
            repository: "DeliciousBuding/dida-cli".to_owned(),
            user_agent: "DidaCLI/dev".to_owned(),
            metadata_retry_policy: RetryPolicy::default(),
            metadata_timeout_policy: TimeoutPolicy::default(),
            artifact_retry_policy: RetryPolicy::default(),
            artifact_timeout_policy: TimeoutPolicy::new(std::time::Duration::from_secs(120)),
            metadata_max_response_bytes: 1 << 20,
            artifact_max_bytes: 200 << 20,
        }
    }
}

#[derive(Clone, Debug, Deserialize, PartialEq, Eq)]
pub struct GitHubRelease {
    #[serde(rename = "tag_name")]
    pub tag_name: String,
    #[serde(default)]
    pub assets: Vec<GitHubAsset>,
}

#[derive(Clone, Debug, Deserialize, PartialEq, Eq)]
pub struct GitHubAsset {
    pub name: String,
    #[serde(rename = "browser_download_url")]
    pub browser_download_url: String,
}

pub struct UpgradeClient<T> {
    metadata_client: JsonClient<T>,
    artifact_client: JsonClient<T>,
    config: UpgradeClientConfig,
}

impl<T: HttpTransport> UpgradeClient<T> {
    pub fn new(transport: Arc<T>, config: UpgradeClientConfig) -> Self {
        let metadata_client = JsonClient::new(transport.clone(), ApiSurface::Upgrade)
            .with_retry_policy(config.metadata_retry_policy.clone())
            .with_timeout_policy(config.metadata_timeout_policy)
            .with_max_response_bytes(config.metadata_max_response_bytes);
        let artifact_client = JsonClient::new(transport, ApiSurface::Upgrade)
            .with_retry_policy(config.artifact_retry_policy.clone())
            .with_timeout_policy(config.artifact_timeout_policy)
            .with_max_response_bytes(config.artifact_max_bytes);
        Self {
            metadata_client,
            artifact_client,
            config,
        }
    }

    pub async fn latest_release(&self) -> Result<GitHubRelease> {
        let url = join_url(
            &self.config.github_api_base,
            &format!("/repos/{}/releases/latest", self.config.repository),
        );
        let mut headers = HeaderMap::new();
        headers.insert(
            header::ACCEPT,
            HeaderValue::from_static("application/vnd.github+json"),
        );
        headers.insert(
            header::USER_AGENT,
            HeaderValue::from_str(&self.config.user_agent)
                .map_err(|error| crate::error::DidaHttpError::Other(error.to_string()))?,
        );
        self.metadata_client
            .json::<(), GitHubRelease>(Method::GET, &url, headers, None)
            .await
    }

    pub async fn download_asset(&self, asset: &GitHubAsset) -> Result<Vec<u8>> {
        download_bounded(
            &self.artifact_client,
            &asset.browser_download_url,
            DownloadOptions {
                max_bytes: self.config.artifact_max_bytes,
                accept: Some("application/octet-stream".to_owned()),
            },
        )
        .await
    }

    pub async fn download_and_verify(
        &self,
        asset: &GitHubAsset,
        checksums_asset: &GitHubAsset,
    ) -> Result<Vec<u8>> {
        let data = self.download_asset(asset).await?;
        let checksums = self.download_asset(checksums_asset).await?;
        let checksums = String::from_utf8_lossy(&checksums);
        verify_sha256_from_checksums(&data, &checksums, &asset.name)?;
        Ok(data)
    }
}
