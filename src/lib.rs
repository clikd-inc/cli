pub mod cli;
pub mod config;
pub mod error;

pub mod cmd {
    pub mod auth;
    pub mod completions;
    pub mod init;
    pub mod start;
    pub mod status;
    pub mod stop;
    pub mod update;

    pub mod release {
        pub mod init;
        pub mod status;
        pub mod prepare;
    }
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
        pub mod clikd_utils;
    }

    pub mod release {
        pub mod changelog;
        pub mod config;
        pub mod env;
        pub mod errors;
        pub mod graph;
        pub mod project;
        pub mod repository;
        pub mod rewriters;
        pub mod session;
        pub mod version;
    }

    pub mod ecosystem {
        pub mod cargo;
        pub mod npm;
        pub mod pypa;
        pub mod go;
        pub mod elixir;
        #[cfg(feature = "csharp")]
        pub mod csproj;
    }

    pub mod github {
        pub mod client;
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

    pub mod status;

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
    pub mod version_check;
}

use anyhow::Result;
use cli::{Cli, Commands};

pub async fn execute(cli: Cli) -> Result<()> {
    let command = cli.command.expect("Command must be present");
    match command {
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
            cmd::status::run(args, config).await.map_err(Into::into)
        }
        Commands::Update(args) => cmd::update::run(args).await,
        Commands::Completions { shell } => {
            cmd::completions::generate(shell);
            Ok(())
        }
        Commands::Release(release_cmd) => match release_cmd {
            cli::ReleaseCommands::Init { force, upstream } => {
                let exit_code = cmd::release::init::run(force, upstream)?;
                if exit_code != 0 {
                    std::process::exit(exit_code);
                }
                Ok(())
            }
            cli::ReleaseCommands::Status => {
                let exit_code = cmd::release::status::run()?;
                if exit_code != 0 {
                    std::process::exit(exit_code);
                }
                Ok(())
            }
            cli::ReleaseCommands::Prepare { bump } => {
                let exit_code = cmd::release::prepare::run(bump)?;
                if exit_code != 0 {
                    std::process::exit(exit_code);
                }
                Ok(())
            }
        },
    }
}
