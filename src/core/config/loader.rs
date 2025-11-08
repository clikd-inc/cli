use super::types::Config;
use crate::error::{CliError, Result};
use config::{Config as ConfigBuilder, Environment, File};
use std::env;

pub fn load(env_name: Option<&str>) -> Result<Config> {
    let env_name = env_name.unwrap_or("development");

    let mut builder = ConfigBuilder::builder();

    let project_root = env::current_dir().unwrap_or_else(|_| std::path::PathBuf::from("."));
    let project_config = project_root.join("clikd/config.toml");

    if project_config.exists() {
        builder = builder.add_source(File::from(project_config).required(true));
    }

    if let Some(config_dir) = dirs::config_dir() {
        let clikd_config = config_dir.join("clikd");
        builder = builder
            .add_source(File::from(clikd_config.join("default.toml")).required(false))
            .add_source(File::from(clikd_config.join(format!("{}.toml", env_name))).required(false))
            .add_source(File::from(clikd_config.join("local.toml")).required(false));
    }

    let mut config: Config = builder
        .add_source(
            Environment::with_prefix("CLIKD")
                .separator("_")
                .try_parsing(true),
        )
        .build()
        .map_err(CliError::Config)?
        .try_deserialize()
        .map_err(CliError::Config)?;

    config.sanitize_project_id();
    config.images = Default::default();

    Ok(config)
}
