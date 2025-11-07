use crate::config::Config;
use minijinja::Environment;
use std::collections::HashMap;
use std::time::Duration;

const APISIX_ROUTES_TEMPLATE: &str = include_str!("../../../templates/apisix-routes.yaml");

#[derive(Clone)]
pub struct ServiceDefinition {
    pub name: String,
    pub image: String,
    pub ports: Vec<(u16, u16)>,
    pub env: HashMap<String, String>,
    pub volumes: Vec<String>,
    pub health_check: Option<HealthCheck>,
    pub depends_on: Vec<String>,
    pub entrypoint: Option<Vec<String>>,
    pub command: Option<Vec<String>>,
    pub platform: Option<String>,
}

#[derive(Clone)]
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
        gate_service(branch, config),
        rig_service(branch, config),
        apisix_service(branch, config),
        studio_service(branch, config),
    ]
}

fn gate_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();

    env.insert("GATE_HOST".into(), "0.0.0.0".into());
    env.insert("GATE_PORT".into(), "8081".into());
    env.insert("GATE_PUBLIC_URL".into(), "http://localhost:8081".into());
    env.insert("GATE_ISSUER".into(), "http://localhost:8081".into());

    env.insert(
        "GATE_DATABASE_URL".into(),
        "postgresql://postgres:development@postgres-auth:5432/clikd_auth".into(),
    );
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());

    env.insert(
        "GATE_JWT_SECRET".into(),
        "dev-jwt-secret-32-bytes-long-enough-for-testing-abc123".into(),
    );
    env.insert(
        "GATE_ENC_KEYS".into(),
        "wMGZCL5U/xmWwY9qyy2cu9PGJ1iokwGX4z16v9mhD8M=".into(),
    );
    env.insert(
        "GATE_COOKIE_SECRET".into(),
        "dev-cookie-secret-32-bytes-long-enough-for-testing-def456".into(),
    );
    env.insert(
        "GATE_INTERNAL_API_SECRET".into(),
        "dev-internal-api-secret-change-this".into(),
    );

    env.insert("GATE_NATS_URL".into(), "nats://nats:4222".into());
    env.insert("GATE_NATS_JETSTREAM_DOMAIN".into(), "default".into());
    env.insert(
        "GATE_OTEL_ENDPOINT".into(),
        "http://otel-collector:4317".into(),
    );

    env.insert("GATE_PROXY_MODE".into(), "false".into());
    env.insert("GATE_TRUSTED_PROXIES".into(), "".into());

    env.insert("APP_ENV".into(), config.dev.app_env.clone());

    env.insert("RIG_INTERNAL_URL".into(), "http://rig:8082".into());
    env.insert(
        "BACKEND_API_KEY".into(),
        "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into(),
    );

    env.insert("RUST_LOG".into(), config.dev.rust_log.clone());

    ServiceDefinition {
        name: "gate".into(),
        image: config.images.gate.clone(),
        ports: vec![(8081, 8081), (9001, 9001)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "curl".into(),
                "-f".into(),
                "http://localhost:8081/health".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["postgres-auth".into(), "keydb".into()],
        entrypoint: None,
        command: None,
        platform: None,
    }
}

fn rig_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();

    env.insert("RIG_PORTS_GRAPHQL".into(), "8082".into());
    env.insert("RIG_PORTS_GRPC".into(), "9002".into());

    env.insert("RIG_DATABASE_POOL_SIZE".into(), "20".into());
    env.insert("RIG_DATABASE_CONNECTION_TIMEOUT_SECS".into(), "30".into());

    env.insert("RIG_GRAPHQL_COMPLEXITY_LIMIT".into(), "1000".into());
    env.insert("RIG_GRAPHQL_DEPTH_LIMIT".into(), "10".into());

    env.insert("RIG_GRPC_MAX_MESSAGE_SIZE".into(), "4194304".into());
    env.insert("RIG_GRPC_KEEPALIVE_INTERVAL_SECS".into(), "60".into());

    env.insert("RIG_OBSERVABILITY_METRICS_ENABLED".into(), "true".into());
    env.insert("RIG_OBSERVABILITY_TRACING_ENABLED".into(), "true".into());

    env.insert(
        "RIG_DATABASE_URL".into(),
        "postgresql://postgres:development@postgres-rig:5432/clikd_rig".into(),
    );
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());
    env.insert("SCYLLADB_HOSTS".into(), "scylladb:9042".into());
    env.insert("NATS_URL".into(), "nats://nats:4222".into());
    env.insert("MINIO_ENDPOINT".into(), "http://minio:9000".into());
    env.insert("MINIO_ROOT_USER".into(), "minioadmin".into());
    env.insert("MINIO_ROOT_PASSWORD".into(), "minioadmin".into());
    env.insert(
        "OTEL_EXPORTER_OTLP_ENDPOINT".into(),
        "http://otel-collector:4317".into(),
    );

    env.insert(
        "JWT_SECRET".into(),
        "dev-jwt-secret-32-bytes-long-enough-for-testing-abc123".into(),
    );
    env.insert(
        "BACKEND_API_KEY".into(),
        "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into(),
    );

    env.insert("APP_ENV".into(), config.dev.app_env.clone());
    env.insert("RUST_LOG".into(), config.dev.rust_log.clone());

    ServiceDefinition {
        name: "rig".into(),
        image: config.images.rig.clone(),
        ports: vec![(8082, 8082), (9002, 9002)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "curl".into(),
                "-f".into(),
                "http://localhost:8082/health".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec![
            "postgres-rig".into(),
            "keydb".into(),
            "scylladb".into(),
            "nats".into(),
            "minio".into(),
        ],
        entrypoint: None,
        command: None,
        platform: None,
    }
}

