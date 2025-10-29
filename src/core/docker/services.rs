use std::collections::HashMap;
use std::time::Duration;
use crate::config::Config;

pub struct ServiceDefinition {
    pub name: String,
    pub image: String,
    pub ports: Vec<(u16, u16)>,
    pub env: HashMap<String, String>,
    pub volumes: Vec<String>,
    pub health_check: Option<HealthCheck>,
    pub depends_on: Vec<String>,
    pub command: Option<Vec<String>>,
}

pub struct HealthCheck {
    pub test: Vec<String>,
    pub interval: Duration,
    pub timeout: Duration,
    pub retries: u32,
    pub start_period: Option<Duration>,
}

pub fn all_services(branch: &str, config: &Config) -> Vec<ServiceDefinition> {
    vec![
        postgres_auth_service(branch, config),
        postgres_rig_service(branch, config),
        keydb_service(branch, config),
        scylladb_service(branch, config),
        minio_service(branch, config),
        nats_service(branch, config),
        zookeeper_service(branch, config),
        clickhouse_service(branch, config),
        schema_migrator_service(branch, config),
        signoz_service(branch, config),
        otel_collector_service(branch, config),
        gate_service(branch, config),
        rig_service(branch, config),
        apisix_service(branch, config),
        studio_service(branch, config),
    ]
}

fn gate_service(branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("APP_ENV".into(), config.dev.app_env.clone());
    env.insert("RUST_LOG".into(), config.dev.rust_log.clone());
    env.insert("HOST".into(), "0.0.0.0".into());
    env.insert("PORT".into(), "8081".into());
    env.insert("DATABASE_URL".into(), "postgresql://postgres:development@postgres-auth:5432/clikd_auth".into());
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());
    env.insert("NATS_URL".into(), "nats://nats:4222".into());
    env.insert("OTEL_EXPORTER_OTLP_ENDPOINT".into(), "http://otel-collector:4317".into());
    env.insert("JWT_SECRET".into(), "dev-jwt-secret-32-bytes-long-enough-for-testing-abc123".into());
    env.insert("ENC_KEY_ACTIVE".into(), "gate1".into());
    env.insert("ENC_KEYS".into(), "gate1/MUKfFPL1zfhKfffX7usQbeWKd5L9iH65K4kCi7B3/KU=".into());
    env.insert("COOKIE_SECRET".into(), "dev-cookie-secret-32-bytes-long-enough-for-testing-def456".into());
    env.insert("INTERNAL_API_SECRET".into(), "dev-internal-api-secret-change-this".into());
    env.insert("PUBLIC_URL".into(), "http://localhost:8081".into());
    env.insert("ISSUER".into(), "http://localhost:8081".into());
    env.insert("RIG_INTERNAL_URL".into(), "http://rig:8082".into());
    env.insert("BACKEND_API_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());
    env.insert("GATE_ANON_KEY".into(), "gt_publishable_dev_aNonymOusK3yForPubl1cAccess".into());
    env.insert("GATE_SECRET_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());
    env.insert("CONFIG_FILE".into(), "/config/gate.env".into());

    ServiceDefinition {
        name: "gate".into(),
        image: config.images.gate.clone(),
        ports: vec![(8081, 8081), (9001, 9001)],
        env,
        volumes: vec![format!("clikd_gate_config_{}:/config", branch)],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:8081/health".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["postgres-auth".into(), "keydb".into()],
        command: None,
    }
}

fn rig_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("APP_ENV".into(), config.dev.app_env.clone());
    env.insert("RUST_LOG".into(), config.dev.rust_log.clone());
    env.insert("PORT".into(), "8082".into());
    env.insert("GRPC_PORT".into(), "9002".into());
    env.insert("DATABASE_URL".into(), "postgresql://postgres:development@postgres-rig:5432/clikd_rig".into());
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());
    env.insert("SCYLLADB_HOSTS".into(), "scylladb:9042".into());
    env.insert("NATS_URL".into(), "nats://nats:4222".into());
    env.insert("MINIO_ENDPOINT".into(), "http://minio:9000".into());
    env.insert("MINIO_ROOT_USER".into(), "minioadmin".into());
    env.insert("MINIO_ROOT_PASSWORD".into(), "minioadmin".into());
    env.insert("OTEL_EXPORTER_OTLP_ENDPOINT".into(), "http://otel-collector:4317".into());

    ServiceDefinition {
        name: "rig".into(),
        image: config.images.rig.clone(),
        ports: vec![(8082, 8082), (9002, 9002)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:8082/health".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["postgres-rig".into(), "keydb".into(), "scylladb".into(), "nats".into(), "minio".into()],
        command: None,
    }
}

