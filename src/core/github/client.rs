// Copyright 2020-2022 Peter Williams <peter@newton.cx> and collaborators
// Licensed under the MIT License.

//! Release automation utilities related to the GitHub service.

use anyhow::{anyhow, Context};
use clap::Parser;
use git_url_parse::types::provider::GenericProvider;
use json::{object, JsonValue};
use reqwest::StatusCode;
use std::thread;
use std::time::Duration;
use tracing::{info, warn};

use crate::core::release::{
    env::require_var,
    errors::Result,
    session::{AppBuilder, AppSession},
};

const MAX_RETRIES: u32 = 3;
const INITIAL_BACKOFF_MS: u64 = 1000;
const MAX_BACKOFF_MS: u64 = 30000;

fn is_retryable_status(status: StatusCode) -> bool {
    matches!(
        status,
        StatusCode::TOO_MANY_REQUESTS
            | StatusCode::INTERNAL_SERVER_ERROR
            | StatusCode::BAD_GATEWAY
            | StatusCode::SERVICE_UNAVAILABLE
            | StatusCode::GATEWAY_TIMEOUT
    )
}

fn calculate_backoff(attempt: u32, base_ms: u64) -> Duration {
    let backoff_ms = base_ms * 2u64.pow(attempt);
    Duration::from_millis(backoff_ms.min(MAX_BACKOFF_MS))
}

fn extract_retry_after(response: &reqwest::blocking::Response) -> Option<Duration> {
    response
        .headers()
        .get("retry-after")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok())
        .map(Duration::from_secs)
}

fn is_retryable_error(error: &reqwest::Error) -> bool {
    error.is_timeout() || error.is_connect() || error.is_request()
}

pub struct GitHubInformation {
    slug: String,
    token: String,
}

impl GitHubInformation {
    pub fn new(sess: &AppSession) -> Result<Self> {
        Self::new_with_scopes(sess, &["repo"])
    }

    pub fn new_with_scopes(sess: &AppSession, required_scopes: &[&str]) -> Result<Self> {
        let is_ci = sess
            .execution_environment()
            .map(|env| matches!(env, crate::core::release::session::ExecutionEnvironment::Ci))
            .unwrap_or(false);

        let token = crate::core::auth::token::load_token()
            .ok()
            .or_else(|| require_var("GITHUB_TOKEN").ok())
            .ok_or_else(|| {
                if is_ci {
                    anyhow!(
                        "GitHub authentication required in CI. Set the GITHUB_TOKEN environment variable \
                        (typically via secrets.GITHUB_TOKEN in GitHub Actions)."
                    )
                } else {
                    anyhow!(
                        "GitHub authentication required. Run 'clikd login' to authenticate."
                    )
                }
            })?;

        if !required_scopes.is_empty() {
            crate::core::auth::github::validate_token_scopes_blocking(&token, required_scopes)
                .context("GitHub token scope validation failed")?;
        }

        let upstream_url = sess.repo.upstream_url()?;
        info!("upstream url: {}", upstream_url);

        let upstream_url = git_url_parse::GitUrl::parse(&upstream_url)
            .map_err(|e| anyhow!("cannot parse upstream Git URL `{}`: {}", upstream_url, e))?;

        let provider: GenericProvider = upstream_url
            .provider_info()
            .map_err(|e| anyhow!("cannot extract provider info from Git URL: {}", e))?;
        let slug = format!("{}/{}", provider.owner(), provider.repo());

        Ok(GitHubInformation { slug, token })
    }

    pub fn make_blocking_client(&self) -> Result<reqwest::blocking::Client> {
        use reqwest::header;
        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::AUTHORIZATION,
            header::HeaderValue::from_str(&format!("token {}", self.token))?,
        );
        headers.insert(header::USER_AGENT, header::HeaderValue::from_str("clikd")?);

