use super::types::Config;
use config::{Config as ConfigBuilder, Environment, File, FileFormat};
use crate::error::{CliError, Result};

const DEFAULT_CONFIG: &str = include_str!("../../../config/default.toml");

pub fn load(env: Option<&str>) -> Result<Config> {
    let env_name = env.unwrap_or("development");

    let mut builder = ConfigBuilder::builder()
        .add_source(File::from_str(DEFAULT_CONFIG, FileFormat::Toml));

    if let Some(config_dir) = dirs::config_dir() {
        let clikd_config = config_dir.join("clikd");
        builder = builder
            .add_source(File::from(clikd_config.join("default.toml")).required(false))
            .add_source(File::from(clikd_config.join(format!("{}.toml", env_name))).required(false))
            .add_source(File::from(clikd_config.join("local.toml")).required(false));
    }

    if std::path::Path::new("config/default.toml").exists() {
        builder = builder
            .add_source(File::with_name("config/default").required(false))
            .add_source(File::with_name(&format!("config/{}", env_name)).required(false))
            .add_source(File::with_name("config/local").required(false));
    }

    let config = builder
        .add_source(
            Environment::with_prefix("CLIKD")
                .separator("_")
                .try_parsing(true),
        )
        .build()
        .map_err(CliError::Config)?;

    config.try_deserialize().map_err(CliError::Config)
}
