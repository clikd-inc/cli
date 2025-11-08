use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    #[serde(default = "default_empty_string")]
    pub project_id: String,
    #[serde(default = "default_github_org_name")]
    pub github_org_name: String,
    #[serde(default = "default_github_oauth_client_id")]
    pub github_oauth_client_id: String,
    #[serde(default)]
    pub services: ServicesConfig,
    #[serde(default)]
    pub ports: PortsConfig,
    #[serde(skip)]
    pub images: ImagesConfig,
    #[serde(default)]
    pub dev: DevConfig,
    #[serde(default)]
    pub workdir: WorkdirConfig,
}

fn default_empty_string() -> String {
    String::new()
}

fn default_github_org_name() -> String {
    "clikd-inc".to_string()
}

fn default_github_oauth_client_id() -> String {
    "Ov23liNPpcjTMYaP841Y".to_string()
}

#[derive(Debug, Clone, Deserialize)]
#[serde(default)]
pub struct ImagesConfig {
    pub gate: String,
    pub rig: String,
    pub studio: String,
    pub postgres: String,
    pub keydb: String,
    pub scylladb: String,
    pub minio: String,
    pub nats: String,
    pub apisix: String,
}

impl Default for ImagesConfig {
    fn default() -> Self {
        use super::{images, version_manager::VersionManager};

        let version_mgr = VersionManager::new(None);

        let gate = if let Some(version) = version_mgr.load_image_version("gate") {
            format!("ghcr.io/clikd-inc/gate:{}", version)
        } else {
            images::get_image("gate").expect("gate image not found in Dockerfile")
        };

        let rig = if let Some(version) = version_mgr.load_image_version("rig") {
            format!("ghcr.io/clikd-inc/rig:{}", version)
        } else {
            images::get_image("rig").expect("rig image not found in Dockerfile")
        };

        let studio = if let Some(version) = version_mgr.load_image_version("studio") {
            format!("ghcr.io/clikd-inc/studio:{}", version)
        } else {
            images::get_image("studio").expect("studio image not found in Dockerfile")
        };

        let postgres = if let Some(version) = version_mgr.load_image_version("postgres") {
            format!("ghcr.io/clikd-inc/postgres:{}", version)
        } else {
            images::get_image("postgres").expect("postgres image not found in Dockerfile")
        };

        let keydb = if let Some(version) = version_mgr.load_image_version("keydb") {
            format!("ghcr.io/clikd-inc/keydb:{}", version)
        } else {
            images::get_image("keydb").expect("keydb image not found in Dockerfile")
        };

        let scylladb = if let Some(version) = version_mgr.load_image_version("scylladb") {
            format!("ghcr.io/clikd-inc/scylladb:{}", version)
        } else {
            images::get_image("scylladb").expect("scylladb image not found in Dockerfile")
        };

        let minio = if let Some(version) = version_mgr.load_image_version("minio") {
            format!("ghcr.io/clikd-inc/minio:{}", version)
        } else {
            images::get_image("minio").expect("minio image not found in Dockerfile")
        };

        let nats = if let Some(version) = version_mgr.load_image_version("nats") {
            format!("ghcr.io/clikd-inc/nats:{}", version)
        } else {
            images::get_image("nats").expect("nats image not found in Dockerfile")
        };

        let apisix = if let Some(version) = version_mgr.load_image_version("apisix") {
            format!("ghcr.io/clikd-inc/apisix:{}", version)
        } else {
            images::get_image("apisix").expect("apisix image not found in Dockerfile")
        };

        Self {
            gate,
            rig,
            studio,
            postgres,
            keydb,
            scylladb,
            minio,
            nats,
            apisix,
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
#[serde(default)]
pub struct DevConfig {
    pub app_env: String,
    pub rust_log: String,
}

impl Default for DevConfig {
    fn default() -> Self {
        Self {
            app_env: "development".to_string(),
            rust_log: "debug".to_string(),
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServicesConfig {
    #[serde(default = "default_true")]
    pub gate: bool,
    #[serde(default = "default_true")]
    pub rig: bool,
    #[serde(default = "default_true")]
    pub studio: bool,
    #[serde(default = "default_true")]
    pub postgres_auth: bool,
    #[serde(default = "default_true")]
    pub postgres_rig: bool,
    #[serde(default = "default_true")]
    pub keydb: bool,
    #[serde(default = "default_true")]
    pub scylladb: bool,
    #[serde(default = "default_true")]
    pub minio: bool,
    #[serde(default = "default_true")]
    pub nats: bool,
    #[serde(default = "default_true")]
    pub apisix: bool,
}

#[derive(Debug, Clone, Deserialize)]
pub struct PortsConfig {
    #[serde(default = "default_gate_port")]
    pub gate: u16,
    #[serde(default = "default_rig_port")]
    pub rig: u16,
    #[serde(default = "default_studio_port")]
    pub studio: u16,
    #[serde(default = "default_postgres_auth_port")]
    pub postgres_auth: u16,
    #[serde(default = "default_postgres_rig_port")]
    pub postgres_rig: u16,
    #[serde(default = "default_keydb_port")]
    pub keydb: u16,
    #[serde(default = "default_scylladb_port")]
    pub scylladb: u16,
    #[serde(default = "default_minio_port")]
    pub minio: u16,
    #[serde(default = "default_minio_console_port")]
    pub minio_console: u16,
    #[serde(default = "default_nats_port")]
    pub nats: u16,
    #[serde(default = "default_apisix_port")]
    pub apisix: u16,
}

#[derive(Debug, Clone, Deserialize, Default)]
pub struct WorkdirConfig {
    pub path: Option<String>,
}

fn default_true() -> bool {
    true
}

fn default_gate_port() -> u16 {
    8081
}

fn default_rig_port() -> u16 {
    8082
}

fn default_studio_port() -> u16 {
    3001
}

fn default_postgres_auth_port() -> u16 {
    5433
}

fn default_postgres_rig_port() -> u16 {
    5434
}

fn default_keydb_port() -> u16 {
    6380
}

fn default_scylladb_port() -> u16 {
    9043
}

fn default_minio_port() -> u16 {
    9000
}

fn default_minio_console_port() -> u16 {
    9901
}

fn default_nats_port() -> u16 {
    4222
}

fn default_apisix_port() -> u16 {
    9080
}

impl Default for ServicesConfig {
    fn default() -> Self {
        Self {
            gate: true,
            rig: true,
            studio: true,
            postgres_auth: true,
            postgres_rig: true,
            keydb: true,
            scylladb: true,
            minio: true,
            nats: true,
            apisix: true,
        }
    }
}

impl Default for PortsConfig {
    fn default() -> Self {
        Self {
            gate: 8081,
            rig: 8082,
            studio: 3001,
            postgres_auth: 5433,
            postgres_rig: 5434,
            keydb: 6380,
            scylladb: 9043,
            minio: 9000,
            minio_console: 9901,
            nats: 4222,
            apisix: 9080,
        }
    }
}

impl Config {
    pub fn get_image(&self, service: &str) -> Option<&String> {
        match service {
            "gate" => Some(&self.images.gate),
            "rig" => Some(&self.images.rig),
            "studio" => Some(&self.images.studio),
            "postgres" => Some(&self.images.postgres),
            "keydb" => Some(&self.images.keydb),
            "scylladb" => Some(&self.images.scylladb),
            "minio" => Some(&self.images.minio),
            "nats" => Some(&self.images.nats),
            "apisix" => Some(&self.images.apisix),
            _ => None,
        }
    }

    pub fn is_service_enabled(&self, service: &str) -> bool {
        match service {
            "gate" => self.services.gate,
            "rig" => self.services.rig,
            "studio" => self.services.studio,
            "postgres_auth" => self.services.postgres_auth,
            "postgres_rig" => self.services.postgres_rig,
            "keydb" => self.services.keydb,
            "scylladb" => self.services.scylladb,
            "minio" => self.services.minio,
            "nats" => self.services.nats,
            "apisix" => self.services.apisix,
            _ => false,
        }
    }

    pub fn get_port(&self, service: &str) -> Option<u16> {
        match service {
            "gate" => Some(self.ports.gate),
            "rig" => Some(self.ports.rig),
            "studio" => Some(self.ports.studio),
            "postgres_auth" => Some(self.ports.postgres_auth),
            "postgres_rig" => Some(self.ports.postgres_rig),
            "keydb" => Some(self.ports.keydb),
            "scylladb" => Some(self.ports.scylladb),
            "minio" => Some(self.ports.minio),
            "minio_console" => Some(self.ports.minio_console),
            "nats" => Some(self.ports.nats),
            "apisix" => Some(self.ports.apisix),
            _ => None,
        }
    }

    pub fn sanitize_project_id(&mut self) {
        if self.project_id.is_empty() {
            if let Ok(cwd) = std::env::current_dir() {
                if let Some(dir_name) = cwd.file_name() {
                    if let Some(name) = dir_name.to_str() {
                        self.project_id = name.to_string();
                    }
                }
            }
            if self.project_id.is_empty() {
                self.project_id = "clikd-project".to_string();
            }
        }

        self.project_id = self
            .project_id
            .chars()
            .filter(|c| c.is_alphanumeric() || *c == '_' || *c == '-' || *c == '.')
            .take(40)
            .collect::<String>()
            .trim_start_matches(['_', '-', '.'])
            .to_string();
    }
}
