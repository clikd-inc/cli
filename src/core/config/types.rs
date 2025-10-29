use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub project: ProjectConfig,
    pub github: GitHubConfig,
    pub images: ImagesConfig,
    pub dev: DevConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ProjectConfig {
    pub name: String,
    pub organization: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct GitHubConfig {
    pub org_name: String,
    pub oauth_client_id: String,
}

#[derive(Debug, Clone, Deserialize)]
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
    pub signoz: String,
    pub zookeeper: String,
    pub clickhouse: String,
    pub schema_migrator: String,
    pub otel_collector: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DevConfig {
    pub app_env: String,
    pub rust_log: String,
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
            "signoz" => Some(&self.images.signoz),
            "zookeeper" => Some(&self.images.zookeeper),
            "clickhouse" => Some(&self.images.clickhouse),
            "schema-migrator" => Some(&self.images.schema_migrator),
            "otel-collector" => Some(&self.images.otel_collector),
            _ => None,
        }
    }
}
