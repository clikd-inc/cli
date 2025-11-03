use std::fs;
use std::path::Path;
use crate::error::Result;

const GITIGNORE_ENTRIES: &str = r#"
# Clikd
clikd/.temp
clikd/.branches
"#;

pub fn update(project_root: &Path) -> Result<()> {
    if !project_root.join(".git").exists() {
        return Ok(());
    }

    let gitignore_path = project_root.join(".gitignore");

    let mut content = if gitignore_path.exists() {
        fs::read_to_string(&gitignore_path)?
    } else {
        String::new()
    };

    if content.contains("# Clikd") {
        return Ok(());
    }

    if !content.is_empty() && !content.ends_with('\n') {
        content.push('\n');
    }
    content.push_str(GITIGNORE_ENTRIES);

    fs::write(&gitignore_path, content)?;
    Ok(())
}

pub fn is_git_repo(project_root: &Path) -> bool {
    project_root.join(".git").exists()
}
