use std::error::Error;
use std::fmt::{self, Display};

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ParseValueError {
    message: String,
}

impl ParseValueError {
    fn new(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
        }
    }
}

impl Display for ParseValueError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.message)
    }
}

impl Error for ParseValueError {}

pub fn parse_i32_strict(value: &str) -> Result<i32, ParseValueError> {
    value
        .trim()
        .parse::<i32>()
        .map_err(|err| ParseValueError::new(err.to_string()))
}

pub fn parse_i64_strict(value: &str) -> Result<i64, ParseValueError> {
    value
        .trim()
        .parse::<i64>()
        .map_err(|err| ParseValueError::new(err.to_string()))
}

pub fn parse_f64_strict(value: &str) -> Result<f64, ParseValueError> {
    let parsed = value
        .trim()
        .parse::<f64>()
        .map_err(|err| ParseValueError::new(err.to_string()))?;
    if !parsed.is_finite() {
        return Err(ParseValueError::new("value must be finite"));
    }
    Ok(parsed)
}

pub fn validate_id_arg(name: &str, value: &str) -> Result<(), ParseValueError> {
    let trimmed = value.trim();
    if trimmed.is_empty() {
        return Err(ParseValueError::new(format!("{name} id is required")));
    }
    if trimmed.starts_with('-') {
        return Err(ParseValueError::new(format!(
            "{name} id must not start with '-'"
        )));
    }
    Ok(())
}

pub fn parse_id_value(
    args: &[String],
    index: usize,
    name: &str,
) -> Result<String, ParseValueError> {
    let value = args
        .get(index)
        .ok_or_else(|| ParseValueError::new(format!("{name} id is required")))?;
    validate_id_arg(name, value)?;
    Ok(value.clone())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_flag_like_ids_after_trimming() {
        let err = validate_id_arg("task", " --bogus").unwrap_err();
        assert_eq!(err.to_string(), "task id must not start with '-'");
    }

    #[test]
    fn parses_strict_integers_and_rejects_trailing_junk() {
        assert_eq!(parse_i32_strict(" 10 ").unwrap(), 10);
        assert_eq!(parse_i64_strict(" 9223372036854775807 ").unwrap(), i64::MAX);
        assert!(parse_i32_strict("10x").is_err());
        assert!(parse_i64_strict("10.5").is_err());
    }

    #[test]
    fn parses_strict_floats_and_rejects_non_finite_values() {
        assert_eq!(parse_f64_strict(" 10.5 ").unwrap(), 10.5);
        assert!(parse_f64_strict("10.5x").is_err());
        assert!(parse_f64_strict("NaN").is_err());
        assert!(parse_f64_strict("+Inf").is_err());
        assert!(parse_f64_strict("-Infinity").is_err());
    }
}
