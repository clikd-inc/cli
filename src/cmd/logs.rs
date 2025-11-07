use crate::utils::theme::*;
use crate::{cli::LogsArgs, config::Config};
use anyhow::Result;

pub async fn run(_args: LogsArgs, _config: Config) -> Result<()> {
    println!("{}", info_message("Logs command - not yet implemented"));
    Ok(())
}
