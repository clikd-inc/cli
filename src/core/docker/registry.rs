use crate::core::auth::{token, github};
use crate::error::Result;
use bollard::auth::DockerCredentials;
use tracing::debug;

pub async fn get_ghcr_credentials() -> Result<DockerCredentials> {
    debug!("Loading GHCR credentials");

    let token_str = token::load_token()?;
    let username = github::get_username(&token_str).await?;

    Ok(DockerCredentials {
        username: Some(username),
        password: Some(token_str),
        serveraddress: Some("ghcr.io".to_string()),
        ..Default::default()
    })
}

pub fn is_ghcr_image(image: &str) -> bool {
    image.starts_with("ghcr.io/") || image.starts_with("ghcr.io:")
}
