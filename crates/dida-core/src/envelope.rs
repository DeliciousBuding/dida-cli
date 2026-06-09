use serde::Serialize;
use serde_json::Value;

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct CliError {
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub kind: Option<String>,
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub hint: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub details: Option<Value>,
}

impl CliError {
    pub fn validation(command: &str, message: impl Into<String>) -> Self {
        let _ = command;
        Self {
            kind: Some("validation".to_string()),
            message: message.into(),
            hint: None,
            details: None,
        }
    }

    pub fn with_hint(mut self, hint: impl Into<String>) -> Self {
        self.hint = Some(hint.into());
        self
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct JsonEnvelope {
    pub ok: bool,
    pub command: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub meta: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<CliError>,
}

pub fn success(command: impl Into<String>, data: Value) -> JsonEnvelope {
    JsonEnvelope {
        ok: true,
        command: command.into(),
        meta: None,
        data: Some(data),
        error: None,
    }
}

pub fn failure(error: CliError) -> JsonEnvelope {
    JsonEnvelope {
        ok: false,
        command: "dida".to_string(),
        meta: None,
        data: None,
        error: Some(error),
    }
}

pub fn to_json_line(envelope: &JsonEnvelope) -> serde_json::Result<String> {
    let mut line = serde_json::to_string_pretty(envelope)?;
    line.push('\n');
    Ok(line)
}
