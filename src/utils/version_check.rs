use serde::Deserialize;
use std::fs;
use std::path::PathBuf;
use std::time::SystemTime;

const GITHUB_API_URL: &str = "https://api.github.com/repos/clikd-inc/cli/releases/latest";
const CHECK_INTERVAL_HOURS: u64 = 10;

#[derive(Deserialize)]
struct GithubRelease {
    tag_name: String,
}

fn is_clikd_project() -> bool {
    let cwd = std::env::current_dir().unwrap_or_else(|_| PathBuf::from("."));
    cwd.join("clikd/config.toml").is_file() || cwd.join("clikd/bootstrap.toml").is_file()
}

pub fn check_for_updates(current_version: &str, force_fetch: bool) {
    let Some(cache_path) = get_cache_path() else {
        if force_fetch {
            if let Some(latest_version) = fetch_latest_from_github() {
                if is_newer_version(&latest_version, current_version) {
                    print_update_message(&latest_version, current_version);
                }
            }
        }
        return;
    };

    if let Some(latest_version) = get_latest_version(&cache_path, force_fetch) {
        if is_newer_version(&latest_version, current_version) {
            print_update_message(&latest_version, current_version);
        }
    }
}

fn get_cache_path() -> Option<PathBuf> {
    if !is_clikd_project() {
        return None;
    }

    let mut path = std::env::current_dir().unwrap_or_else(|_| PathBuf::from("."));
    path.push("clikd");
    path.push(".temp");

    if !path.exists() {
        let _ = fs::create_dir_all(&path);
    }

    path.push("cli-latest");
    Some(path)
}

fn should_fetch_latest(cache_path: &PathBuf, force_fetch: bool) -> bool {
    if force_fetch {
        return true;
    }

    if let Ok(metadata) = fs::metadata(cache_path) {
        if let Ok(modified) = metadata.modified() {
            if let Ok(elapsed) = SystemTime::now().duration_since(modified) {
                return elapsed.as_secs() > CHECK_INTERVAL_HOURS * 3600;
            }
        }
    }
    true
}

fn get_latest_version(cache_path: &PathBuf, force_fetch: bool) -> Option<String> {
    if should_fetch_latest(cache_path, force_fetch) {
        if let Some(version) = fetch_latest_from_github() {
            let _ = fs::write(cache_path, &version);
            return Some(version);
        }
    }

    fs::read_to_string(cache_path).ok()
}

fn fetch_latest_from_github() -> Option<String> {
    let mut response = ureq::get(GITHUB_API_URL)
        .header("User-Agent", "clikd")
        .call()
        .ok()?;

    let release: GithubRelease = response.body_mut().read_json().ok()?;
    Some(release.tag_name)
}

fn is_newer_version(latest: &str, current: &str) -> bool {
    let latest_clean = latest.trim_start_matches('v');
    let current_clean = current.trim_start_matches('v');

    match (
        semver::Version::parse(latest_clean),
        semver::Version::parse(current_clean),
    ) {
        (Ok(latest_ver), Ok(current_ver)) => latest_ver > current_ver,
        _ => false,
    }
}

fn print_update_message(latest: &str, current: &str) {
    use crate::utils::theme::*;

    eprintln!();
    eprintln!(
        "{} A new version of Clikd CLI is available: {} (currently installed v{})",
        warning_message("UPDATE AVAILABLE:"),
        highlight(latest),
        current
    );
    eprintln!("   Run {} to update", highlight("brew upgrade clikd"));
    eprintln!();
}
