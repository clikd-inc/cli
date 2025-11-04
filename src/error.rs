use thiserror::Error;

#[derive(Error, Debug)]
pub enum CliError {
    #[error("Docker error: {0}")]
    Docker(#[from] bollard::errors::Error),

    #[error("Git error: {0}")]
    Git(#[from] git2::Error),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Configuration error: {0}")]
    Config(#[from] config::ConfigError),

    #[error("Service '{0}' is not running")]
    ServiceNotRunning(String),

    #[error("Service '{0}' failed health check")]
    HealthCheckFailed(String),

    #[error("Service '{0}' not found")]
    ServiceNotFound(String),

    #[error("Authentication required. Run 'clikd login'")]
    AuthenticationRequired,

    #[error("Not a member of organization '{0}'")]
    UnauthorizedOrg(String),

    #[error("GitHub API error: {0}")]
    GitHubApi(String),

    #[error("Token storage error: {0}")]
    TokenStorage(String),

    #[error("Environment already running for branch '{0}'")]
    AlreadyRunning(String),

    #[error("No environment running")]
    NotRunning,

    #[error("Project already initialized. Run 'clikd init --force' to overwrite.")]
    AlreadyInitialized,

    #[error("Dialog error: {0}")]
    Dialog(#[from] dialoguer::Error),

    #[error("Project not initialized. Run 'clikd init' to get started.")]
    ProjectNotInitialized,
}

pub type Result<T> = std::result::Result<T, CliError>;