fn studio_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("NODE_ENV".into(), "development".into());
    env.insert("APP_ENV".into(), config.dev.app_env.clone());
    env.insert("CLIKD_URL".into(), "http://apisix:9080".into());
    env.insert("CLIKD_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());
    env.insert("NEXT_PUBLIC_STUDIO_URL".into(), "http://localhost:3001".into());
    env.insert("NEXT_PUBLIC_APP_ENV".into(), config.dev.app_env.clone());

    ServiceDefinition {
        name: "studio".into(),
        image: config.images.studio.clone(),
        ports: vec![(3001, 3001)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "bun".into(),
                "--eval".into(),
                "fetch('http://localhost:3001/api/health').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))".into()
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["apisix".into()],
        command: None,
    }
}

fn postgres_auth_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("POSTGRES_DB".into(), "clikd_auth".into());
    env.insert("POSTGRES_USER".into(), "postgres".into());
    env.insert("POSTGRES_PASSWORD".into(), "development".into());

    ServiceDefinition {
        name: "postgres-auth".into(),
        image: config.images.postgres.clone(),
        ports: vec![(5433, 5432)],
        env,
        volumes: vec!["clikd_postgres_auth_data:/var/lib/postgresql".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD-SHELL".into(), "pg_isready -U postgres".into()],
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
            retries: 5,
            start_period: None,
        }),
        depends_on: vec![],
        command: None,
    }
}

fn postgres_rig_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("POSTGRES_DB".into(), "clikd_rig".into());
    env.insert("POSTGRES_USER".into(), "postgres".into());
    env.insert("POSTGRES_PASSWORD".into(), "development".into());

    ServiceDefinition {
        name: "postgres-rig".into(),
        image: config.images.postgres.clone(),
        ports: vec![(5434, 5432)],
        env,
        volumes: vec!["clikd_postgres_rig_data:/var/lib/postgresql".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD-SHELL".into(), "pg_isready -U postgres".into()],
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
            retries: 5,
            start_period: None,
        }),
        depends_on: vec![],
        command: None,
    }
}

fn keydb_service(_branch: &str, config: &Config) -> ServiceDefinition {
    ServiceDefinition {
        name: "keydb".into(),
        image: config.images.keydb.clone(),
        ports: vec![(6380, 6379)],
        env: HashMap::new(),
        volumes: vec!["clikd_keydb_data:/data".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "keydb-cli".into(), "ping".into()],
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
            retries: 5,
            start_period: None,
        }),
        depends_on: vec![],
        command: Some(vec![
            "keydb-server".into(),
            "--protected-mode".into(),
            "no".into(),
            "--appendonly".into(),
            "yes".into(),
            "--server-threads".into(),
            "4".into(),
        ]),
    }
}

fn scylladb_service(_branch: &str, config: &Config) -> ServiceDefinition {
    ServiceDefinition {
        name: "scylladb".into(),
        image: config.images.scylladb.clone(),
        ports: vec![(9043, 9042), (10000, 10000)],
        env: HashMap::new(),
        volumes: vec!["clikd_scylladb_data:/var/lib/scylla".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "cqlsh".into(), "-e".into(), "describe keyspaces".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 10,
            start_period: Some(Duration::from_secs(60)),
        }),
        depends_on: vec![],
        command: Some(vec![
            "--smp".into(),
            "4".into(),
            "--memory".into(),
            "4G".into(),
            "--overprovisioned".into(),
            "1".into(),
            "--api-address".into(),
            "0.0.0.0".into(),
        ]),
    }
}

fn minio_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("MINIO_ROOT_USER".into(), "minioadmin".into());
    env.insert("MINIO_ROOT_PASSWORD".into(), "minioadmin".into());

    ServiceDefinition {
        name: "minio".into(),
        image: config.images.minio.clone(),
        ports: vec![(9000, 9000), (9901, 9001)],
        env,
        volumes: vec!["clikd_minio_data:/data".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:9000/minio/health/live".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(20),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec![],
        command: Some(vec!["server".into(), "/data".into(), "--console-address".into(), ":9001".into()]),
    }
}

