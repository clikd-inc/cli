use crate::error::{CliError, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tokio::time::sleep;
use tracing::debug;

const DEVICE_CODE_URL: &str = "https://github.com/login/device/code";
const ACCESS_TOKEN_URL: &str = "https://github.com/login/oauth/access_token";

#[derive(Debug, Serialize)]
struct DeviceCodeRequest {
    client_id: String,
    scope: String,
}

#[derive(Debug, Deserialize)]
pub struct DeviceCodeResponse {
    pub device_code: String,
    pub user_code: String,
    pub verification_uri: String,
    pub expires_in: u64,
    pub interval: u64,
}

#[derive(Debug, Serialize)]
struct AccessTokenRequest {
    client_id: String,
    device_code: String,
    grant_type: String,
}

#[derive(Debug, Deserialize)]
#[serde(untagged)]
enum AccessTokenResponse {
    Success {
        access_token: String,
        token_type: String,
        scope: String,
    },
    Error {
        error: String,
        error_description: String,
    },
}

pub async fn request_device_code(client_id: &str) -> Result<DeviceCodeResponse> {
    let client = Client::new();

    let response = client
        .post(DEVICE_CODE_URL)
        .header("Accept", "application/json")
        .form(&DeviceCodeRequest {
            client_id: client_id.to_string(),
            scope: "repo read:org user:email read:packages".to_string(),
        })
        .send()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to request device code: {}", e)))?;

    if !response.status().is_success() {
        let status = response.status();
        let body = response.text().await.unwrap_or_default();
        return Err(CliError::GitHubApi(format!(
            "Device code request failed with status {}: {}",
            status, body
        )));
    }

    response
        .json::<DeviceCodeResponse>()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to parse device code response: {}", e)))
}

pub async fn poll_for_token(
    client_id: &str,
    device_code: &str,
    interval: u64,
    expires_in: u64,
) -> Result<String> {
    let client = Client::new();
    let mut current_interval = interval;
    let start_time = std::time::Instant::now();
    let timeout = Duration::from_secs(expires_in);

    loop {
        if start_time.elapsed() > timeout {
            return Err(CliError::GitHubApi(
                "Device code expired. Please try again.".to_string(),
            ));
        }

        sleep(Duration::from_secs(current_interval)).await;

        let response = client
            .post(ACCESS_TOKEN_URL)
            .header("Accept", "application/json")
            .form(&AccessTokenRequest {
                client_id: client_id.to_string(),
                device_code: device_code.to_string(),
                grant_type: "urn:ietf:params:oauth:grant-type:device_code".to_string(),
            })
            .send()
            .await
            .map_err(|e| CliError::GitHubApi(format!("Failed to poll for token: {}", e)))?;

        let result = response
            .json::<AccessTokenResponse>()
            .await
            .map_err(|e| CliError::GitHubApi(format!("Failed to parse token response: {}", e)))?;

        match result {
            AccessTokenResponse::Success {
                access_token,
                token_type,
                scope,
            } => {
                if token_type != "bearer" {
                    return Err(CliError::GitHubApi(format!(
                        "Unexpected token type: {}. Expected 'bearer'",
                        token_type
                    )));
                }

                debug!("Token received with scope: {}", scope);
                return Ok(access_token);
            }
            AccessTokenResponse::Error {
                error,
                error_description,
            } => match error.as_str() {
                "authorization_pending" => {
                    continue;
                }
                "slow_down" => {
                    current_interval += 5;
                    continue;
                }
                "expired_token" => {
                    return Err(CliError::GitHubApi(
                        "Device code expired. Please try again.".to_string(),
                    ));
                }
                "access_denied" => {
                    return Err(CliError::GitHubApi("Authorization was denied.".to_string()));
                }
                _ => {
                    return Err(CliError::GitHubApi(format!(
                        "GitHub API error: {} - {}",
                        error, error_description
                    )));
                }
            },
        }
    }
}

pub async fn get_username(token: &str) -> Result<String> {
    let client = Client::new();

    let response = client
        .get("https://api.github.com/user")
        .header("Authorization", format!("Bearer {}", token))
        .header("Accept", "application/vnd.github+json")
        .header("User-Agent", "clikd")
        .send()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to get user info: {}", e)))?;

    if !response.status().is_success() {
        return Err(CliError::GitHubApi(format!(
            "Failed to get user info: HTTP {}",
            response.status()
        )));
    }

    #[derive(Deserialize)]
    struct UserResponse {
        login: String,
    }

    let user: UserResponse = response
        .json()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to parse user response: {}", e)))?;

    Ok(user.login)
}

pub async fn validate_token_scopes(token: &str, required_scopes: &[&str]) -> Result<()> {
    let client = Client::new();

    let response = client
        .get("https://api.github.com/user")
        .header("Authorization", format!("Bearer {}", token))
        .header("Accept", "application/vnd.github+json")
        .header("User-Agent", "clikd")
        .send()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to validate token: {}", e)))?;

    if !response.status().is_success() {
        return Err(CliError::GitHubApi(format!(
            "Token validation failed: HTTP {}",
            response.status()
        )));
    }

    let scopes = response
        .headers()
        .get("x-oauth-scopes")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");

    check_required_scopes(scopes, required_scopes)?;

    debug!("Token scopes validated: {}", scopes);
    Ok(())
}

pub fn validate_token_scopes_blocking(token: &str, required_scopes: &[&str]) -> Result<()> {
    let client = reqwest::blocking::Client::new();

    let response = client
        .get("https://api.github.com/user")
        .header("Authorization", format!("Bearer {}", token))
        .header("Accept", "application/vnd.github+json")
        .header("User-Agent", "clikd")
        .send()
        .map_err(|e| CliError::GitHubApi(format!("Failed to validate token: {}", e)))?;

    if !response.status().is_success() {
        return Err(CliError::GitHubApi(format!(
            "Token validation failed: HTTP {}",
            response.status()
        )));
    }

    let scopes = response
        .headers()
        .get("x-oauth-scopes")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");

    check_required_scopes(scopes, required_scopes)?;

    debug!("Token scopes validated: {}", scopes);
    Ok(())
}

fn check_required_scopes(scopes: &str, required_scopes: &[&str]) -> Result<()> {
    let token_scopes: Vec<&str> = scopes.split(", ").map(|s| s.trim()).collect();

    let missing: Vec<&str> = required_scopes
        .iter()
        .filter(|&required| !token_scopes.iter().any(|s| s == required))
        .copied()
        .collect();

    if !missing.is_empty() {
        return Err(CliError::GitHubApi(format!(
            "Token missing required scopes: {}. Please re-authenticate with 'clikd auth login'",
            missing.join(", ")
        )));
    }

    Ok(())
}
