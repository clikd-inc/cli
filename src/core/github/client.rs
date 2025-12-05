//! Release automation utilities related to the GitHub service.

use anyhow::{anyhow, Context};
use clap::Parser;
use git_url_parse::types::provider::GenericProvider;
use octocrab::Octocrab;
use tracing::info;

use crate::core::release::{
    env::require_var,
    errors::Result,
    session::{AppBuilder, AppSession},
};

pub struct GitHubInformation {
    owner: String,
    repo: String,
    client: Octocrab,
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

        let owner = provider.owner().to_string();
        let repo = provider.repo().to_string();

        let client = Octocrab::builder()
            .personal_token(token)
            .build()
            .context("failed to build GitHub client")?;

        Ok(GitHubInformation {
            owner,
            repo,
            client,
        })
    }

    pub fn create_pull_request(
        &self,
        head: &str,
        base: &str,
        title: &str,
        body: &str,
    ) -> Result<String> {
        let rt = tokio::runtime::Runtime::new().context("failed to create async runtime")?;

        rt.block_on(async {
            let pr = self
                .client
                .pulls(&self.owner, &self.repo)
                .create(title, head, base)
                .body(body)
                .send()
                .await
                .context("failed to create pull request")?;

            let html_url = pr
                .html_url
                .ok_or_else(|| anyhow!("PR response missing html_url"))?
                .to_string();

            info!("created pull request: {}", html_url);
            Ok(html_url)
        })
    }

    fn delete_release(&self, tag_name: &str) -> Result<()> {
        let rt = tokio::runtime::Runtime::new().context("failed to create async runtime")?;

        rt.block_on(async {
            let release = self
                .client
                .repos(&self.owner, &self.repo)
                .releases()
                .get_by_tag(tag_name)
                .await
                .with_context(|| format!("no GitHub release for tag `{}`", tag_name))?;

            self.client
                .repos(&self.owner, &self.repo)
                .releases()
                .delete(release.id.into_inner())
                .await
                .with_context(|| {
                    format!("could not delete GitHub release for tag `{}`", tag_name)
                })?;

            Ok(())
        })
    }

    fn create_custom_release(
        &self,
        tag_name: String,
        release_name: String,
        body: String,
        is_draft: bool,
        is_prerelease: bool,
    ) -> Result<octocrab::models::repos::Release> {
        let rt = tokio::runtime::Runtime::new().context("failed to create async runtime")?;

        rt.block_on(async {
            let release = self
                .client
                .repos(&self.owner, &self.repo)
                .releases()
                .create(&tag_name)
                .name(&release_name)
                .body(&body)
                .draft(is_draft)
                .prerelease(is_prerelease)
                .send()
                .await
                .with_context(|| format!("failed to create GitHub release for {}", tag_name))?;

            info!("created GitHub release for {}", tag_name);
            Ok(release)
        })
    }
}

/// The `github` subcommands.
#[derive(Debug, Eq, PartialEq, Parser)]
pub enum GithubCommands {
    #[command(name = "create-custom-release")]
    /// Create a single, customized GitHub release
    CreateCustomRelease(CreateCustomReleaseCommand),

    #[command(name = "_credential-helper", hide = true)]
    /// (hidden) github credential helper
    CredentialHelper(CredentialHelperCommand),

    #[command(name = "delete-release")]
    /// Delete an existing GitHub release
    DeleteRelease(DeleteReleaseCommand),

    #[command(name = "install-credential-helper")]
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
    #[arg(long = "name", help = "The user-facing name for the release")]
    release_name: String,

    #[arg(
        long = "desc",
        help = "The release description text (Markdown-formatted)",
        default_value = "Release automatically created by Clikd."
    )]
    body: String,

    #[arg(long = "draft", help = "Whether to mark this release as a draft")]
    is_draft: bool,

    #[arg(
        long = "prerelease",
        help = "Whether to mark this release as a pre-release"
    )]
    is_prerelease: bool,

    #[arg(help = "Name of the Git(Hub) tag to use as the release basis")]
    tag_name: String,
}

impl CreateCustomReleaseCommand {
    pub fn execute(self) -> Result<i32> {
        let sess = AppBuilder::new()?.populate_graph(false).initialize()?;
        let info = GitHubInformation::new(&sess)?;
        info.create_custom_release(
            self.tag_name,
            self.release_name,
            self.body,
            self.is_draft,
            self.is_prerelease,
        )?;
        Ok(0)
    }
}

/// hidden Git credential helper command
#[derive(Debug, Eq, PartialEq, Parser)]
pub struct CredentialHelperCommand {
    #[arg(help = "The operation")]
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
    #[arg(help = "Name of the release's tag on GitHub")]
    tag_name: String,
}

impl DeleteReleaseCommand {
    pub fn execute(self) -> Result<i32> {
        let sess = AppSession::initialize_default()?;
        let info = GitHubInformation::new(&sess)?;
        info.delete_release(&self.tag_name)?;
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
