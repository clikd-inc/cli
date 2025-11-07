use crate::error::{CliError, Result};
use bollard::models::HealthStatusEnum;
use bollard::query_parameters::InspectContainerOptionsBuilder;
use bollard::Docker;
use std::time::Duration;
use tokio::time::{sleep, timeout};
use tracing::{debug, info, warn};

pub async fn wait_healthy(
    docker: &Docker,
    container_name: &str,
    timeout_duration: Duration,
) -> Result<()> {
    info!(
        "Waiting for container '{}' to become healthy",
        container_name
    );

    let result = timeout(timeout_duration, async {
        loop {
            let options = InspectContainerOptionsBuilder::default()
                .size(false)
                .build();

            let inspect = docker
                .inspect_container(container_name, Some(options))
                .await
                .map_err(CliError::Docker)?;

            if let Some(state) = inspect.state {
                if let Some(health) = state.health {
                    if let Some(status) = health.status {
                        match status {
                            HealthStatusEnum::HEALTHY => {
                                info!("Container '{}' is healthy", container_name);
                                return Ok::<(), CliError>(());
                            }
                            HealthStatusEnum::UNHEALTHY => {
                                warn!("Container '{}' is unhealthy", container_name);
                                return Err(CliError::Docker(
                                    bollard::errors::Error::DockerResponseServerError {
                                        status_code: 500,
                                        message: format!(
                                            "Container '{}' became unhealthy",
                                            container_name
                                        ),
                                    },
                                ));
                            }
                            _ => {
                                debug!(
                                    "Container '{}' health status: {:?}",
                                    container_name, status
                                );
                            }
                        }
                    }
                } else if let Some(running) = state.running {
                    if !running {
                        return Err(CliError::Docker(
                            bollard::errors::Error::DockerResponseServerError {
                                status_code: 500,
                                message: format!("Container '{}' is not running", container_name),
                            },
                        ));
                    } else {
                        info!(
                            "Container '{}' has no health check, assuming healthy",
                            container_name
                        );
                        return Ok(());
                    }
                }
            }

            sleep(Duration::from_secs(2)).await;
        }
    })
    .await;

    match result {
        Ok(inner_result) => inner_result,
        Err(_) => Err(CliError::Docker(
            bollard::errors::Error::DockerResponseServerError {
                status_code: 500,
                message: format!(
                    "Timeout waiting for container '{}' to become healthy",
                    container_name
                ),
            },
        )),
    }
}