        Ok(reqwest::blocking::Client::builder()
            .default_headers(headers)
            .build()?)
    }

    fn api_url(&self, rest: &str) -> String {
        format!("https://api.github.com/repos/{}/{}", self.slug, rest)
    }

    fn delete_release(&self, tag_name: &str, client: &reqwest::blocking::Client) -> Result<()> {
        let query_url = self.api_url(&format!("releases/tags/{tag_name}"));

        let resp = self.send_with_retry(|| client.get(&query_url))?;
        if !resp.status().is_success() {
            return Err(anyhow!(
                "no GitHub release for tag `{}`: {}",
                tag_name,
                resp.text()
                    .unwrap_or_else(|_| "[non-textual server response]".to_owned())
            ));
        }

        let metadata = json::parse(&resp.text()?)?;
        let id = metadata["id"].to_string();

        let delete_url = self.api_url(&format!("releases/{id}"));
        let resp = self.send_with_retry(|| client.delete(&delete_url))?;
        if !resp.status().is_success() {
            return Err(anyhow!(
                "could not delete GitHub release for tag `{}`: {}",
                tag_name,
                resp.text()
                    .unwrap_or_else(|_| "[non-textual server response]".to_owned())
            ));
        }

        Ok(())
    }

    fn create_custom_release(
        &self,
        tag_name: String,
        release_name: String,
        body: String,
        is_draft: bool,
        is_prerelease: bool,
        client: &reqwest::blocking::Client,
    ) -> Result<JsonValue> {
        let saved_tag_name = tag_name.clone();
        let release_info = object! {
            "tag_name" => tag_name,
            "name" => release_name,
            "body" => body,
            "draft" => is_draft,
            "prerelease" => is_prerelease,
        };

        let create_url = self.api_url("releases");
        let request_body = json::stringify(release_info);
        let resp = self.send_with_retry(|| client.post(&create_url).body(request_body.clone()))?;
        let status = resp.status();
        let parsed = json::parse(&resp.text()?)?;

        if status.is_success() {
            info!("created GitHub release for {}", saved_tag_name);
            Ok(parsed)
        } else {
            Err(anyhow!(
                "failed to create GitHub release for {}: {}",
                saved_tag_name,
                parsed
            ))
        }
    }

    pub fn create_pull_request(
        &self,
        head: &str,
        base: &str,
        title: &str,
        body: &str,
        client: &reqwest::blocking::Client,
    ) -> Result<String> {
        let pr_info = object! {
            "title" => title,
            "head" => head,
            "base" => base,
            "body" => body,
        };

        let create_url = self.api_url("pulls");
        let request_body = json::stringify(pr_info);
        let resp = self.send_with_retry(|| client.post(&create_url).body(request_body.clone()))?;

        let status = resp.status();
        let parsed = json::parse(&resp.text()?)?;

        if status.is_success() {
            let html_url = parsed["html_url"]
                .as_str()
                .ok_or_else(|| anyhow!("PR response missing html_url"))?
                .to_string();
            info!("created pull request: {}", html_url);
            Ok(html_url)
        } else {
            Err(anyhow!("failed to create pull request: {}", parsed))
        }
    }

    fn send_with_retry<F>(&self, build_request: F) -> Result<reqwest::blocking::Response>
    where
        F: Fn() -> reqwest::blocking::RequestBuilder,
    {
        let mut last_error = None;

        for attempt in 0..=MAX_RETRIES {
            let request = build_request();

            match request.send() {
                Ok(response) => {
                    let status = response.status();

                    if status.is_success() || !is_retryable_status(status) {
                        return Ok(response);
                    }

                    if attempt < MAX_RETRIES {
                        let backoff = extract_retry_after(&response)
                            .unwrap_or_else(|| calculate_backoff(attempt, INITIAL_BACKOFF_MS));

                        warn!(
                            "GitHub API returned {} (attempt {}/{}), retrying in {:?}",
                            status,
                            attempt + 1,
                            MAX_RETRIES + 1,
                            backoff
                        );

                        thread::sleep(backoff);
                    } else {
                        return Ok(response);
                    }
                }
                Err(e) => {
                    if attempt < MAX_RETRIES && is_retryable_error(&e) {
                        let backoff = calculate_backoff(attempt, INITIAL_BACKOFF_MS);

                        warn!(
                            "GitHub API request failed: {} (attempt {}/{}), retrying in {:?}",
                            e,
                            attempt + 1,
                            MAX_RETRIES + 1,
                            backoff
                        );

                        thread::sleep(backoff);
                        last_error = Some(e);
                    } else {
                        return Err(e.into());
                    }
                }
            }
        }

        Err(last_error.map_or_else(
            || {
                anyhow!(
                    "GitHub API request failed after {} retries",
                    MAX_RETRIES + 1
                )
            },
            |e| {
                anyhow!(
                    "GitHub API request failed after {} retries: {}",
                    MAX_RETRIES + 1,
                    e
                )
            },
        ))
    }
}

