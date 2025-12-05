//! Pull Request content generation for release PRs.
//!
//! Generates formatted PR titles and bodies for release pull requests,
//! including version tables, ecosystem badges, and changelog summaries.
//!
//! # Generated PR Format
//!
//! ## Title Examples
//!
//! Single package:
//! ```text
//! chore(release): my-crate v1.2.0
//! ```
//!
//! Multiple packages (≤3):
//! ```text
//! chore(release): core v1.0.0, utils v2.1.0
//! ```
//!
//! Many packages (>3):
//! ```text
//! chore(release): 5 packages
//! ```
//!
//! ## Body Structure
//!
//! ```text
//! ## 🚀 Release Preparation
//!
//! This PR was automatically created by `clikd release prepare`.
//!
//! ### 📦 Packages
//!
//! | Package | Ecosystem | Version | Bump |
//! |---------|-----------|---------|------|
//! | **my-crate** | 🦀 Rust | `1.0.0` → `1.1.0` | 🟡 MINOR |
//!
//! ### 📝 Changelogs
//! [changelog content here]
//!
//! ### 📋 Release Manifest
//! 📄 `clikd/releases/release-20250605-123456.json`
//!
//! ---
//!
//! ### ✅ Next Steps
//! [automation steps]
//! ```

use std::collections::HashMap;

use super::workflow::SelectedProject;

/// Generates the PR title for a release.
///
/// # Output Examples
///
/// - Single: `chore(release): my-crate v1.2.0`
/// - Few: `chore(release): core v1.0.0, utils v2.1.0`
/// - Many: `chore(release): 5 packages`
pub fn generate_pr_title(projects: &[SelectedProject]) -> String {
    if projects.len() == 1 {
        let p = &projects[0];
        format!("chore(release): {} v{}", p.name, p.new_version)
    } else if projects.len() <= 3 {
        let names: Vec<String> = projects
            .iter()
            .map(|p| format!("{} v{}", p.name, p.new_version))
            .collect();
        format!("chore(release): {}", names.join(", "))
    } else {
        format!("chore(release): {} packages", projects.len())
    }
}

/// Generates the PR body with version table, changelogs, and next steps.
///
/// # Sections
///
/// 1. **Packages table** - Shows each package with ecosystem badge, version diff, and bump badge
/// 2. **Changelogs** - Inline for single package, collapsible `<details>` for multiple
/// 3. **Manifest link** - Points to `clikd/releases/{filename}.json`
/// 4. **Next steps** - Documents GitHub App automation
///
/// # Badge Examples
///
/// Ecosystem badges: `🦀 Rust`, `📦 Node.js`, `🐍 Python`, `🐹 Go`
///
/// Bump badges: `🔴 **MAJOR**`, `🟡 MINOR`, `🟢 patch`
pub fn generate_pr_body(
    projects: &[SelectedProject],
    manifest_filename: &str,
    changelog_contents: &HashMap<String, String>,
) -> String {
    let mut body = String::new();

    body.push_str("## 🚀 Release Preparation\n\n");
    body.push_str("This PR was automatically created by `clikd release prepare`.\n\n");

    body.push_str("### 📦 Packages\n\n");
    body.push_str("| Package | Ecosystem | Version | Bump |\n");
    body.push_str("|---------|-----------|---------|------|\n");

    for project in projects {
        body.push_str(&format!(
            "| **{}** | {} | `{}` → `{}` | {} |\n",
            project.name,
            ecosystem_badge(project.ecosystem.display_name()),
            project.old_version,
            project.new_version,
            bump_badge(&project.bump_type)
        ));
    }

    body.push_str("\n### 📝 Changelogs\n\n");

    if projects.len() == 1 {
        let project = &projects[0];
        if let Some(changelog) = changelog_contents.get(&project.name) {
            body.push_str(changelog);
            body.push('\n');
        }
    } else {
        for project in projects {
            if let Some(changelog) = changelog_contents.get(&project.name) {
                body.push_str(&format!(
                    "<details>\n<summary><strong>{}</strong> - {} → {}</summary>\n\n",
                    project.name, project.old_version, project.new_version
                ));
                body.push_str(changelog);
                body.push_str("\n</details>\n\n");
            }
        }
    }

    body.push_str("### 📋 Release Manifest\n\n");
    body.push_str(&format!("📄 `clikd/releases/{}`\n\n", manifest_filename));

    body.push_str("---\n\n");
    body.push_str("### ✅ Next Steps\n\n");
    body.push_str("After merging this PR, the **clikd GitHub App** will automatically:\n");
    body.push_str("1. Create Git tags for each package\n");
    body.push_str("2. Create GitHub Releases with changelogs\n");
    body.push_str("3. Trigger any configured release workflows\n");

    body
}

fn ecosystem_badge(ecosystem: &str) -> String {
    match ecosystem {
        "Rust" => "🦀 Rust".to_string(),
        "Node.js" => "📦 Node.js".to_string(),
        "Python" => "🐍 Python".to_string(),
        "Go" => "🐹 Go".to_string(),
        "Elixir" => "💧 Elixir".to_string(),
        "C#" => "🔷 C#".to_string(),
        _ => ecosystem.to_string(),
    }
}

fn bump_badge(bump_type: &str) -> String {
    match bump_type.to_lowercase().as_str() {
        "major" => "🔴 **MAJOR**".to_string(),
        "minor" => "🟡 MINOR".to_string(),
        "patch" => "🟢 patch".to_string(),
        _ => bump_type.to_string(),
    }
}
