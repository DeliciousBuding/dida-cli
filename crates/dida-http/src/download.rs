use http::{HeaderMap, Method, header};

use crate::{
    client::JsonClient,
    error::{DidaHttpError, Result},
    transport::HttpTransport,
};

#[derive(Clone, Debug)]
pub struct DownloadOptions {
    pub max_bytes: u64,
    pub accept: Option<String>,
}

impl Default for DownloadOptions {
    fn default() -> Self {
        Self {
            max_bytes: 200 << 20,
            accept: Some("application/octet-stream".to_owned()),
        }
    }
}

pub async fn download_bounded<T: HttpTransport>(
    client: &JsonClient<T>,
    url: &str,
    options: DownloadOptions,
) -> Result<Vec<u8>> {
    let mut headers = HeaderMap::new();
    if let Some(accept) = options.accept.as_deref() {
        headers.insert(
            header::ACCEPT,
            accept
                .parse()
                .map_err(|error| DidaHttpError::Other(format!("invalid Accept header: {error}")))?,
        );
    }

    let response = client
        .request_raw(
            Method::GET,
            url,
            headers,
            Vec::new(),
            Some(options.max_bytes),
        )
        .await?;
    if let Some(_length) = response
        .headers
        .get(header::CONTENT_LENGTH)
        .and_then(|value| value.to_str().ok())
        .and_then(|value| value.parse::<u64>().ok())
        .filter(|length| *length > options.max_bytes)
    {
        return Err(DidaHttpError::DownloadTooLarge {
            max_bytes: options.max_bytes,
        });
    }
    if response.body.len() as u64 > options.max_bytes {
        return Err(DidaHttpError::DownloadTooLarge {
            max_bytes: options.max_bytes,
        });
    }
    Ok(response.body.to_vec())
}
