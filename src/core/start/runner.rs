use crate::config::Config;
use crate::core::docker::{health, manager::DockerManager, network, services};
use crate::core::git::branch;
use crate::error::Result;
use crate::utils::theme::*;
use std::collections::{HashMap, HashSet};
use std::time::Duration;

pub async fn run(config: &Config, exclude: Vec<String>, ignore_health_check: bool) -> Result<()> {
    println!("{}", header("Starting Clikd"));

    let docker = DockerManager::new()?;
    let network_name = format!("clikd_network_{}", &config.project_id);

    network::create_network(docker.client(), &network_name).await?;

    let all_services = services::all_services("", config);
    let exclude_set: HashSet<String> = exclude.into_iter().collect();

    let services_to_start: Vec<_> = all_services
        .into_iter()
        .filter(|s| !exclude_set.contains(&s.name))
        .collect();

    if services_to_start.is_empty() {
        println!("\n{}", warning_message("No services to start"));
        return Ok(());
    }

    let ordered_services = resolve_dependencies(&services_to_start)?;

    println!("\n{}", step_message("Pulling Docker images..."));
    for service in &ordered_services {
        docker
            .pull_image_if_not_cached(&service.image, service.platform.as_deref())
            .await?;
    }

    println!("\n{}", step_message("Starting containers..."));
    let mut started_containers: Vec<String> = Vec::new();

    for service in &ordered_services {
        let container_name = docker
            .create_and_start_container(service, &network_name, &config.project_id)
            .await?;

        started_containers.push(container_name.clone());
    }

    if !ignore_health_check {
        let containers_with_health: Vec<String> = ordered_services
            .iter()
            .zip(started_containers.iter())
            .filter(|(svc, _)| svc.health_check.is_some())
            .map(|(_, name)| name.clone())
            .collect();

        if !containers_with_health.is_empty() {
            let mut sp = create_spinner("Waiting for health checks...");

            for container_name in &containers_with_health {
                match health::wait_healthy(
                    docker.client(),
                    container_name,
                    Duration::from_secs(120),
                )
                .await
                {
                    Ok(_) => {}
                    Err(e) => {
                        let service_name = container_name
                            .strip_prefix("clikd_")
                            .and_then(|s| s.rsplit_once('_'))
                            .map(|(name, _)| name)
                            .unwrap_or(container_name);
                        sp.fail(&format!("Container '{}' became unhealthy", service_name));
                        return Err(e);
                    }
                }
            }

            sp.success("All containers healthy!");
        }
    }

    branch::init_current_branch()?;

    println!(
        "\n{}\n",
        success_message(&format!(
            "Started {} local development setup",
            highlight("clikd")
        ))
    );

    let service_map: HashMap<String, &services::ServiceDefinition> = ordered_services
        .iter()
        .map(|s| (s.name.clone(), s))
        .collect();

    println!(
        "    {}: {}",
        highlight("Rig API URL"),
        url("http://127.0.0.1:9080/graphql")
    );
    println!(
        "   {}: {}",
        highlight("Gate Auth URL"),
        url("http://127.0.0.1:9080/auth")
    );
    if let Some(studio) = service_map.get("studio") {
        if !studio.ports.is_empty() {
            let (port, _) = studio.ports[0];
            println!(
                "      {}: {}",
                highlight("Studio URL"),
                url(&format!("http://127.0.0.1:{}", port))
            );
        }
    }
    if let Some(postgres_auth) = service_map.get("postgres-auth") {
        if !postgres_auth.ports.is_empty() {
            let (port, _) = postgres_auth.ports[0];
            let user = postgres_auth
                .env
                .get("POSTGRES_USER")
                .map(|s| s.as_str())
                .unwrap_or("postgres");
            let pass = postgres_auth
                .env
                .get("POSTGRES_PASSWORD")
                .map(|s| s.as_str())
                .unwrap_or("development");
            let db = postgres_auth
                .env
                .get("POSTGRES_DB")
                .map(|s| s.as_str())
                .unwrap_or("clikd_auth");
            println!(
                "  {}: {}",
                highlight("Database (Auth)"),
                dimmed(&format!(
                    "postgresql://{}:{}@127.0.0.1:{}/{}",
                    user, pass, port, db
                ))
            );
        }
    }
    if let Some(postgres_rig) = service_map.get("postgres-rig") {
        if !postgres_rig.ports.is_empty() {
            let (port, _) = postgres_rig.ports[0];
            let user = postgres_rig
                .env
                .get("POSTGRES_USER")
                .map(|s| s.as_str())
                .unwrap_or("postgres");
            let pass = postgres_rig
                .env
                .get("POSTGRES_PASSWORD")
                .map(|s| s.as_str())
                .unwrap_or("development");
            let db = postgres_rig
                .env
                .get("POSTGRES_DB")
                .map(|s| s.as_str())
                .unwrap_or("clikd_rig");
            println!(
                "   {}: {}",
                highlight("Database (Rig)"),
                dimmed(&format!(
                    "postgresql://{}:{}@127.0.0.1:{}/{}",
                    user, pass, port, db
                ))
            );
        }
    }
    if let Some(scylla) = service_map.get("scylladb") {
        if !scylla.ports.is_empty() {
            let (port, _) = scylla.ports[0];
            println!(
                "       {}: {}",
                highlight("ScyllaDB"),
                dimmed(&format!("127.0.0.1:{}", port))
            );
        }
    }
    if let Some(gate) = service_map.get("gate") {
        if let Some(backend_key) = gate.env.get("BACKEND_API_KEY") {
            println!("      {}: {}", highlight("Studio Key"), code(backend_key));
        }
    }

    println!();

    Ok(())
}

fn resolve_dependencies(
    services: &[services::ServiceDefinition],
) -> Result<Vec<services::ServiceDefinition>> {
    let mut ordered = Vec::new();
    let mut visited = HashSet::new();
    let mut visiting = HashSet::new();

    let service_map: HashMap<_, _> = services.iter().map(|s| (s.name.clone(), s)).collect();

    fn visit(
        name: &str,
        service_map: &HashMap<String, &services::ServiceDefinition>,
        visited: &mut HashSet<String>,
        visiting: &mut HashSet<String>,
        ordered: &mut Vec<services::ServiceDefinition>,
    ) -> Result<()> {
        if visited.contains(name) {
            return Ok(());
        }

        if visiting.contains(name) {
            return Ok(());
        }

        if let Some(service) = service_map.get(name) {
            visiting.insert(name.to_string());

            for dep in &service.depends_on {
                visit(dep, service_map, visited, visiting, ordered)?;
            }

            visiting.remove(name);
            visited.insert(name.to_string());
            ordered.push((*service).clone());
        }

        Ok(())
    }

    for service in services {
        visit(
            &service.name,
            &service_map,
            &mut visited,
            &mut visiting,
            &mut ordered,
        )?;
    }

    Ok(ordered)
}
