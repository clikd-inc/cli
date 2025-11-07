use crate::error::Result;
use std::fs;
use std::path::Path;

const VSCODE_SETTINGS: &str = include_str!("../../../templates/vscode-settings.json");
const VSCODE_EXTENSIONS: &str = include_str!("../../../templates/vscode-extensions.json");

pub fn create_settings(project_root: &Path) -> Result<()> {
    let vscode_dir = project_root.join(".vscode");
    fs::create_dir_all(&vscode_dir)?;

    let settings_path = vscode_dir.join("settings.json");
    fs::write(settings_path, VSCODE_SETTINGS)?;

    let extensions_path = vscode_dir.join("extensions.json");
    fs::write(extensions_path, VSCODE_EXTENSIONS)?;

    Ok(())
}

pub fn settings_exist(project_root: &Path) -> bool {
    project_root.join(".vscode/settings.json").exists()
}
