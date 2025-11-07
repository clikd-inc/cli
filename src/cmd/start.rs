use anyhow::Result;
use crate::{cli::StartArgs, config::Config};
use crate::core::config::{images, version_manager::{VersionManager, compare_versions}};
use crate::core::start::runner;
use crate::utils::theme::*;

pub async fn run(args: StartArgs, config: Config) -> Result<()> {
    check_version_diff();

    let exclude = args.exclude.unwrap_or_default();
    runner::run(&config, exclude, args.ignore_health_check).await?;
    Ok(())
}

fn check_version_diff() {
    let version_mgr = VersionManager::new(None);

    if !version_mgr.has_pinned_versions() {
        return;
    }

    let local_versions = version_mgr.load_all_image_versions();
    let dockerfile_images = images::get_all_images();

    let diffs = compare_versions(&local_versions, &dockerfile_images);

    if !diffs.is_empty() {
        let outdated: Vec<_> = diffs.iter().filter(|d| d.is_outdated()).collect();

        if !outdated.is_empty() {
            eprintln!("\n{} You are running different service versions locally than the latest CLI:\n",
                warning_message("WARNING:")
            );

            for diff in &outdated {
                eprintln!("  {} {} → {}",
                    highlight(&diff.service),
                    dimmed(&diff.local_version),
                    highlight(&diff.latest_version)
                );
            }

            eprintln!("\n  Run {} to update them.\n",
                highlight("clikd update")
            );
        }
    }
}
