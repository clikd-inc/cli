use anyhow::Result;
use crate::{cli::StatusArgs, config::Config};

pub async fn run(_args: StatusArgs, _config: Config) -> Result<()> {
    println!("Status command - not yet implemented");
    Ok(())
}
