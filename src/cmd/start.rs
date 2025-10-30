use anyhow::Result;
use crate::{cli::StartArgs, config::Config};
use crate::core::start::runner;

pub async fn run(args: StartArgs, config: Config) -> Result<()> {
    let exclude = args.exclude.unwrap_or_default();
    runner::run(&config, exclude, args.ignore_health_check).await?;
    Ok(())
}
