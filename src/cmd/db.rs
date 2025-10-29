use anyhow::Result;
use crate::config::Config;

pub async fn migrate(_config: Config) -> Result<()> {
    println!("DB migrate - not yet implemented");
    Ok(())
}

pub async fn reset(_force: bool, _config: Config) -> Result<()> {
    println!("DB reset - not yet implemented");
    Ok(())
}

pub async fn seed(_config: Config) -> Result<()> {
    println!("DB seed - not yet implemented");
    Ok(())
}
