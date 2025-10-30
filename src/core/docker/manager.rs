use crate::error::{CliError, Result};
use crate::core::docker::registry;
use crate::core::docker::services::ServiceDefinition;
use bollard::Docker;
use bollard::query_parameters::CreateImageOptionsBuilder;
use bollard::container::{Config, CreateContainerOptions, StartContainerOptions};
use bollard::models::{HostConfig, PortBinding, HealthConfig};
use futures::StreamExt;
use std::collections::HashMap;
use tracing::{debug, info};
use owo_colors::OwoColorize;

pub struct DockerManager {
    client: Docker,
}

impl DockerManager {
    pub fn new() -> Result<Self> {
        let client = Docker::connect_with_local_defaults()
            .map_err(|e| CliError::Docker(e))?;

        Ok(Self { client })
    }

    pub fn client(&self) -> &Docker {
        &self.client
    }

    pub async fn pull_image(&self, image: &str, platform: Option<&str>) -> Result<()> {
        use indicatif::{ProgressBar, ProgressStyle};
        use std::collections::HashMap as StdHashMap;

        let credentials = if registry::is_ghcr_image(image) {
            Some(registry::get_ghcr_credentials().await?)
        } else {
            None
        };

        let mut options_builder = CreateImageOptionsBuilder::default().from_image(image);
        if let Some(plat) = platform {
            options_builder = options_builder.platform(plat);
        }
        let options = options_builder.build();

        let mut stream = self.client.create_image(
            Some(options),
            None,
            credentials,
        );

        let mut extracting_spinners: StdHashMap<String, ProgressBar> = StdHashMap::new();
        let mut downloading_lines: StdHashMap<String, String> = StdHashMap::new();

        while let Some(result) = stream.next().await {
            match result {
                Ok(info) => {
                    if let Some(error) = info.error {
                        for spinner in extracting_spinners.values() {
                            spinner.finish_and_clear();
                        }
                        return Err(CliError::Docker(
                            bollard::errors::Error::DockerResponseServerError {
                                status_code: 500,
                                message: error,
                            }
                        ));
                    }

                    if let Some(id) = info.id {
                        if let Some(status) = info.status {
                            let status_lower = status.to_lowercase();

                            if status_lower.contains("already exists") {
                                println!("{}: {}", id.dimmed(), "Already exists".dimmed());
                            } else if status_lower.contains("pull complete") {
                                if let Some(spinner) = extracting_spinners.remove(&id) {
                                    spinner.finish_and_clear();
                                }
                                if let Some(line) = downloading_lines.remove(&id) {
                                    print!("\r{}\r", " ".repeat(line.len()));
                                }
                                println!("{}: {}", id.bright_cyan(), "Pull complete".green());
                            } else if status_lower.contains("downloading") {
                                if let (Some(current), Some(total)) = (info.progress_detail.as_ref().and_then(|p| p.current), info.progress_detail.as_ref().and_then(|p| p.total)) {
                                    let percent = if total > 0 { (current as f64 / total as f64 * 100.0) as u32 } else { 0 };
                                    let bar_width = 40;
                                    let filled = (bar_width * percent as usize) / 100;
                                    let bar = format!("[{}{}]", "=".repeat(filled), ">".repeat(bar_width.saturating_sub(filled)).chars().take(1).collect::<String>() + &" ".repeat(bar_width.saturating_sub(filled).saturating_sub(1)));

                                    let line = format!("{}: {} {} {:.1}MB/{:.1}MB",
                                        id.bright_cyan(),
                                        status,
                                        bar,
                                        current as f64 / 1_000_000.0,
                                        total as f64 / 1_000_000.0
                                    );
                                    print!("\r{}", line);
                                    downloading_lines.insert(id.clone(), line);
                                    use std::io::Write;
                                    std::io::stdout().flush().ok();
                                }
                            } else if status_lower.contains("extracting") {
                                if let Some(line) = downloading_lines.remove(&id) {
                                    print!("\r{}\r", " ".repeat(line.len()));
                                }

                                if !extracting_spinners.contains_key(&id) {
                                    let pb = ProgressBar::new_spinner();
                                    pb.set_style(
                                        ProgressStyle::default_spinner()
                                            .tick_chars("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
                                            .template(&format!("  {{spinner:.cyan}} {}: Extracting...", id.bright_cyan()))
                                            .unwrap(),
                                    );
                                    pb.enable_steady_tick(std::time::Duration::from_millis(100));
                                    extracting_spinners.insert(id.clone(), pb);
                                }
                            } else if status_lower.contains("waiting") {
                            } else if !status_lower.contains("pulling fs layer") && !status_lower.contains("verifying checksum") && !status_lower.contains("download complete") {
                                println!("{}: {}", id, status);
                            }
                        }
                    } else if let Some(status) = info.status {
                        if status.starts_with("Digest:") || status.starts_with("Status:") {
                            println!("\n{}", status.green());
                        } else if !status.contains("Pulling from") {
                            println!("{}", status);
                        }
                    }
                }
                Err(e) => {
                    for spinner in extracting_spinners.values() {
                        spinner.finish_and_clear();
                    }
                    return Err(CliError::Docker(e));
                }
            }
        }

        for spinner in extracting_spinners.values() {
            spinner.finish_and_clear();
        }

        Ok(())
    }

    pub async fn pull_image_if_not_cached(&self, image: &str, platform: Option<&str>) -> Result<()> {
        if self.image_exists(image).await? {
            return Ok(());
        }

        let parts: Vec<&str> = image.split(':').collect();
        let image_name = parts.get(0).unwrap_or(&image);
        let tag = parts.get(1).unwrap_or(&"latest");

        if let Some(plat) = platform {
            println!("{}: Pulling from {} ({})", tag.cyan(), image_name, plat.dimmed());
        } else {
            println!("{}: Pulling from {}", tag.cyan(), image_name);
        }

        self.pull_image(image, platform).await
    }

    pub async fn image_exists(&self, image: &str) -> Result<bool> {
        debug!("Checking if image exists: {}", image);

        match self.client.inspect_image(image).await {
            Ok(_) => Ok(true),
            Err(bollard::errors::Error::DockerResponseServerError { status_code: 404, .. }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn container_exists(&self, name: &str) -> Result<bool> {
        debug!("Checking if container exists: {}", name);

        use bollard::container::InspectContainerOptions;

        match self.client.inspect_container(name, Some(InspectContainerOptions { size: false })).await {
            Ok(_) => Ok(true),
            Err(bollard::errors::Error::DockerResponseServerError { status_code: 404, .. }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn container_running(&self, name: &str) -> Result<bool> {
        debug!("Checking if container is running: {}", name);

        use bollard::container::InspectContainerOptions;

        match self.client.inspect_container(name, Some(InspectContainerOptions { size: false })).await {
            Ok(inspect) => {
                if let Some(state) = inspect.state {
                    Ok(state.running.unwrap_or(false))
                } else {
                    Ok(false)
                }
            }
            Err(bollard::errors::Error::DockerResponseServerError { status_code: 404, .. }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn remove_container(&self, name: &str, force: bool) -> Result<()> {
        debug!("Removing container: {}", name);

        use bollard::container::RemoveContainerOptions;

        self.client
            .remove_container(name, Some(RemoveContainerOptions {
                force,
                ..Default::default()
            }))
            .await
            .map_err(|e| CliError::Docker(e))?;

        info!("Removed container '{}'", name);
        Ok(())
    }

    pub async fn create_and_start_container(
        &self,
        service: &ServiceDefinition,
        network_name: &str,
    ) -> Result<String> {
        let container_name = format!("clikd_{}", service.name);

        if self.container_exists(&container_name).await? {
            if self.container_running(&container_name).await? {
                info!("Container '{}' is already running", container_name);
                return Ok(container_name);
            } else {
                debug!("Removing existing stopped container '{}'", container_name);
                self.remove_container(&container_name, false).await?;
            }
        }

        let mut port_bindings: HashMap<String, Option<Vec<PortBinding>>> = HashMap::new();
        let mut exposed_ports: HashMap<String, HashMap<(), ()>> = HashMap::new();

        for (host_port, container_port) in &service.ports {
            let port_key = format!("{}/tcp", container_port);
            exposed_ports.insert(port_key.clone(), HashMap::new());
            port_bindings.insert(
                port_key,
                Some(vec![PortBinding {
                    host_ip: Some("0.0.0.0".to_string()),
                    host_port: Some(host_port.to_string()),
                }]),
            );
        }

        let env: Vec<String> = service.env.iter()
            .map(|(k, v)| format!("{}={}", k, v))
            .collect();

        let health_config = service.health_check.as_ref().map(|hc| HealthConfig {
            test: Some(hc.test.clone()),
            interval: Some(hc.interval.as_nanos() as i64),
            timeout: Some(hc.timeout.as_nanos() as i64),
            retries: Some(hc.retries as i64),
            start_period: hc.start_period.map(|sp| sp.as_nanos() as i64),
            ..Default::default()
        });

        let host_config = Some(HostConfig {
            port_bindings: Some(port_bindings),
            binds: if service.volumes.is_empty() {
                None
            } else {
                Some(service.volumes.clone())
            },
            network_mode: Some(network_name.to_string()),
            ..Default::default()
        });

        let config = Config {
            image: Some(service.image.clone()),
            env: Some(env),
            exposed_ports: Some(exposed_ports),
            host_config,
            healthcheck: health_config,
            cmd: service.command.clone(),
            ..Default::default()
        };

        info!("Creating container '{}'", container_name);

        self.client
            .create_container(
                Some(CreateContainerOptions {
                    name: &container_name,
                    platform: service.platform.as_ref(),
                }),
                config,
            )
            .await
            .map_err(|e| CliError::Docker(e))?;

        info!("Starting container '{}'", container_name);

        self.client
            .start_container(&container_name, None::<StartContainerOptions<String>>)
            .await
            .map_err(|e| CliError::Docker(e))?;

        Ok(container_name)
    }
}