fn nats_service(_branch: &str, config: &Config) -> ServiceDefinition {
    ServiceDefinition {
        name: "nats".into(),
        image: config.images.nats.clone(),
        ports: vec![(4222, 4222), (8222, 8222)],
        env: HashMap::new(),
        volumes: vec!["clikd_nats_data:/data".into()],
        health_check: None,
        depends_on: vec![],
        command: Some(vec![
            "-js".into(),
            "-m".into(),
            "8222".into(),
            "-p".into(),
            "4222".into(),
            "--store_dir=/data".into(),
        ]),
    }
}

fn zookeeper_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("ALLOW_ANONYMOUS_LOGIN".into(), "yes".into());
    env.insert("ZOO_AUTOPURGE_INTERVAL".into(), "1".into());

    ServiceDefinition {
        name: "zookeeper-1".into(),
        image: config.images.zookeeper.clone(),
        ports: vec![(2181, 2181), (2888, 2888), (3888, 3888), (8094, 8080)],
        env,
        volumes: vec!["clikd_zookeeper_data:/bitnami/zookeeper".into()],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "zkServer.sh".into(), "status".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec![],
        command: None,
    }
}

fn clickhouse_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("CLICKHOUSE_DB".into(), "signoz_traces".into());
    env.insert("CLICKHOUSE_SKIP_USER_SETUP".into(), "1".into());

    ServiceDefinition {
        name: "clickhouse".into(),
        image: config.images.clickhouse.clone(),
        ports: vec![(8123, 8123), (9100, 9000), (9009, 9009)],
        env,
        volumes: vec!["clikd_clickhouse_data:/var/lib/clickhouse/".into()],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "wget".into(),
                "--no-verbose".into(),
                "--tries=1".into(),
                "--spider".into(),
                "http://localhost:8123/?query=SELECT%201".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["zookeeper-1".into()],
        command: None,
    }
}

fn schema_migrator_service(_branch: &str, config: &Config) -> ServiceDefinition {
    ServiceDefinition {
        name: "schema-migrator".into(),
        image: config.images.schema_migrator.clone(),
        ports: vec![],
        env: HashMap::new(),
        volumes: vec![],
        health_check: None,
        depends_on: vec!["clickhouse".into()],
        command: Some(vec![
            "sync".into(),
            "--dsn=tcp://clickhouse:9000".into(),
            "--up=".into(),
        ]),
    }
}

fn signoz_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("SIGNOZ_TELEMETRYSTORE_CLICKHOUSE_DSN".into(), "tcp://clickhouse:9000".into());
    env.insert("SIGNOZ_SQLSTORE_SQLITE_PATH".into(), "/var/lib/signoz/signoz.db".into());
    env.insert("SIGNOZ_JWT_SECRET".into(), "development-jwt-secret-change-in-production".into());
    env.insert("STORAGE".into(), "clickhouse".into());
    env.insert("DEPLOYMENT_TYPE".into(), "docker-standalone-amd".into());

    ServiceDefinition {
        name: "signoz".into(),
        image: config.images.signoz.clone(),
        ports: vec![(3301, 8080)],
        env,
        volumes: vec!["clikd_signoz_data:/var/lib/signoz/".into()],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "wget".into(),
                "--no-verbose".into(),
                "--tries=1".into(),
                "--spider".into(),
                "http://localhost:3301/api/v1/health".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["schema-migrator".into()],
        command: None,
    }
}

fn otel_collector_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("OTEL_RESOURCE_ATTRIBUTES".into(), "service.name=signoz-otel-collector,service.version=v0.129.6".into());

    ServiceDefinition {
        name: "otel-collector".into(),
        image: config.images.otel_collector.clone(),
        ports: vec![(4317, 4317), (4318, 4318), (14250, 14250), (9411, 9411)],
        env,
        volumes: vec![],
        health_check: None,
        depends_on: vec!["clickhouse".into()],
        command: None,
    }
}

fn apisix_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("GATE_HOST".into(), "gate".into());
    env.insert("RIG_HOST".into(), "rig".into());

    ServiceDefinition {
        name: "apisix".into(),
        image: config.images.apisix.clone(),
        ports: vec![(9080, 9080)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:9080/".into()],
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["gate".into(), "rig".into()],
        command: None,
    }
}
