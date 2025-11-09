use crate::cli::StatusArgs;
use crate::config::Config;
use crate::core::docker::manager::DockerManager;
use crate::core::status::{
    config::{AppColors, Config as StatusConfig, Keymap},
    AppData, DockerData, GuiState, InputHandler, Rerender, Ui,
};
use crate::error::Result;
use bollard::Docker;
use parking_lot::Mutex;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;
use tokio::sync::mpsc;

pub async fn run(_args: StatusArgs, _config: Config) -> Result<()> {
    let docker_manager = match DockerManager::new() {
        Ok(docker) => docker,
        Err(e) => {
            if let Some(socket_path) = extract_docker_socket_error(&e) {
                return Err(crate::error::CliError::DockerNotRunning(socket_path).into());
            }
            return Err(e.into());
        }
    };

    if !docker_manager.is_docker_running().await {
        let socket = std::env::var("DOCKER_HOST")
            .unwrap_or_else(|_| "unix:///var/run/docker.sock".to_string());
        return Err(crate::error::CliError::DockerNotRunning(socket).into());
    }

    let docker = docker_manager.client().clone();
    run_tui(docker).await
}

fn extract_docker_socket_error(err: &crate::error::CliError) -> Option<String> {
    match err {
        crate::error::CliError::Docker(bollard::errors::Error::SocketNotFoundError(path)) => {
            Some(path.clone())
        }
        crate::error::CliError::Docker(bollard::errors::Error::IOError { .. }) => Some(
            std::env::var("DOCKER_HOST").unwrap_or_else(|_| "/var/run/docker.sock".to_string()),
        ),
        _ => None,
    }
}

fn create_status_config() -> StatusConfig {
    StatusConfig {
        app_colors: AppColors::new(),
        color_logs: true,
        docker_interval_ms: 1000,
        gui: true,
        host: None,
        in_container: false,
        keymap: Keymap::new(),
        log_search_case_sensitive: false,
        raw_logs: false,
        save_dir: None,
        show_logs: true,
        show_self: false,
        show_std_err: true,
        show_timestamp: true,
        timestamp_format: "%Y-%m-%d %H:%M:%S".to_string(),
        timezone: None,
        use_cli: false,
    }
}

async fn run_tui(docker: Docker) -> Result<()> {
    let status_config = create_status_config();
    let show_logs = status_config.show_logs;

    let rerender = Arc::new(Rerender::new());
    let app_data = Arc::new(Mutex::new(AppData::new(status_config, &rerender)));
    let gui_state = Arc::new(Mutex::new(GuiState::new(&rerender, show_logs)));
    let is_running = Arc::new(AtomicBool::new(true));

    let (docker_tx, docker_rx) = mpsc::channel(32);
    let (input_tx, input_rx) = mpsc::channel(32);

    tokio::spawn(DockerData::start(
        Arc::clone(&app_data),
        docker,
        docker_rx,
        docker_tx.clone(),
        Arc::clone(&gui_state),
    ));

    tokio::spawn(InputHandler::start(
        Arc::clone(&app_data),
        docker_tx,
        Arc::clone(&gui_state),
        Arc::clone(&is_running),
        input_rx,
    ));

    Ui::start(app_data, gui_state, input_tx, is_running, rerender).await;

    Ok(())
}
