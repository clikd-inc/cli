pub mod app_data;
pub mod app_error;
pub mod config;
pub mod docker_data;
pub mod exec;
pub mod input_handler;
pub mod ui;

pub use app_data::AppData;
pub use app_error::AppError;
pub use docker_data::DockerData;
pub use input_handler::InputHandler;
pub use ui::{GuiState, Rerender, Ui};

pub const ENTRY_POINT: &str = "/app/clikd";
pub const ENV_KEY: &str = "CLIKD_ENV";
pub const ENV_VALUE: &str = "container";

#[cfg(test)]
pub mod tests {
    use crate::core::status::{
        app_data::{
            AppData, ContainerId, ContainerItem, ContainerPorts, ContainerStatus, Filter,
            RunningState, State, StatefulList,
        },
        config::{AppColors, Config, Keymap},
        ui::Rerender,
    };
    use bollard::service::{ContainerSummary, Port};
    use std::{str::FromStr, sync::Arc};

    pub fn gen_config() -> Config {
        Config {
            color_logs: false,
            docker_interval_ms: 1000,
            gui: true,
            host: None,
            show_std_err: false,
            in_container: false,
            save_dir: None,
            log_search_case_sensitive: true,
            raw_logs: false,
            show_self: false,
            app_colors: AppColors::new(),
            keymap: Keymap::new(),
            timestamp_format: "HH:MM:SS.NNNNN dd-mm-yyyy".to_owned(),
            show_timestamp: false,
            use_cli: false,
            show_logs: true,
            timezone: None,
        }
    }

    pub fn gen_item(id: &ContainerId, index: usize) -> ContainerItem {
        ContainerItem::new(
            u64::try_from(index).expect("BUG: test index should fit in u64"),
            id.clone(),
            format!("image_{index}"),
            false,
            format!("container_{index}"),
            vec![ContainerPorts {
                ip: None,
                private: u16::try_from(index).unwrap_or(1) + 8000,
                public: None,
            }],
            State::Running(RunningState::Healthy),
            ContainerStatus::from(format!("Up {index} hour")),
        )
    }

    pub fn gen_appdata(containers: &[ContainerItem]) -> AppData {
        AppData {
            containers: StatefulList::new(containers.to_vec()),
            hidden_containers: vec![],
            current_sorted_id: vec![],
            error: None,
            sorted_by: None,
            rerender: Arc::new(Rerender::new()),
            filter: Filter::new(),
            config: gen_config(),
        }
    }

    pub fn gen_containers() -> (Vec<ContainerId>, Vec<ContainerItem>) {
        let ids = (1..=3)
            .map(|i| ContainerId::from(format!("{i}").as_str()))
            .collect::<Vec<_>>();
        let containers = ids
            .iter()
            .enumerate()
            .map(|(index, id)| gen_item(id, index + 1))
            .collect::<Vec<_>>();
        (ids, containers)
    }

    pub fn gen_container_summary(index: usize, state: &str) -> ContainerSummary {
        ContainerSummary {
            image_manifest_descriptor: None,
            id: Some(format!("{index}")),
            names: Some(vec![format!("container_{}", index)]),
            image: Some(format!("image_{index}")),
            image_id: Some(format!("{index}")),
            command: None,
            created: Some(i64::try_from(index).expect("BUG: test index should fit in i64")),
            ports: Some(vec![Port {
                ip: None,
                private_port: u16::try_from(index).unwrap_or(1) + 8000,
                public_port: None,
                typ: None,
            }]),
            size_rw: None,
            size_root_fs: None,
            labels: None,
            state: Some(
                bollard::secret::ContainerSummaryStateEnum::from_str(state)
                    .expect("BUG: test state string should be valid"),
            ),
            status: Some(format!("Up {index} hour")),
            host_config: None,
            network_settings: None,
            mounts: None,
        }
    }
}
