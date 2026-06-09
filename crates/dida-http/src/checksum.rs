use sha2::{Digest, Sha256};

use crate::error::{DidaHttpError, Result};

pub fn sha256_hex(data: &[u8]) -> String {
    let digest = Sha256::digest(data);
    format!("{digest:x}")
}

pub fn verify_sha256_from_checksums(data: &[u8], checksums: &str, name: &str) -> Result<()> {
    for line in checksums.lines() {
        let mut fields = line.split_whitespace();
        let Some(expected) = fields.next() else {
            continue;
        };
        let Some(candidate_name) = fields.next() else {
            continue;
        };
        let candidate_name = candidate_name.trim_start_matches('*');
        if candidate_name == name {
            let actual = sha256_hex(data);
            if actual == expected {
                return Ok(());
            }
            return Err(DidaHttpError::ChecksumMismatch {
                name: name.to_owned(),
                actual,
                expected: expected.to_owned(),
            });
        }
    }
    Err(DidaHttpError::ChecksumMissing {
        name: name.to_owned(),
    })
}
