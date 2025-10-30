pub mod cli;
pub mod error;
pub mod config;

pub mod cmd {
    pub mod auth;
    pub mod start;
    pub mod stop;
    pub mod status;
    pub mod logs;
    pub mod db;
    pub mod completions;
}

pub mod core {
    pub mod auth {
        pub mod github;
        pub mod token;
        pub mod org_check;
    }

    pub mod docker {
        pub mod manager;
        pub mod services;
        pub mod health;
        pub mod network;
        pub mod registry;
    }

    pub mod git {
        pub mod branch;
    }

    pub mod start {
        pub mod runner;
    }

    pub mod stop {
        pub mod runner;
    }

    pub mod status {
        pub mod reporter;
    }

    pub mod config {
        pub mod loader;
        pub mod types;
    }
}

pub mod utils {
    pub mod terminal;
    pub mod retry;
}

use anyhow::Result;
use cli::{Cli, Commands};

pub async fn execute(cli: Cli, config: config::Config) -> Result<()> {
    match cli.command {
        Commands::Login { no_browser } => cmd::auth::login(no_browser, &config).await,
        Commands::Logout => cmd::auth::logout().await,
        Commands::Auth(auth_cmd) => match auth_cmd {
            cli::AuthCommands::Status => cmd::auth::status().await,
        },
        Commands::Start(args) => cmd::start::run(args, config).await,
        Commands::Stop(args) => cmd::stop::run(args, config).await,
        Commands::Status(args) => cmd::status::run(args, config).await,
        Commands::Logs(args) => cmd::logs::run(args, config).await,
        Commands::Db(db_cmd) => match db_cmd {
            cli::DbCommands::Migrate => cmd::db::migrate(config).await,
            cli::DbCommands::Reset { force } => cmd::db::reset(force, config).await,
            cli::DbCommands::Seed => cmd::db::seed(config).await,
        },
        Commands::Completions { shell } => {
            cmd::completions::generate(shell);
            Ok(())
        }
    }
}
