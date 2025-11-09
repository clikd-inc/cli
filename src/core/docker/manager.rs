use crate::core::docker::registry;
use crate::core::docker::services::ServiceDefinition;
use crate::error::{CliError, Result};
use crate::utils::theme::*;
use bollard::models::{
    ContainerCreateBody, EndpointSettings, HealthConfig, HostConfig, NetworkingConfig, PortBinding,
    RestartPolicy, RestartPolicyNameEnum, VolumeCreateOptions,
};
use bollard::query_parameters::{
    CreateContainerOptionsBuilder, CreateImageOptionsBuilder, InspectContainerOptionsBuilder,
    ListContainersOptionsBuilder, PruneContainersOptionsBuilder, PruneNetworksOptionsBuilder,
    PruneVolumesOptionsBuilder, RemoveContainerOptionsBuilder, StartContainerOptionsBuilder,
    StopContainerOptionsBuilder,
};
use bollard::Docker;
use futures::StreamExt;
use std::collections::HashMap;
use tracing::{debug, info};

#[derive(Clone)]
pub struct DockerManager {
    client: Docker,
}

impl DockerManager {
    pub fn new() -> Result<Self> {
        let client = Docker::connect_with_local_defaults().map_err(CliError::Docker)?;

        Ok(Self { client })
    }

    pub fn client(&self) -> &Docker {
        &self.client
    }

    pub async fn is_docker_running(&self) -> bool {
        let ping_future = self.client.ping();
        let timeout = tokio::time::timeout(std::time::Duration::from_secs(2), ping_future);

        match timeout.await {
            Ok(Ok(_)) => true,
            Ok(Err(e)) => {
                debug!("Docker ping failed: {}", e);
                false
            }
            Err(_) => {
                debug!("Docker ping timed out");
                false
            }
        }
    }

    pub async fn pull_image(&self, image: &str, platform: Option<&str>) -> Result<()> {
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

        let mut stream = self.client.create_image(Some(options), None, credentials);

        let mut shown_layers: StdHashMap<String, bool> = StdHashMap::new();
        let mut pb = create_progress_bar();
        let mut last_status = String::new();

        while let Some(result) = stream.next().await {
            match result {
                Ok(info) => {
                    if let Some(error) = info.error {
                        pb.finish_and_clear();
                        return Err(CliError::Docker(
                            bollard::errors::Error::DockerResponseServerError {
                                status_code: 500,
                                message: error,
                            },
                        ));
                    }

                    if let Some(id) = info.id {
                        if let Some(status) = info.status {
                            let status_lower = status.to_lowercase();

                            if status_lower.contains("pulling fs layer") {
                                if !shown_layers.contains_key(&id) {
                                    shown_layers.insert(id.clone(), false);
                                }
                            } else if status_lower.contains("downloading") {
                                if let (Some(current), Some(total)) = (
                                    info.progress_detail.as_ref().and_then(|p| p.current),
                                    info.progress_detail.as_ref().and_then(|p| p.total),
                                ) {
                                    pb.set_length(total as u64);
                                    pb.set_position(current as u64);
                                    pb.set_message(format!("Downloading {}", highlight(&id)));
                                }
                            } else if status_lower.contains("extracting") {
                                if let (Some(current), Some(total)) = (
                                    info.progress_detail.as_ref().and_then(|p| p.current),
                                    info.progress_detail.as_ref().and_then(|p| p.total),
                                ) {
                                    pb.set_length(total as u64);
                                    pb.set_position(current as u64);
                                    pb.set_message(format!("Extracting {}", highlight(&id)));
                                }
                            } else if status_lower.contains("already exists") {
                                if shown_layers.get(&id) == Some(&false) {
                                    shown_layers.insert(id.clone(), true);
                                }
                            } else if status_lower.contains("pull complete")
                                && shown_layers.get(&id) == Some(&false)
                            {
                                shown_layers.insert(id.clone(), true);
                            }
                        }
                    } else if let Some(status) = info.status {
                        if status.starts_with("Status:") {
                            last_status = status;
                        }
                    }
                }
                Err(e) => {
                    pb.finish_and_clear();
                    return Err(CliError::Docker(e));
                }
            }
        }

        pb.finish_and_clear();
        if !last_status.is_empty() {
            println!("{} {}", success_icon(), format_docker_status(&last_status));
        }
        Ok(())
    }

