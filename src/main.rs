use clap::{Parser, Subcommand};
use anyhow::Result;

#[derive(Parser)]
#[command(name = "clikd")]
#[command(about = "Clikd Development CLI - Service orchestration and development tools")]
#[command(version)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Option<Commands>,

    #[arg(long, global = true)]
    pub verbose: bool,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Start development services
    Start {
        #[arg(long)]
        exclude: Option<Vec<String>>,
        #[arg(long, help = "Skip TUI dashboard and run in background")]
        headless: bool,
    },
    /// Stop development services
    Stop {
        #[arg(long)]
        force: bool,
    },
    /// Monitor service status
    Status,
    /// Switch between environments
    Switch,
    /// View and filter service logs
    Logs {
        #[arg(long)]
        service: Option<String>,
    },
    /// Manage databases
    Db {
        #[command(subcommand)]
        command: Option<DbCommands>,
    },
    /// Generate client code
    Gen {
        #[command(subcommand)]
        command: Option<GenCommands>,
    },
    /// Deploy to environments
    Deploy {
        environment: Option<String>,
    },
    /// Launch unified TUI interface
    Tui,
}

#[derive(Subcommand)]
pub enum DbCommands {
    /// Run database migrations
    Migrate {
        #[arg(long)]
        target: Option<String>,
    },
    /// View schema differences
    Diff {
        #[arg(long)]
        branch: Option<String>,
    },
    /// Reset database with confirmation
    Reset,
    /// Seed database with test data
    Seed,
    /// Create database dump
    Dump,
}

#[derive(Subcommand)]
pub enum GenCommands {
    /// Generate Swift client
    Swift {
        #[arg(long)]
        output: Option<String>,
    },
    /// Generate Kotlin client
    Kotlin {
        #[arg(long)]
        output: Option<String>,
    },
    /// Generate TypeScript client
    Typescript {
        #[arg(long)]
        output: Option<String>,
    },
    /// Generate all client libraries
    All,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    setup_tracing(cli.verbose)?;

    match cli.command {
        None => {
            commands::selector::run_interactive().await
        }
        Some(command) => match command {
        Commands::Start { exclude, headless } => {
            if headless {
                commands::start::run_headless(exclude).await
            } else {
                commands::start::run_interactive(exclude).await
            }
        }
        Commands::Stop { force } => {
            commands::stop::run_interactive(force).await
        }
        Commands::Status => {
            commands::status::run_tui().await
        }
        Commands::Switch => {
            commands::switch::run_interactive().await
        }
        Commands::Logs { service } => {
            commands::logs::run_tui(service).await
        }
        Commands::Db { command } => {
            match command {
                Some(DbCommands::Migrate { target }) => commands::db::migrate::run_tui(target).await,
                Some(DbCommands::Diff { branch }) => commands::db::diff::run_tui(branch).await,
                Some(DbCommands::Reset) => commands::db::reset::run_tui().await,
                Some(DbCommands::Seed) => commands::db::seed::run_tui().await,
                Some(DbCommands::Dump) => commands::db::dump::run_tui().await,
                None => commands::db::run_main_tui().await,
            }
        }
        Commands::Gen { command } => {
            match command {
                Some(GenCommands::Swift { output }) => commands::gen::swift::run_tui(output).await,
                Some(GenCommands::Kotlin { output }) => commands::gen::kotlin::run_tui(output).await,
                Some(GenCommands::Typescript { output }) => commands::gen::typescript::run_tui(output).await,
                Some(GenCommands::All) => commands::gen::all::run_tui().await,
                None => commands::gen::run_main_tui().await,
            }
        }
        Commands::Deploy { environment } => {
            commands::deploy::run_interactive(environment).await
        }
        Commands::Tui => {
            commands::tui::run_main_app().await
        }
        }
    }
}

fn setup_tracing(verbose: bool) -> Result<()> {
    let filter = if verbose { "debug" } else { "info" };

    tracing_subscriber::fmt()
        .with_env_filter(filter)
        .with_target(false)
        .init();

    Ok(())
}

mod commands;