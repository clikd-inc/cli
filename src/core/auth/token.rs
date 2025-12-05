use crate::error::{CliError, Result};
use keyring::Entry;

const SERVICE_NAME: &str = "clikd";
const TOKEN_KEY: &str = "github-token";
const MANIFEST_SECRET_KEY: &str = "manifest-secret";

pub fn save_token(token: &str) -> Result<()> {
    let entry = Entry::new(SERVICE_NAME, TOKEN_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    entry
        .set_password(token)
        .map_err(|e| CliError::TokenStorage(format!("Failed to save token: {}", e)))?;

    Ok(())
}

pub fn load_token() -> Result<String> {
    let entry = Entry::new(SERVICE_NAME, TOKEN_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    entry.get_password().map_err(|e| match e {
        keyring::Error::NoEntry => CliError::AuthenticationRequired,
        _ => CliError::TokenStorage(format!("Failed to load token: {}", e)),
    })
}

pub fn delete_token() -> Result<()> {
    let entry = Entry::new(SERVICE_NAME, TOKEN_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    match entry.delete_credential() {
        Ok(_) => Ok(()),
        Err(keyring::Error::NoEntry) => Ok(()),
        Err(e) => Err(CliError::TokenStorage(format!(
            "Failed to delete token: {}",
            e
        ))),
    }
}

pub fn save_manifest_secret(secret: &str) -> Result<()> {
    let entry = Entry::new(SERVICE_NAME, MANIFEST_SECRET_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    entry
        .set_password(secret)
        .map_err(|e| CliError::TokenStorage(format!("Failed to save manifest secret: {}", e)))?;

    Ok(())
}

pub fn load_manifest_secret() -> Result<String> {
    let entry = Entry::new(SERVICE_NAME, MANIFEST_SECRET_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    entry.get_password().map_err(|e| match e {
        keyring::Error::NoEntry => CliError::TokenStorage(
            "Manifest secret not configured. Run 'clikd auth secret' to set it.".to_string(),
        ),
        _ => CliError::TokenStorage(format!("Failed to load manifest secret: {}", e)),
    })
}

pub fn delete_manifest_secret() -> Result<()> {
    let entry = Entry::new(SERVICE_NAME, MANIFEST_SECRET_KEY)
        .map_err(|e| CliError::TokenStorage(format!("Failed to create keyring entry: {}", e)))?;

    match entry.delete_credential() {
        Ok(_) => Ok(()),
        Err(keyring::Error::NoEntry) => Ok(()),
        Err(e) => Err(CliError::TokenStorage(format!(
            "Failed to delete manifest secret: {}",
            e
        ))),
    }
}
