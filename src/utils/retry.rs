use std::time::Duration;

pub async fn retry_with_backoff<F, Fut, T, E>(
    _f: F,
    _max_retries: u32,
    _initial_delay: Duration,
) -> Result<T, E>
where
    F: Fn() -> Fut,
    Fut: std::future::Future<Output = Result<T, E>>,
{
    unimplemented!()
}
