use crate::utils::theme::*;
use serde::Deserialize;
use std::env;
use std::fs;

#[derive(Debug, Deserialize)]
struct GithubRelease {
    tag_name: String,
    html_url: String,
}

pub fn pre_execute() {
    write_cli_version();
    check_for_updates();
}

pub fn write_cli_version() {
    if let Ok(cwd) = env::current_dir() {
        let temp_dir = cwd.join("clikd/.temp");
        let _ = fs::create_dir_all(&temp_dir);
        let version_file = temp_dir.join("cli-latest");
        let version = format!("v{}", env!("CARGO_PKG_VERSION"));
        let _ = fs::write(version_file, version);
    }
}

pub fn check_for_updates() {
    let current_version = env!("CARGO_PKG_VERSION");

    let result = std::thread::spawn(move || {
        let token = match crate::core::auth::token::load_token() {
            Ok(t) => t,
            Err(_) => return,
        };

        let config = ureq::Agent::config_builder()
            .timeout_global(Some(std::time::Duration::from_secs(2)))
            .build();
        let agent: ureq::Agent = config.into();

        let mut response = match agent
            .get("https://api.github.com/repos/clikd-inc/clikd/releases")
            .header("User-Agent", &format!("clikd/{}", current_version))
            .header("Authorization", &format!("Bearer {}", token))
            .call()
        {
            Ok(r) => r,
            Err(_) => return,
        };

        let releases: Vec<GithubRelease> = match response.body_mut().read_json() {
            Ok(r) => r,
            Err(_) => return,
        };

        let cli_release = releases.iter().find(|r| r.tag_name.starts_with("cli-v"));

        if let Some(latest) = cli_release {
            let latest_version = latest.tag_name.trim_start_matches("cli-v");

            if latest_version != current_version {
                eprintln!(
                    "\n{} {} → {}\n{} {}\n",
                    warning_message("Update available"),
                    dimmed(current_version),
                    highlight(latest_version),
                    dimmed("Release:"),
                    url(&latest.html_url)
                );
            }
        }
    })
    .join();

    let _ = result;
}
