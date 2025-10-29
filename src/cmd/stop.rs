use anyhow::Result;
use crate::{cli::StopArgs, config::Config};

pub async fn run(_args: StopArgs, _config: Config) -> Result<()> {
    println!("Stop command - not yet implemented");
    Ok(())
}
