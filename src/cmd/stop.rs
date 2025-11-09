use crate::core::docker::manager::DockerManager;
use crate::utils::theme::*;
use crate::{cli::StopArgs, config::Config};
use anyhow::Result;

pub async fn run(args: StopArgs, config: Config) -> Result<()> {
    println!("{}", header("Stopping Clikd"));

    let docker = DockerManager::new()?;

    if !docker.is_docker_running().await {
        let socket = std::env::var("DOCKER_HOST")
            .unwrap_or_else(|_| "unix:///var/run/docker.sock".to_string());
        return Err(crate::error::CliError::DockerNotRunning(socket).into());
    }

    let mut sp = create_spinner("Stopping containers...");

    let keep_volumes = !args.purge;
    match docker
        .stop_all_containers(&config.project_id, keep_volumes)
        .await
    {
        Ok(_) => {
            sp.success("All containers stopped!");
        }
        Err(e) => {
            sp.fail("Failed to stop containers");
            return Err(e.into());
        }
    }

    println!(
        "\n{}",
        success_message(&format!(
            "Stopped {} local development setup",
            highlight("clikd")
        ))
    );

    if keep_volumes {
        println!("\n{}", info_message("Docker volumes have been preserved"));
        println!(
            "{}",
            dimmed(&format!(
                "  Use 'docker volume ls --filter label=com.clikd.cli.project={}' to list them",
                config.project_id
            ))
        );
        println!("{}", dimmed("  Run with --purge to remove volumes"));
    } else {
        println!("\n{}", warning_message("All volumes have been deleted"));
    }

    Ok(())
}
