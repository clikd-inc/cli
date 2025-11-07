use crate::error::{CliError, Result};
use bollard::models::NetworkCreateRequest;
use bollard::query_parameters::InspectNetworkOptionsBuilder;
use bollard::Docker;
use tracing::{debug, info};

pub async fn create_network(docker: &Docker, name: &str) -> Result<()> {
    debug!("Creating network: {}", name);

    let inspect_options = InspectNetworkOptionsBuilder::default()
        .verbose(false)
        .scope("")
        .build();

    match docker.inspect_network(name, Some(inspect_options)).await {
        Ok(_) => {
            info!("Network '{}' already exists", name);
            return Ok(());
        }
        Err(bollard::errors::Error::DockerResponseServerError {
            status_code: 404, ..
        }) => {}
        Err(e) => return Err(CliError::Docker(e)),
    }

    let options = NetworkCreateRequest {
        name: name.to_string(),
        driver: Some("bridge".to_string()),
        ..Default::default()
    };

    docker
        .create_network(options)
        .await
        .map_err(CliError::Docker)?;

    info!("Created network '{}'", name);
    Ok(())
}
