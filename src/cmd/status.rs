use anyhow::Result;
use crate::{cli::StatusArgs, config::Config};
use crate::utils::theme::*;

pub async fn run(_args: StatusArgs, _config: Config) -> Result<()> {
    println!("{}", info_message("Status command - not yet implemented"));
    Ok(())
}
