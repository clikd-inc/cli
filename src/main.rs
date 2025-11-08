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

    clikd::core::root::pre_execute();

    let result = clikd::execute(cli).await;

    clikd::utils::version_check::check_for_updates(env!("CARGO_PKG_VERSION"));

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
