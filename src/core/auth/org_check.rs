use crate::error::{CliError, Result};
use reqwest::Client;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct OrgResponse {
    login: String,
}

pub async fn verify_membership(token: &str, required_org: &str) -> Result<()> {
    let client = Client::new();

    let response = client
        .get("https://api.github.com/user/orgs")
        .header("Authorization", format!("Bearer {}", token))
        .header("Accept", "application/vnd.github+json")
        .header("User-Agent", "clikd-cli")
        .send()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to fetch user organizations: {}", e)))?;

    if !response.status().is_success() {
        return Err(CliError::GitHubApi(format!(
            "Failed to fetch user organizations: HTTP {}",
            response.status()
        )));
    }

    let orgs: Vec<OrgResponse> = response
        .json()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to parse organizations response: {}", e)))?;

    let is_member = orgs.iter().any(|org| org.login == required_org);

    if is_member {
        Ok(())
    } else {
        Err(CliError::UnauthorizedOrg(required_org.to_string()))
    }
}
