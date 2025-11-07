pub mod cli;
pub mod config;
pub mod error;

pub mod cmd {
    pub mod auth;
    pub mod completions;
    pub mod db;
    pub mod init;
    pub mod logs;
    pub mod start;
    pub mod status;
    pub mod stop;
    pub mod update;
}

pub mod core {
    pub mod root;

    pub mod auth {
        pub mod github;
        pub mod org_check;
        pub mod token;
    }

    pub mod docker {
        pub mod health;
        pub mod manager;
        pub mod network;
        pub mod registry;
        pub mod services;
    }

    pub mod git {
        pub mod branch;
        pub mod gitignore;
    }

    pub mod ide {
        pub mod intellij;
        pub mod vscode;
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
        pub mod images;
        pub mod loader;
        pub mod types;
        pub mod version_manager;
    }
}

pub mod utils {
    pub mod retry;
    pub mod terminal;
    pub mod theme;
}

use anyhow::Result;
use cli::{Cli, Commands};

pub async fn execute(cli: Cli) -> Result<()> {
    match cli.command {
        Commands::Login { no_browser } => {
            let config = config::load(cli.env.as_deref())?;
            cmd::auth::login(no_browser, &config).await
        }
        Commands::Logout => cmd::auth::logout().await,
        Commands::Auth(auth_cmd) => match auth_cmd {
            cli::AuthCommands::Status => cmd::auth::status().await,
        },
        Commands::Init(args) => cmd::init::run(args).await.map_err(|e| e.into()),
        Commands::Start(args) => {
            let config = config::load(cli.env.as_deref())?;
            cmd::start::run(args, config).await
        }
        Commands::Stop(args) => {
            let config = config::load(cli.env.as_deref())?;
            cmd::stop::run(args, config).await
        }
        Commands::Status(args) => {
            let config = config::load(cli.env.as_deref())?;
            cmd::status::run(args, config).await
        }
        Commands::Logs(args) => {
            let config = config::load(cli.env.as_deref())?;
            cmd::logs::run(args, config).await
        }
        Commands::Db(db_cmd) => {
            let config = config::load(cli.env.as_deref())?;
            match db_cmd {
                cli::DbCommands::Migrate => cmd::db::migrate(config).await,
                cli::DbCommands::Reset { force } => cmd::db::reset(force, config).await,
                cli::DbCommands::Seed => cmd::db::seed(config).await,
            }
        }
        Commands::Update(args) => cmd::update::run(args).await,
        Commands::Completions { shell } => {
            cmd::completions::generate(shell);
            Ok(())
        }
    }
}