fn studio_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("PORT".into(), "3001".into());
    env.insert("NODE_ENV".into(), "development".into());
    env.insert("APP_ENV".into(), config.dev.app_env.clone());
    env.insert("CLIKD_URL".into(), "http://apisix:9080".into());
    env.insert(
        "CLIKD_KEY".into(),
        "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into(),
    );
    env.insert(
        "NEXT_PUBLIC_STUDIO_URL".into(),
        "http://localhost:3001".into(),
    );
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
                "curl".into(),
                "-f".into(),
                "http://localhost:3001/api/health".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: Some(Duration::from_secs(15)),
        }),
        depends_on: vec!["apisix".into()],
        entrypoint: None,
        command: None,
        platform: None,
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
        entrypoint: None,
        command: None,
        platform: None,
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
        entrypoint: None,
        command: None,
        platform: None,
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
        entrypoint: None,
        command: Some(vec![
            "keydb-server".into(),
            "--protected-mode".into(),
            "no".into(),
            "--appendonly".into(),
            "yes".into(),
            "--server-threads".into(),
            "4".into(),
        ]),
        platform: None,
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
            test: vec![
                "CMD".into(),
                "cqlsh".into(),
                "-e".into(),
                "describe keyspaces".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 10,
            start_period: Some(Duration::from_secs(60)),
        }),
        depends_on: vec![],
        entrypoint: None,
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
        platform: None,
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
            test: vec![
                "CMD".into(),
                "wget".into(),
                "--spider".into(),
                "-q".into(),
                "http://localhost:9000/minio/health/live".into(),
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(20),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec![],
        entrypoint: None,
        command: Some(vec![
            "server".into(),
            "/data".into(),
            "--console-address".into(),
            ":9001".into(),
        ]),
        platform: None,
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
        entrypoint: None,
        command: Some(vec![
            "-js".into(),
            "-m".into(),
            "8222".into(),
            "-p".into(),
            "4222".into(),
            "--store_dir=/data".into(),
        ]),
        platform: None,
    }
}

fn render_apisix_routes() -> String {
    let mut env = Environment::new();
    env.add_template("routes", APISIX_ROUTES_TEMPLATE).unwrap();

    let template = env.get_template("routes").unwrap();
    template
        .render(minijinja::context! {
            gate_host => "gate",
            gate_port => 8081,
            rig_host => "rig",
            rig_port => 8082,
        })
        .unwrap()
}

fn apisix_service(_branch: &str, config: &Config) -> ServiceDefinition {
    let routes_config = render_apisix_routes();
    let entrypoint_script = format!(
        "cat <<'EOF' > /usr/local/apisix/conf/apisix.yaml && /docker-entrypoint.sh docker-start\n{}\nEOF",
        routes_config
    );

    ServiceDefinition {
        name: "apisix".into(),
        image: config.images.apisix.clone(),
        ports: vec![(9080, 9080)],
        env: HashMap::new(),
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "curl".into(),
                "-f".into(),
                "http://localhost:9080/".into(),
            ],
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
            retries: 3,
            start_period: Some(Duration::from_secs(10)),
        }),
        depends_on: vec!["gate".into(), "rig".into()],
        entrypoint: Some(vec!["sh".into(), "-c".into(), entrypoint_script]),
        command: None,
        platform: None,
    }
}
