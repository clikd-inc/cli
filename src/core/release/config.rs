// Copyright 2020-2022 Peter Williams <peter@newton.cx> and collaborators
// Licensed under the MIT License.

//! The Clikd configuration file.
//!
//! Given the same input repository, Clikd should give reproducible results no
//! matter who’s running it. So we really want all configuration to be at the
//! per-repository level.

use anyhow::Context;
use std::{collections::HashMap, fs::File, io::Read, path::Path};

use crate::atry;
use crate::core::release::errors::{Error, Result};

/// The configuration file structures as explicitly serialized into the TOML
/// format.
pub mod syntax {
    use serde::{Deserialize, Serialize};
    use std::collections::HashMap;

    /// The toplevel unified configuration structure with optional sections.
    #[derive(Clone, Debug, Default, Deserialize, Serialize)]
    pub struct UnifiedConfiguration {
        /// Release management configuration (optional).
        #[serde(skip_serializing_if = "Option::is_none")]
        pub release: Option<ReleaseConfiguration>,
    }

    /// Release-specific configuration nested under [release].
    #[derive(Clone, Debug, Deserialize, Serialize)]
    pub struct ReleaseConfiguration {
        /// General per-repository configuration.
        #[serde(default)]
        pub repo: RepoConfiguration,

        /// NPM integration configuration.
        #[serde(default)]
        pub npm: NpmConfiguration,

        /// Centralized per-project configuration.
        #[serde(default)]
        pub projects: HashMap<String, ProjectConfiguration>,
    }

    /// Legacy structure for backwards compatibility.
    #[derive(Clone, Debug, Deserialize, Serialize)]
    pub struct SerializedConfiguration {
        /// General per-repository configuration.
        pub repo: RepoConfiguration,

        /// NPM integration configuration.
        #[serde(default)]
        pub npm: NpmConfiguration,

        /// Centralized per-project configuration.
        #[serde(default)]
        pub projects: HashMap<String, ProjectConfiguration>,
    }

    /// Configuration relating to the backing repository. This is applied
    /// directly to the runtime Repository instance.
    #[derive(Clone, Debug, Default, Deserialize, Serialize)]
    pub struct RepoConfiguration {
        /// Git URLs that the upstream remote might be using.
        pub upstream_urls: Vec<String>,

        /// The name of the `rc`-like branch.
        pub rc_name: Option<String>,

        /// The name of the `release`-like branch.
        pub release_name: Option<String>,

        /// The format for release tag names.
        pub release_tag_name_format: Option<String>,
    }

    /// Configuration related to the NPM integration.
    #[derive(Clone, Debug, Default, Deserialize, Serialize)]
    pub struct NpmConfiguration {
        /// A custom "resolution protocol" to use for internal dependencies; if
        /// using Yarn workspaces, `"workspace"` may be useful here.
        pub internal_dep_protocol: Option<String>,
    }

    /// Configuration relating to individual projects.
    ///
    /// Whenever possible, this configuration should be specified in per-project
    /// metadata files to preserve locality. But some pieces of configuration
    /// need to be centralized.
    #[derive(Clone, Debug, Default, Deserialize, Serialize)]
    pub struct ProjectConfiguration {
        /// Ignore this project if/when it is automatically detected.
        pub ignore: bool,
    }
}

#[derive(Clone, Debug)]
pub struct ConfigurationFile {
    pub repo: syntax::RepoConfiguration,
    pub npm: syntax::NpmConfiguration,
    pub projects: HashMap<String, syntax::ProjectConfiguration>,
}

impl Default for ConfigurationFile {
    fn default() -> Self {
        let repo = syntax::RepoConfiguration::default();
        let npm = Default::default();
        let projects = Default::default();

        ConfigurationFile {
            repo,
            npm,
            projects,
        }
    }
}

impl ConfigurationFile {
    pub fn get<P: AsRef<Path>>(path: P) -> Result<Self> {
        let mut f = match File::open(&path) {
            Ok(f) => f,

            Err(e) => {
                return if e.kind() == std::io::ErrorKind::NotFound {
                    Ok(Self::default())
                } else {
                    Err(Error::new(e).context(format!(
                        "failed to open config file `{}`",
                        path.as_ref().display()
                    )))
                }
            }
        };

        let mut text = String::new();
        f.read_to_string(&mut text)
            .with_context(|| format!("failed to read config file `{}`", path.as_ref().display()))?;

        // Try new unified format first ([release] section)
        if let Ok(unified) = toml::from_str::<syntax::UnifiedConfiguration>(&text) {
            if let Some(release_cfg) = unified.release {
                return Ok(ConfigurationFile {
                    repo: release_cfg.repo,
                    npm: release_cfg.npm,
                    projects: release_cfg.projects,
                });
            }
        }

        // Fall back to legacy format (backwards compatibility)
        let sercfg: syntax::SerializedConfiguration = toml::from_str(&text).with_context(|| {
            format!(
                "could not parse config file `{}` as TOML",
                path.as_ref().display()
            )
        })?;

        Ok(ConfigurationFile {
            repo: sercfg.repo,
            npm: sercfg.npm,
            projects: sercfg.projects,
        })
    }

    pub fn into_toml(self) -> Result<String> {
        let unified_cfg = syntax::UnifiedConfiguration {
            release: Some(syntax::ReleaseConfiguration {
                repo: self.repo,
                npm: self.npm,
                projects: self.projects,
            }),
        };
        Ok(atry!(
            toml::to_string_pretty(&unified_cfg);
            ["could not serialize configuration into TOML format"]
        ))
    }
}
