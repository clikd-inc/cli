use anyhow::Result;
use crate::{cli::StartArgs, config::Config};

pub async fn run(_args: StartArgs, _config: Config) -> Result<()> {
    println!("Start command - not yet implemented");
    Ok(())
}
