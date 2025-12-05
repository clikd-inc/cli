use crate::error::{CliError, Result};
use octocrab::Octocrab;

pub async fn verify_membership(token: &str, required_org: &str) -> Result<()> {
    let client = Octocrab::builder()
        .personal_token(token.to_string())
        .build()
        .map_err(|e| CliError::GitHubApi(format!("Failed to build GitHub client: {}", e)))?;

    let orgs = client
        .current()
        .list_org_memberships_for_authenticated_user()
        .send()
        .await
        .map_err(|e| CliError::GitHubApi(format!("Failed to fetch user organizations: {}", e)))?;

    let is_member = orgs
        .items
        .iter()
        .any(|membership| membership.organization.login == required_org);

    if is_member {
        Ok(())
    } else {
        Err(CliError::UnauthorizedOrg(required_org.to_string()))
    }
}
