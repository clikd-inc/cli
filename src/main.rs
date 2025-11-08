use anyhow::Result;
use clap::Parser;
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

    let (result, exit_code) = if let Some(command) = cli.command {
        clikd::core::root::pre_execute();

        let res = clikd::execute(clikd::cli::Cli {
            verbose: cli.verbose,
            no_color: cli.no_color,
            env: cli.env,
            version: false,
            command: Some(command),
        }).await;

        let code = if res.is_err() { 1 } else { 0 };
        (res, code)
    } else {
        eprintln!("Error: No command provided. Use --help for usage information.");
        (Ok(()), 1)
    };

    clikd::utils::version_check::check_for_updates(env!("CARGO_PKG_VERSION"), false);

    if exit_code != 0 {
        std::process::exit(exit_code);
    }

    result
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
