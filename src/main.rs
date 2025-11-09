use anyhow::Result;
use clap::Parser;
use owo_colors::OwoColorize;
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> Result<()> {
    let cli = clikd::cli::Cli::parse();
    init_logging(cli.verbose);

    if cli.no_color {
        owo_colors::set_override(false);
    }

    if cli.version {
        println!("clikd {}", env!("CARGO_PKG_VERSION"));
        clikd::utils::version_check::check_for_updates(env!("CARGO_PKG_VERSION"), true);
        return Ok(());
    }

    if let Some(command) = cli.command {
        clikd::core::root::pre_execute();

        let res = clikd::execute(clikd::cli::Cli {
            verbose: cli.verbose,
            no_color: cli.no_color,
            env: cli.env,
            version: false,
            command: Some(command),
        })
        .await;

        clikd::utils::version_check::check_for_updates(env!("CARGO_PKG_VERSION"), false);

        if let Err(e) = res {
            print_error(&e);
            std::process::exit(1);
        }
    } else {
        eprintln!("Error: No command provided. Use --help for usage information.");
        std::process::exit(1);
    }

    Ok(())
}

fn print_error(error: &anyhow::Error) {
    if let Some(cli_err) = error.downcast_ref::<clikd::error::CliError>() {
        match cli_err {
            clikd::error::CliError::DockerNotRunning(socket) => {
                eprintln!(
                    "{} Cannot connect to the Docker daemon at {}. Is the docker daemon running?",
                    "failed to connect to docker:".yellow(),
                    socket.bright_cyan()
                );
                eprintln!(
                    "Try running {} or install Docker Desktop: {}",
                    "orbstack".bright_green(),
                    "https://docs.docker.com/desktop".bright_blue()
                );
                return;
            }
            _ => {}
        }
    }

    eprintln!("{} {}", "Error:".red().bold(), error);
}

fn init_logging(verbosity: u8) {
    let level = match verbosity {
        0 => "warn",
        1 => "info",
        2 => "debug",
        _ => "trace",
    };

    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::new(level))
        .init();
}