    pub async fn pull_image_if_not_cached(
        &self,
        image: &str,
        platform: Option<&str>,
    ) -> Result<()> {
        if self.image_exists(image).await? {
            return Ok(());
        }

        let parts: Vec<&str> = image.split(':').collect();
        let image_name = parts.first().unwrap_or(&image);
        let tag = parts.get(1).unwrap_or(&"latest");

        if let Some(plat) = platform {
            println!(
                "{}: {} {} ({})",
                highlight(tag),
                dimmed("Pulling from"),
                image_name,
                dimmed(plat)
            );
        } else {
            println!(
                "{}: {} {}",
                highlight(tag),
                dimmed("Pulling from"),
                image_name
            );
        }

        self.pull_image(image, platform).await
    }

    pub async fn image_exists(&self, image: &str) -> Result<bool> {
        debug!("Checking if image exists: {}", image);

        match self.client.inspect_image(image).await {
            Ok(_) => Ok(true),
            Err(bollard::errors::Error::DockerResponseServerError {
                status_code: 404, ..
            }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn container_exists(&self, name: &str) -> Result<bool> {
        debug!("Checking if container exists: {}", name);

        let options = InspectContainerOptionsBuilder::default()
            .size(false)
            .build();

        match self.client.inspect_container(name, Some(options)).await {
            Ok(_) => Ok(true),
            Err(bollard::errors::Error::DockerResponseServerError {
                status_code: 404, ..
            }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn container_running(&self, name: &str) -> Result<bool> {
        debug!("Checking if container is running: {}", name);

        let options = InspectContainerOptionsBuilder::default()
            .size(false)
            .build();

        match self.client.inspect_container(name, Some(options)).await {
            Ok(inspect) => {
                if let Some(state) = inspect.state {
                    Ok(state.running.unwrap_or(false))
                } else {
                    Ok(false)
                }
            }
            Err(bollard::errors::Error::DockerResponseServerError {
                status_code: 404, ..
            }) => Ok(false),
            Err(e) => Err(CliError::Docker(e)),
        }
    }

    pub async fn remove_container(&self, name: &str, force: bool) -> Result<()> {
        debug!("Removing container: {}", name);

        let options = RemoveContainerOptionsBuilder::default()
            .force(force)
            .build();

        self.client
            .remove_container(name, Some(options))
            .await
            .map_err(CliError::Docker)?;

        info!("Removed container '{}'", name);
        Ok(())
    }

    pub async fn create_and_start_container(
        &self,
        service: &ServiceDefinition,
        network_name: &str,
        project_id: &str,
    ) -> Result<String> {
        let container_name = format!("clikd_{}_{}", service.name, project_id);

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

        let env: Vec<String> = service
            .env
            .iter()
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
            restart_policy: Some(RestartPolicy {
                name: Some(RestartPolicyNameEnum::ALWAYS),
                maximum_retry_count: None,
            }),
            ..Default::default()
        });

        let endpoint = EndpointSettings {
            aliases: Some(vec![service.name.clone()]),
            ..Default::default()
        };

        let mut endpoints_config: HashMap<String, EndpointSettings> = HashMap::new();
        endpoints_config.insert(network_name.to_string(), endpoint);

        let networking_config = NetworkingConfig {
            endpoints_config: Some(endpoints_config),
        };

        let mut labels = HashMap::new();
        labels.insert(
            "com.docker.compose.project".to_string(),
            project_id.to_string(),
        );
        labels.insert("com.clikd.cli.project".to_string(), project_id.to_string());

        for volume_bind in &service.volumes {
            if let Some(colon_pos) = volume_bind.find(':') {
                let source = &volume_bind[..colon_pos];
                if !source.starts_with('/') && !source.starts_with('.') {
                    let volume_config = VolumeCreateOptions {
                        name: Some(source.to_string()),
                        labels: Some(labels.clone()),
                        ..Default::default()
                    };

                    match self.client.create_volume(volume_config).await {
                        Ok(_) => info!("Created volume '{}' with labels", source),
                        Err(e) => {
                            if e.to_string().contains("already exists") {
                                debug!("Volume '{}' already exists", source);
                            } else {
                                return Err(CliError::Docker(e));
                            }
                        }
                    }
                }
            }
        }

        let config = ContainerCreateBody {
            image: Some(service.image.clone()),
            env: Some(env),
            exposed_ports: Some(exposed_ports),
            host_config,
            healthcheck: health_config,
            entrypoint: service.entrypoint.clone(),
            cmd: service.command.clone(),
            networking_config: Some(networking_config),
            labels: Some(labels),
            ..Default::default()
        };

        info!("Creating container '{}'", container_name);

        let create_options = CreateContainerOptionsBuilder::default()
            .name(&container_name)
            .platform(service.platform.as_deref().unwrap_or(""))
            .build();

        self.client
            .create_container(Some(create_options), config)
            .await
            .map_err(CliError::Docker)?;

        info!("Starting container '{}'", container_name);

        let start_options = StartContainerOptionsBuilder::default().build();

        self.client
            .start_container(&container_name, Some(start_options))
            .await
            .map_err(CliError::Docker)?;

        Ok(container_name)
    }

    pub async fn stop_all_containers(&self, project_id: &str, keep_volumes: bool) -> Result<()> {
        use bollard::models::ContainerSummaryStateEnum;

        info!("Stopping all containers for project: {}", project_id);

        let label_filter = format!("com.clikd.cli.project={}", project_id);
        let filters = HashMap::from([("label".to_string(), vec![label_filter.clone()])]);

        let list_options = ListContainersOptionsBuilder::default()
            .all(true)
            .filters(&filters)
            .build();

        let containers = self
            .client
            .list_containers(Some(list_options))
            .await
            .map_err(CliError::Docker)?;

        let mut running_ids = Vec::new();
        for container in &containers {
            if let Some(state) = &container.state {
                if *state == ContainerSummaryStateEnum::RUNNING {
                    if let Some(id) = &container.id {
                        running_ids.push(id.clone());
                    }
                }
            }
        }

        for id in running_ids {
            info!("Stopping container: {}", id);
            let stop_options = StopContainerOptionsBuilder::default().build();
            self.client
                .stop_container(&id, Some(stop_options))
                .await
                .map_err(CliError::Docker)?;
        }

        let prune_filters = HashMap::from([("label".to_string(), vec![label_filter.clone()])]);

        let prune_options = PruneContainersOptionsBuilder::default()
            .filters(&prune_filters)
            .build();

        let report = self
            .client
            .prune_containers(Some(prune_options))
            .await
            .map_err(CliError::Docker)?;

        info!("Pruned containers: {:?}", report.containers_deleted);

        if !keep_volumes {
            let volume_filters = HashMap::from([("label".to_string(), vec![label_filter.clone()])]);

            let volume_prune_options = PruneVolumesOptionsBuilder::default()
                .filters(&volume_filters)
                .build();

            let volume_report = self
                .client
                .prune_volumes(Some(volume_prune_options))
                .await
                .map_err(CliError::Docker)?;

            info!("Pruned volumes: {:?}", volume_report.volumes_deleted);
        }

        let network_filters = HashMap::from([("label".to_string(), vec![label_filter])]);

        let network_prune_options = PruneNetworksOptionsBuilder::default()
            .filters(&network_filters)
            .build();

        let network_report = self
            .client
            .prune_networks(Some(network_prune_options))
            .await
            .map_err(CliError::Docker)?;

        info!("Pruned networks: {:?}", network_report.networks_deleted);

        Ok(())
    }
}
