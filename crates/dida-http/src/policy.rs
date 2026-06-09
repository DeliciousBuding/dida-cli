use std::time::Duration;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum RetryDecision {
    RetryAfter(Duration),
    DoNotRetry,
}

#[derive(Clone, Debug)]
pub struct RetryPolicy {
    pub max_retries: usize,
    pub base_delay: Duration,
    pub max_delay: Duration,
    pub retry_429: bool,
    pub retry_server_errors: bool,
}

impl Default for RetryPolicy {
    fn default() -> Self {
        Self {
            max_retries: 2,
            base_delay: Duration::from_millis(200),
            max_delay: Duration::from_secs(5),
            retry_429: true,
            retry_server_errors: true,
        }
    }
}

impl RetryPolicy {
    pub fn none() -> Self {
        Self {
            max_retries: 0,
            ..Self::default()
        }
    }

    pub fn classify_status(&self, status: u16, attempt: usize) -> RetryDecision {
        let retryable = (self.retry_429 && status == 429)
            || (self.retry_server_errors && (500..=599).contains(&status));
        if !retryable || attempt >= self.max_retries {
            return RetryDecision::DoNotRetry;
        }
        RetryDecision::RetryAfter(self.delay_for_attempt(attempt))
    }

    pub fn classify_transport_error(&self, attempt: usize) -> RetryDecision {
        if attempt >= self.max_retries {
            return RetryDecision::DoNotRetry;
        }
        RetryDecision::RetryAfter(self.delay_for_attempt(attempt))
    }

    pub fn delay_for_attempt(&self, attempt: usize) -> Duration {
        let multiplier = 1u32.checked_shl(attempt as u32).unwrap_or(u32::MAX);
        self.base_delay
            .saturating_mul(multiplier)
            .min(self.max_delay)
    }
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct TimeoutPolicy {
    pub request_timeout: Duration,
}

impl TimeoutPolicy {
    pub fn new(request_timeout: Duration) -> Self {
        Self { request_timeout }
    }
}

impl Default for TimeoutPolicy {
    fn default() -> Self {
        Self::new(Duration::from_secs(30))
    }
}
