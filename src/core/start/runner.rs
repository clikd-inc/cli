use crate::error::Result;
use crate::config::Config;
use crate::core::docker::{manager::DockerManager, services, network, health};
use owo_colors::OwoColorize;
use std::collections::{HashMap, HashSet};
use std::time::Duration;

pub async fn run(config: &Config, exclude: Vec<String>, ignore_health_check: bool) -> Result<()> {
    println!("\n{} Clikd", "Starting".cyan().bold());
    println!("{}", "─".repeat(60).dimmed());

    let docker = DockerManager::new()?;
    let network_name = "clikd_network";

    network::create_network(docker.client(), network_name).await?;

    let all_services = services::all_services("", config);
    let exclude_set: HashSet<String> = exclude.into_iter().collect();

    let services_to_start: Vec<_> = all_services
        .into_iter()
        .filter(|s| !exclude_set.contains(&s.name))
        .collect();

    if services_to_start.is_empty() {
        println!("\n{} No services to start", "⚠".yellow());
        return Ok(());
    }

    let ordered_services = resolve_dependencies(&services_to_start)?;

    println!("\n{} Pulling Docker images...", "▶".green().bold());
    for service in &ordered_services {
        docker.pull_image_if_not_cached(&service.image, service.platform.as_deref()).await?;
    }

    println!("\n{} Starting containers...", "▶".green().bold());
    let mut started_containers: Vec<String> = Vec::new();

    use indicatif::{ProgressBar, ProgressStyle};

    for service in &ordered_services {
        let pb = ProgressBar::new_spinner();
        pb.set_style(
            ProgressStyle::default_spinner()
                .tick_chars("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
                .template("  {spinner:.cyan} Starting {msg}...")
                .unwrap(),
        );
        pb.enable_steady_tick(std::time::Duration::from_millis(100));
        pb.set_message(service.name.bright_blue().to_string());

        let container_name = docker
            .create_and_start_container(service, network_name)
            .await?;

        pb.finish_with_message(format!("{} {}", "✓".green(), service.name.bright_green()));
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
            println!("\n{} Waiting for health checks...", "⏳".yellow());

            for container_name in &containers_with_health {
                let service_name = container_name.strip_prefix("clikd_").unwrap_or(container_name);

                let pb = ProgressBar::new_spinner();
                pb.set_style(
                    ProgressStyle::default_spinner()
                        .tick_chars("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
                        .template("  {spinner:.yellow} Checking {msg}...")
                        .unwrap(),
                );
                pb.enable_steady_tick(std::time::Duration::from_millis(100));
                pb.set_message(service_name.bright_blue().to_string());

                match health::wait_healthy(docker.client(), container_name, Duration::from_secs(120)).await {
                    Ok(_) => pb.finish_with_message(format!("{} {}", "✓".green(), service_name.bright_green())),
                    Err(e) => {
                        pb.finish_with_message(format!("{} {}", "✗".red(), service_name.red()));
                        return Err(e);
                    }
                }
            }
        }
    }

    println!("\n{} All services started successfully!", "✓".green().bold());
    println!("\n{}", "Service URLs:".cyan().bold());
    for service in &ordered_services {
        if !service.ports.is_empty() {
            let (host_port, _) = service.ports[0];
            println!("  {} http://localhost:{}", service.name.dimmed(), host_port.to_string().bright_blue());
        }
    }

    Ok(())
}

fn resolve_dependencies(services: &[services::ServiceDefinition]) -> Result<Vec<services::ServiceDefinition>> {
    let mut ordered = Vec::new();
    let mut visited = HashSet::new();
    let mut visiting = HashSet::new();

    let service_map: HashMap<_, _> = services.iter()
        .map(|s| (s.name.clone(), s))
        .collect();

    fn visit<'a>(
        name: &str,
        service_map: &HashMap<String, &'a services::ServiceDefinition>,
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
        visit(&service.name, &service_map, &mut visited, &mut visiting, &mut ordered)?;
    }

    Ok(ordered)
}
