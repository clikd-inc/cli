use crate::utils::theme::*;
use crate::{cli::StatusArgs, config::Config};
use anyhow::Result;

pub async fn run(_args: StatusArgs, _config: Config) -> Result<()> {
    println!("{}", info_message("Status command - not yet implemented"));
    Ok(())
}
