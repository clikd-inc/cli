use anyhow::{Result, Context};
use serde::{Deserialize, Serialize};
use std::path::Path;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClikdConfig {
    pub project: ProjectConfig,
    pub git: GitConfig,
    pub registry: RegistryConfig,
    pub services: HashMap<String, ServiceConfig>,
    pub databases: DatabasesConfig,
    pub codegen: CodegenConfig,
    pub clients: ClientsConfig,
    pub deployment: DeploymentConfig,
    pub development: DevelopmentConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectConfig {
    pub name: String,
    pub monorepo_root: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GitConfig {
    pub main_branch: String,
    pub auto_detect_branch: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegistryConfig {
    pub url: String,
    pub organization: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceConfig {
    pub image: String,
    pub port: u16,
    pub grpc_port: Option<u16>,
    pub health_endpoint: Option<String>,
    pub dependencies: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DatabasesConfig {
    pub postgresql: PostgreSQLConfig,
    pub scylladb: ScyllaDBConfig,
    pub keydb: KeyDBConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PostgreSQLConfig {
    pub port: u16,
    pub user: String,
    pub password: String,
    pub databases: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScyllaDBConfig {
    pub port: u16,
    pub keyspace_prefix: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KeyDBConfig {
    pub port: u16,
    pub database_prefix: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CodegenConfig {
    pub openapi_endpoint: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientsConfig {
    pub swift: ClientConfig,
    pub kotlin: ClientConfig,
    pub typescript: ClientConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientConfig {
    pub output: String,
    pub package: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeploymentConfig {
    pub kubectl_context: String,
    pub namespace_prefix: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DevelopmentConfig {
    pub auto_migrate: bool,
    pub auto_seed: bool,
    pub hot_reload: bool,
}

impl ClikdConfig {
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let content = std::fs::read_to_string(&path)
            .with_context(|| format!("Failed to read config file: {}", path.as_ref().display()))?;

        let config: ClikdConfig = toml::from_str(&content)
            .with_context(|| format!("Failed to parse config file: {}", path.as_ref().display()))?;

        Ok(config)
    }

    pub fn load_or_default() -> Result<Self> {
        let config_path = std::env::current_dir()?.join("clikd.toml");

        if config_path.exists() {
            Self::load(config_path)
        } else {
            Ok(Self::default())
        }
    }

    pub fn save<P: AsRef<Path>>(&self, path: P) -> Result<()> {
        let content = toml::to_string_pretty(self)
            .context("Failed to serialize config")?;

        std::fs::write(&path, content)
            .with_context(|| format!("Failed to write config file: {}", path.as_ref().display()))?;

        Ok(())
    }

    pub fn get_service_image(&self, service: &str, branch: &str) -> String {
        let base_image = if let Some(service_config) = self.services.get(service) {
            service_config.image.clone()
        } else {
            format!("ghcr.io/{}/{}", self.registry.organization, service)
        };

        let sanitized_branch = crate::git::sanitize_branch_name(branch);
        if sanitized_branch == "main" {
            format!("{}:latest", base_image)
        } else {
            format!("{}:{}", base_image, sanitized_branch)
        }
    }

    pub fn get_database_name(&self, db_type: &str, branch: &str) -> String {
        let sanitized_branch = crate::git::sanitize_branch_name(branch);
        match db_type {
            "auth" => format!("clikd_auth_{}", sanitized_branch),
            "main" => format!("clikd_main_{}", sanitized_branch),
            _ => format!("clikd_{}_{}", db_type, sanitized_branch),
        }
    }

    pub fn get_keyspace_name(&self, branch: &str) -> String {
        let sanitized_branch = crate::git::sanitize_branch_name(branch);
        format!("{}_{}", self.databases.scylladb.keyspace_prefix, sanitized_branch)
    }

    pub fn get_keydb_prefix(&self, branch: &str) -> String {
        let sanitized_branch = crate::git::sanitize_branch_name(branch);
        format!("{}_{}", self.databases.keydb.database_prefix, sanitized_branch)
    }
}

impl Default for ClikdConfig {
    fn default() -> Self {
        let mut services = HashMap::new();

        services.insert("gate".to_string(), ServiceConfig {
            image: "ghcr.io/clikd-inc/gate".to_string(),
            port: 3001,
            grpc_port: Some(9001),
            health_endpoint: Some("/health".to_string()),
            dependencies: None,
        });

        services.insert("api".to_string(), ServiceConfig {
            image: "ghcr.io/clikd-inc/api".to_string(),
            port: 3002,
            grpc_port: Some(9002),
            health_endpoint: Some("/health".to_string()),
            dependencies: Some(vec!["gate".to_string()]),
        });

        services.insert("realtime".to_string(), ServiceConfig {
            image: "ghcr.io/clikd-inc/realtime".to_string(),
            port: 3003,
            grpc_port: Some(9003),
            health_endpoint: Some("/health".to_string()),
            dependencies: None,
        });

        services.insert("media".to_string(), ServiceConfig {
            image: "ghcr.io/clikd-inc/media".to_string(),
            port: 3004,
            grpc_port: Some(9004),
            health_endpoint: Some("/health".to_string()),
            dependencies: None,
        });

        services.insert("studio".to_string(), ServiceConfig {
            image: "ghcr.io/clikd-inc/studio".to_string(),
            port: 3000,
            grpc_port: None,
            health_endpoint: Some("/api/health".to_string()),
            dependencies: Some(vec!["api".to_string()]),
        });

        Self {
            project: ProjectConfig {
                name: "clikd".to_string(),
                monorepo_root: "../".to_string(),
            },
            git: GitConfig {
                main_branch: "main".to_string(),
                auto_detect_branch: true,
            },
            registry: RegistryConfig {
                url: "ghcr.io".to_string(),
                organization: "clikd-inc".to_string(),
            },
            services,
            databases: DatabasesConfig {
                postgresql: PostgreSQLConfig {
                    port: 5432,
                    user: "postgres".to_string(),
                    password: "dev_password".to_string(),
                    databases: Some(vec!["auth".to_string(), "main".to_string()]),
                },
                scylladb: ScyllaDBConfig {
                    port: 9042,
                    keyspace_prefix: "clikd".to_string(),
                },
                keydb: KeyDBConfig {
                    port: 6379,
                    database_prefix: "clikd".to_string(),
                },
            },
            codegen: CodegenConfig {
                openapi_endpoint: "http://localhost:3002/api/openapi.json".to_string(),
            },
            clients: ClientsConfig {
                swift: ClientConfig {
                    output: "../clients/ios".to_string(),
                    package: "ClikdAPI".to_string(),
                },
                kotlin: ClientConfig {
                    output: "../clients/android".to_string(),
                    package: "com.clikd.api".to_string(),
                },
                typescript: ClientConfig {
                    output: "../clients/web".to_string(),
                    package: "@clikd/api".to_string(),
                },
            },
            deployment: DeploymentConfig {
                kubectl_context: "clikd-cluster".to_string(),
                namespace_prefix: "clikd".to_string(),
            },
            development: DevelopmentConfig {
                auto_migrate: true,
                auto_seed: true,
                hot_reload: true,
            },
        }
    }
}