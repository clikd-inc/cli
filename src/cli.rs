use clap::{Parser, Subcommand, Args, ValueEnum};

#[derive(Parser)]
#[command(name = "clikd", version, about = "Development CLI for Clikd platform")]
#[command(propagate_version = true)]
pub struct Cli {
    #[arg(short, long, action = clap::ArgAction::Count, global = true)]
    pub verbose: u8,

    #[arg(long, global = true)]
    pub no_color: bool,

    #[arg(short, long, global = true, env = "CLIKD_ENV")]
    pub env: Option<String>,

    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand)]
pub enum Commands {
    Login {
        #[arg(long)]
        no_browser: bool,
    },

    Logout,

    #[command(subcommand)]
    Auth(AuthCommands),

    Init(InitArgs),

    Start(StartArgs),

    Stop(StopArgs),

    Status(StatusArgs),

    Logs(LogsArgs),

    #[command(subcommand)]
    Db(DbCommands),

    Update(UpdateArgs),

    Completions {
        #[arg(value_enum)]
        shell: clap_complete::Shell,
    },
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
    Status,
}

#[derive(Args)]
pub struct StartArgs {
    #[arg(short = 'x', long, value_delimiter = ',')]
    pub exclude: Option<Vec<String>>,

    #[arg(long)]
    pub ignore_health_check: bool,
}

#[derive(Args)]
pub struct StopArgs {
    #[arg(short, long)]
    pub force: bool,

    #[arg(long)]
    pub purge: bool,
}

#[derive(Args)]
pub struct StatusArgs {
    #[arg(short, long, value_enum, default_value = "table")]
    pub format: OutputFormat,
}

#[derive(Clone, ValueEnum)]
pub enum OutputFormat {
    Table,
    Json,
    Env,
}

#[derive(Args)]
pub struct LogsArgs {
    #[arg(short, long)]
    pub service: Option<String>,

    #[arg(short, long)]
    pub follow: bool,

    #[arg(short = 'n', long, default_value = "100")]
    pub tail: usize,
}

#[derive(Subcommand)]
pub enum DbCommands {
    Migrate,

    Reset {
        #[arg(short, long)]
        force: bool,
    },

    Seed,
}

#[derive(Args)]
pub struct UpdateArgs {
    #[arg(long, help = "Skip confirmation prompts")]
    pub yes: bool,
}
