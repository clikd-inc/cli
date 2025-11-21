use clap::{Args, Parser, Subcommand, ValueEnum};

#[derive(Parser)]
#[command(
    name = "clikd",
    about = "Development CLI for Clikd platform",
    long_about = "A powerful CLI tool for managing your Clikd development environment.\nProvides commands for authentication, service orchestration, and container monitoring.",
    version,
    after_help = "For detailed command help, run: clikd <COMMAND> --help"
)]
#[command(disable_version_flag = true)]
pub struct Cli {
    #[arg(
        short,
        long,
        action = clap::ArgAction::Count,
        global = true,
        help = "Increase logging verbosity (-v, -vv, -vvv)"
    )]
    pub verbose: u8,

    #[arg(long, global = true, help = "Disable colored output")]
    pub no_color: bool,

    #[arg(
        short,
        long,
        global = true,
        env = "CLIKD_ENV",
        help = "Environment configuration to use (development, staging, production)"
    )]
    pub env: Option<String>,

    #[arg(short = 'V', long, help = "Print version information")]
    pub version: bool,

    #[command(subcommand)]
    pub command: Option<Commands>,
}

#[derive(Subcommand)]
pub enum Commands {
    #[command(about = "Authenticate with Clikd platform")]
    Login {
        #[arg(long, help = "Skip opening browser and show URL only")]
        no_browser: bool,
    },

    #[command(about = "Sign out from Clikd platform")]
    Logout,

    #[command(subcommand, about = "Authentication management commands")]
    Auth(AuthCommands),

    #[command(about = "Initialize a new Clikd project")]
    Init(InitArgs),

    #[command(
        about = "Start all services",
        long_about = "Starts all configured services in Docker containers.\nCreates network, pulls images, and ensures health checks pass."
    )]
    Start(StartArgs),

    #[command(
        about = "Stop all running services",
        long_about = "Stops all running containers.\nUse --purge to also remove volumes and clean up completely."
    )]
    Stop(StopArgs),

    #[command(
        about = "Interactive container monitoring TUI",
        long_about = "Launch an interactive terminal UI for real-time container monitoring.\n\nFeatures:\n  • Live container metrics (CPU, memory, network)\n  • Interactive log viewer with search and export\n  • Container controls (start, stop, restart, pause, delete)\n  • Sortable columns and mouse support\n  • Press 'h' in TUI for keyboard shortcuts"
    )]
    Status(StatusArgs),

    #[command(about = "Update CLI to the latest version")]
    Update(UpdateArgs),

    #[command(about = "Generate shell completions")]
    Completions {
        #[arg(value_enum, help = "Shell type to generate completions for")]
        shell: clap_complete::Shell,
    },

    #[command(subcommand, about = "Release management commands")]
    Release(ReleaseCommands),
}

#[derive(Args)]
pub struct InitArgs {
    #[arg(long, help = "Generate VSCode settings")]
    pub vscode: bool,

    #[arg(long, help = "Generate IntelliJ/Android Studio settings")]
    pub intellij: bool,

    #[arg(long, help = "Custom working directory")]
    pub workdir: Option<std::path::PathBuf>,
}

#[derive(Subcommand)]
pub enum AuthCommands {
    #[command(about = "Show current authentication status")]
    Status,
}

#[derive(Subcommand)]
pub enum ReleaseCommands {
    #[command(about = "Initialize Clikd release management")]
    Init {
        #[arg(short, long, help = "Force operation even in unexpected conditions")]
        force: bool,

        #[arg(short, long, help = "The name of the Git upstream remote")]
        upstream: Option<String>,
    },

    #[command(about = "Show release status and changelog")]
    Status,

    #[command(about = "Prepare a release (bump versions)")]
    Prepare {
        #[arg(help = "Version bump type: major, minor, patch, or manual")]
        bump: Option<String>,
    },
}

#[derive(Args)]
pub struct StartArgs {
    #[arg(
        short = 'x',
        long,
        value_delimiter = ',',
        help = "Exclude specific services from starting (comma-separated)"
    )]
    pub exclude: Option<Vec<String>>,

    #[arg(long, help = "Skip health checks and start immediately")]
    pub ignore_health_check: bool,
}

#[derive(Args)]
pub struct StopArgs {
    #[arg(short, long, help = "Force stop containers immediately")]
    pub force: bool,

    #[arg(long, help = "Remove volumes and clean up all resources")]
    pub purge: bool,
}

#[derive(Args)]
pub struct StatusArgs {
    #[arg(
        short,
        long,
        value_enum,
        default_value = "table",
        help = "Output format (currently unused, TUI mode is always interactive)"
    )]
    pub format: OutputFormat,
}

#[derive(Clone, ValueEnum)]
pub enum OutputFormat {
    #[value(help = "Interactive table (TUI mode)")]
    Table,
    #[value(help = "JSON output")]
    Json,
    #[value(help = "Environment variables")]
    Env,
}

#[derive(Args)]
pub struct UpdateArgs {
    #[arg(long, help = "Skip confirmation prompts and update immediately")]
    pub yes: bool,
}