/// The `github` subcommands.
#[derive(Debug, Eq, PartialEq, Parser)]
pub enum GithubCommands {
    #[structopt(name = "create-custom-release")]
    /// Create a single, customized GitHub release
    CreateCustomRelease(CreateCustomReleaseCommand),

    #[command(name = "_credential-helper", hide = true)]
    /// (hidden) github credential helper
    CredentialHelper(CredentialHelperCommand),

    #[structopt(name = "delete-release")]
    /// Delete an existing GitHub release
    DeleteRelease(DeleteReleaseCommand),

    #[structopt(name = "install-credential-helper")]
    /// Install Clikd as a Git "credential helper", using $GITHUB_TOKEN to log in
    InstallCredentialHelper(InstallCredentialHelperCommand),
}

#[derive(Debug, Eq, PartialEq, Parser)]
pub struct GithubCommand {
    #[command(subcommand)]
    command: GithubCommands,
}

impl GithubCommand {
    pub fn execute(self) -> Result<i32> {
        match self.command {
            GithubCommands::CreateCustomRelease(o) => o.execute(),
            GithubCommands::CredentialHelper(o) => o.execute(),
            GithubCommands::DeleteRelease(o) => o.execute(),
            GithubCommands::InstallCredentialHelper(o) => o.execute(),
        }
    }
}

/// Create a single custom GitHub release.
#[derive(Debug, Eq, PartialEq, Parser)]
pub struct CreateCustomReleaseCommand {
    #[structopt(long = "name", help = "The user-facing name for the release")]
    release_name: String,

    #[structopt(
        long = "desc",
        help = "The release description text (Markdown-formatted)",
        default_value = "Release automatically created by Clikd."
    )]
    body: String,

    #[structopt(long = "draft", help = "Whether to mark this release as a draft")]
    is_draft: bool,

    #[structopt(
        long = "prerelease",
        help = "Whether to mark this release as a pre-release"
    )]
    is_prerelease: bool,

    #[structopt(help = "Name of the Git(Hub) tag to use as the release basis")]
    tag_name: String,
}

impl CreateCustomReleaseCommand {
    pub fn execute(self) -> Result<i32> {
        let sess = AppBuilder::new()?.populate_graph(false).initialize()?;
        let info = GitHubInformation::new(&sess)?;
        let client = info.make_blocking_client()?;
        info.create_custom_release(
            self.tag_name,
            self.release_name,
            self.body,
            self.is_draft,
            self.is_prerelease,
            &client,
        )?;
        Ok(0)
    }
}

/// hidden Git credential helper command
#[derive(Debug, Eq, PartialEq, Parser)]
pub struct CredentialHelperCommand {
    #[structopt(help = "The operation")]
    operation: String,
}

impl CredentialHelperCommand {
    pub fn execute(self) -> Result<i32> {
        if self.operation != "get" {
            info!("ignoring Git credential operation `{}`", self.operation);
        } else {
            let token = require_var("GITHUB_TOKEN")?;
            println!("username=token");
            println!("password={token}");
        }

        Ok(0)
    }
}

/// Delete a release from GitHub.
#[derive(Debug, Eq, PartialEq, Parser)]
pub struct DeleteReleaseCommand {
    #[structopt(help = "Name of the release's tag on GitHub")]
    tag_name: String,
}

impl DeleteReleaseCommand {
    pub fn execute(self) -> Result<i32> {
        let sess = AppSession::initialize_default()?;
        let info = GitHubInformation::new(&sess)?;
        let client = info.make_blocking_client()?;
        info.delete_release(&self.tag_name, &client)?;
        info!(
            "deleted GitHub release associated with tag `{}`",
            self.tag_name
        );
        Ok(0)
    }
}

/// Install as a Git credential helper
#[derive(Debug, Eq, PartialEq, Parser)]
pub struct InstallCredentialHelperCommand {}

impl InstallCredentialHelperCommand {
    pub fn execute(self) -> Result<i32> {
        let this_exe = std::env::current_exe()?;
        let this_exe = this_exe.to_str().ok_or_else(|| {
            anyhow!(
                "cannot install clikd as a Git \
                 credential helper because its executable path is not Unicode"
            )
        })?;
        let mut cfg = git2::Config::open_default().context("cannot open Git configuration")?;
        cfg.set_str(
            "credential.helper",
            &format!("{this_exe} github _credential-helper"),
        )
        .context("cannot update Git configuration setting `credential.helper`")?;
        Ok(0)
    }
}
