use anyhow::Result;
use crate::{cli::StopArgs, config::Config};
use crate::core::docker::manager::DockerManager;
use crate::utils::theme::*;

pub async fn run(args: StopArgs, config: Config) -> Result<()> {
    println!("{}", header("Stopping Clikd"));

    let docker = DockerManager::new()?;

    let mut sp = create_spinner("Stopping containers...");

    let keep_volumes = !args.purge;
    match docker.stop_all_containers(&config.project_id, keep_volumes).await {
        Ok(_) => {
            sp.success("All containers stopped!");
        }
        Err(e) => {
            sp.fail("Failed to stop containers");
            return Err(e.into());
        }
    }

    println!("\n{}", success_message(&format!("Stopped {} local development setup", highlight("clikd"))));

    if keep_volumes {
        println!("\n{}", info_message("Docker volumes have been preserved"));
        println!("{}", dimmed(&format!("  Use 'docker volume ls --filter label=com.clikd.cli.project={}' to list them", config.project_id)));
        println!("{}", dimmed("  Run with --purge to remove volumes"));
    } else {
        println!("\n{}", warning_message("All volumes have been deleted"));
    }

    Ok(())
}
