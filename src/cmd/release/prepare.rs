use anyhow::{bail, Result};
use log::info;

pub fn run(bump: Option<String>) -> Result<i32> {
    info!("preparing release with clikd version {}", env!("CARGO_PKG_VERSION"));

    if let Some(bump_type) = bump {
        info!("bump type: {}", bump_type);
    }

    bail!("release prepare not yet fully implemented - use `clikd release` commands");
}
