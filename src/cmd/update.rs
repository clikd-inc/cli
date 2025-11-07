use anyhow::Result;
use crate::cli::UpdateArgs;
use crate::core::config::{images, version_manager::{VersionManager, compare_versions}};
use crate::utils::theme::*;
use dialoguer::Confirm;

pub async fn run(args: UpdateArgs) -> Result<()> {
    println!("{}", header("Checking for updates"));

    let version_mgr = VersionManager::new(None);

    if !version_mgr.has_pinned_versions() {
        println!("\n{}", warning_message("No pinned versions found. This project might not have been initialized with version pinning."));
        println!("Run {} to initialize.", highlight("clikd init"));
        return Ok(());
    }

    let local_versions = version_mgr.load_all_image_versions();
    let dockerfile_images = images::get_all_images();

    let diffs = compare_versions(&local_versions, &dockerfile_images);

    if diffs.is_empty() {
        println!("\n{}", success_message("All services are up to date!"));
        return Ok(());
    }

    let outdated: Vec<_> = diffs.iter().filter(|d| d.is_outdated()).collect();

    if outdated.is_empty() {
        println!("\n{}", success_message("All services are up to date!"));
        return Ok(());
    }

    println!("\n{}", step_message("Available updates:"));
    for diff in &outdated {
        println!("  {} {} → {}",
            highlight(&diff.service),
            dimmed(&diff.local_version),
            highlight(&diff.latest_version)
        );
    }

    let should_update = if args.yes {
        true
    } else {
        Confirm::new()
            .with_prompt("\nUpdate all services to latest versions?")
            .default(true)
            .interact()?
    };

    if !should_update {
        println!("\n{}", dimmed("Update cancelled."));
        return Ok(());
    }

    println!("\n{}", step_message("Updating service versions..."));

    version_mgr.save_image_versions(&dockerfile_images)?;

    println!("\n{}", success_message("Successfully updated all services!"));
    println!("\n{}", dimmed("Run `clikd start` to use the new versions."));

    Ok(())
}
