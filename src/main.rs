use anyhow::Result;
use clap::Parser;
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> Result<()> {
    let cli = clikd_cli::cli::Cli::parse();
    init_logging(cli.verbose);

    if cli.no_color {
        owo_colors::set_override(false);
    }

    clikd_cli::core::root::pre_execute();

    clikd_cli::execute(cli).await
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
