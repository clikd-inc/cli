use crate::error::{CliError, Result};
use bollard::Docker;
use bollard::network::{CreateNetworkOptions, InspectNetworkOptions};
use tracing::{debug, info};

pub async fn create_network(docker: &Docker, name: &str) -> Result<()> {
    debug!("Creating network: {}", name);

    match docker.inspect_network(name, Some(InspectNetworkOptions { verbose: false, scope: String::new() })).await {
        Ok(_) => {
            info!("Network '{}' already exists", name);
            return Ok(());
        }
        Err(bollard::errors::Error::DockerResponseServerError { status_code: 404, .. }) => {
        }
        Err(e) => return Err(CliError::Docker(e)),
    }

    let options = CreateNetworkOptions {
        name,
        check_duplicate: true,
        driver: "bridge",
        ..Default::default()
    };

    docker.create_network(options)
        .await
        .map_err(|e| CliError::Docker(e))?;

    info!("Created network '{}'", name);
    Ok(())
}
