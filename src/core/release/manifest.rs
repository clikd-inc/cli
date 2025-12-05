use serde::{Deserialize, Serialize};
use std::fs;
use std::path::Path;
use time::OffsetDateTime;

const SCHEMA_VERSION: &str = "1.0";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReleaseManifest {
    pub schema_version: String,
    pub created_at: String,
    pub created_by: String,
    pub base_branch: String,
    pub releases: Vec<ProjectRelease>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectRelease {
    pub name: String,
    pub ecosystem: String,
    pub previous_version: String,
    pub new_version: String,
    pub bump_type: String,
    pub changelog: String,
    pub tag_name: String,
    pub prefix: String,
}

impl ReleaseManifest {
    pub fn new(base_branch: String, created_by: String) -> Self {
        let now = OffsetDateTime::now_utc();
        let format = time::format_description::well_known::Rfc3339;
        let created_at = now.format(&format).unwrap_or_else(|_| now.to_string());

        Self {
            schema_version: SCHEMA_VERSION.to_string(),
            created_at,
            created_by,
            base_branch,
            releases: Vec::new(),
        }
    }

    pub fn add_release(&mut self, release: ProjectRelease) {
        self.releases.push(release);
    }

    pub fn to_json(&self) -> Result<String, serde_json::Error> {
        serde_json::to_string_pretty(self)
    }

    pub fn from_json(json: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(json)
    }

    pub fn save_to_file(&self, path: &Path) -> Result<(), std::io::Error> {
        let json = self
            .to_json()
            .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e.to_string()))?;

        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }

        fs::write(path, json)
    }

    pub fn generate_filename() -> String {
        let now = OffsetDateTime::now_utc();
        let format = time::format_description::parse("[year][month][day]-[hour][minute][second]")
            .expect("valid format");
        format!(
            "release-{}.json",
            now.format(&format).expect("format datetime")
        )
    }
}

impl ProjectRelease {
    pub fn new(
        name: String,
        ecosystem: String,
        previous_version: String,
        new_version: String,
        bump_type: String,
        changelog: String,
        prefix: String,
    ) -> Self {
        let tag_name = if prefix.is_empty() {
            format!("v{new_version}")
        } else {
            format!("{prefix}/v{new_version}")
        };

        Self {
            name,
            ecosystem,
            previous_version,
            new_version,
            bump_type,
            changelog,
            tag_name,
            prefix,
        }
    }
}
