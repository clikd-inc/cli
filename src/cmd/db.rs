use crate::config::Config;
use crate::utils::theme::*;
use anyhow::Result;

pub async fn migrate(_config: Config) -> Result<()> {
    println!("{}", info_message("DB migrate - not yet implemented"));
    Ok(())
}

pub async fn reset(_force: bool, _config: Config) -> Result<()> {
    println!("{}", info_message("DB reset - not yet implemented"));
    Ok(())
}

pub async fn seed(_config: Config) -> Result<()> {
    println!("{}", info_message("DB seed - not yet implemented"));
    Ok(())
}
