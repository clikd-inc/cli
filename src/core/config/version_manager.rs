use anyhow::{Context, Result};
use std::collections::HashMap;
use std::fs;
use std::path::{Path, PathBuf};

const TEMP_DIR: &str = "clikd/.temp";

pub struct VersionManager {
    temp_dir: PathBuf,
}

impl VersionManager {
    pub fn new(project_root: Option<&Path>) -> Self {
        let temp_dir = if let Some(root) = project_root {
            root.join(TEMP_DIR)
        } else {
            PathBuf::from(TEMP_DIR)
        };
        Self { temp_dir }
    }

    pub fn ensure_temp_dir(&self) -> Result<()> {
        fs::create_dir_all(&self.temp_dir).context("Failed to create .temp directory")?;
        Ok(())
    }

    pub fn save_cli_version(&self) -> Result<()> {
        let version = env!("CARGO_PKG_VERSION");
        let path = self.temp_dir.join("cli-version");
        fs::write(path, version).context("Failed to write CLI version")?;
        Ok(())
    }

    pub fn save_image_versions(&self, images: &HashMap<String, String>) -> Result<()> {
        self.ensure_temp_dir()?;

        for (service, image) in images {
            let version = Self::extract_version(image)?;
            let path = self.temp_dir.join(format!("{}-version", service));
            fs::write(path, version)
                .with_context(|| format!("Failed to write version for {}", service))?;
        }

        self.save_cli_version()?;
        Ok(())
    }

    pub fn load_image_version(&self, service: &str) -> Option<String> {
        let path = self.temp_dir.join(format!("{}-version", service));
        fs::read_to_string(path).ok().map(|v| v.trim().to_string())
    }

    pub fn load_all_image_versions(&self) -> HashMap<String, String> {
        let services = [
            "gate", "rig", "studio", "postgres", "keydb", "scylladb", "minio", "nats", "apisix",
        ];
        let mut versions = HashMap::new();

        for service in services {
            if let Some(version) = self.load_image_version(service) {
                versions.insert(service.to_string(), version);
            }
        }

        versions
    }

    pub fn has_pinned_versions(&self) -> bool {
        self.temp_dir.join("gate-version").exists()
    }

    fn extract_version(image: &str) -> Result<String> {
        image
            .rsplit_once(':')
            .map(|(_, version)| version.to_string())
            .context("Invalid image format, expected format: 'image:version'")
    }
}

pub fn compare_versions(
    local: &HashMap<String, String>,
    dockerfile: &HashMap<String, String>,
) -> Vec<VersionDiff> {
    let mut diffs = Vec::new();

    for (service, dockerfile_image) in dockerfile {
        if let Some(local_version) = local.get(service) {
            if let Some((_, dockerfile_version)) = dockerfile_image.rsplit_once(':') {
                if local_version != dockerfile_version {
                    diffs.push(VersionDiff {
                        service: service.clone(),
                        local_version: local_version.clone(),
                        latest_version: dockerfile_version.to_string(),
                    });
                }
            }
        }
    }

    diffs
}

#[derive(Debug, Clone)]
pub struct VersionDiff {
    pub service: String,
    pub local_version: String,
    pub latest_version: String,
}

impl VersionDiff {
    pub fn is_outdated(&self) -> bool {
        version_compare(&self.local_version, &self.latest_version) < 0
    }
}

fn version_compare(v1: &str, v2: &str) -> i32 {
    let parts1: Vec<&str> = v1.split('.').collect();
    let parts2: Vec<&str> = v2.split('.').collect();

    for i in 0..parts1.len().max(parts2.len()) {
        let p1 = parts1
            .get(i)
            .and_then(|s| s.parse::<u32>().ok())
            .unwrap_or(0);
        let p2 = parts2
            .get(i)
            .and_then(|s| s.parse::<u32>().ok())
            .unwrap_or(0);

        if p1 < p2 {
            return -1;
        } else if p1 > p2 {
            return 1;
        }
    }

    0
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_version() {
        let version = VersionManager::extract_version("ghcr.io/clikd-inc/gate:1.0.0")
            .expect("BUG: test image string should extract version");
        assert_eq!(version, "1.0.0");
    }

    #[test]
    fn test_version_compare() {
        assert_eq!(version_compare("1.0.0", "1.0.1"), -1);
        assert_eq!(version_compare("1.0.1", "1.0.0"), 1);
        assert_eq!(version_compare("1.0.0", "1.0.0"), 0);
        assert_eq!(version_compare("2.0.0", "1.9.9"), 1);
    }

    #[test]
    fn test_version_diff_is_outdated() {
        let diff = VersionDiff {
            service: "gate".to_string(),
            local_version: "1.0.0".to_string(),
            latest_version: "1.1.0".to_string(),
        };
        assert!(diff.is_outdated());

        let diff_same = VersionDiff {
            service: "gate".to_string(),
            local_version: "1.1.0".to_string(),
            latest_version: "1.1.0".to_string(),
        };
        assert!(!diff_same.is_outdated());
    }
}
