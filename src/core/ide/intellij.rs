use crate::error::Result;
use std::fs;
use std::path::Path;

const INTELLIJ_SETTINGS: &str = include_str!("../../../templates/intellij-clikd.xml");

pub fn create_settings(project_root: &Path) -> Result<()> {
    let idea_dir = project_root.join(".idea");
    fs::create_dir_all(&idea_dir)?;

    let settings_path = idea_dir.join("clikd.xml");
    fs::write(settings_path, INTELLIJ_SETTINGS)?;

    Ok(())
}

pub fn settings_exist(project_root: &Path) -> bool {
    project_root.join(".idea/clikd.xml").exists()
}
