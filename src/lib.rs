pub mod config;
pub mod git;

pub use config::ClikdConfig;
pub use git::{GitInfo, detect_git_info, get_branch_name, sanitize_branch_name};