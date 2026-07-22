use std::thread;
use std::time::Duration;

#[derive(Debug, Clone, Copy)]
pub struct RetryPolicy {
    pub attempts: u32,
    pub initial_delay: Duration,
    pub maximum_delay: Duration,
}

#[derive(Debug, PartialEq, Eq)]
pub enum RetryError<E> {
    InvalidPolicy,
    Exhausted { attempts: u32, source: E },
}

pub fn retry<T, E, F>(policy: RetryPolicy, mut operation: F) -> Result<T, RetryError<E>>
where
    F: FnMut(u32) -> Result<T, E>,
{
    if policy.attempts == 0 || policy.initial_delay > policy.maximum_delay {
        return Err(RetryError::InvalidPolicy);
    }

    let mut delay = policy.initial_delay;
    for attempt in 1..=policy.attempts {
        match operation(attempt) {
            Ok(value) => return Ok(value),
            Err(source) if attempt == policy.attempts => {
                return Err(RetryError::Exhausted {
                    attempts: attempt,
                    source,
                });
            }
            Err(_) => {
                thread::sleep(delay);
                delay = delay.saturating_mul(2).min(policy.maximum_delay);
            }
        }
    }
    unreachable!("positive attempt count always returns")
}
