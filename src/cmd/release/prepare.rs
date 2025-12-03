use anyhow::{Context, Result};
use tracing::{info, warn};

use crate::{
    atry,
    core::release::{
        changelog_generator::{self, ChangelogEntry},
        commit_analyzer,
        graph::GraphQueryBuilder,
        repository::RepoPathBuf,
        session::AppSession,
    },
};

#[path = "prepare/wizard.rs"]
mod wizard;

#[derive(Debug, Clone)]
struct PreparedProject {
    name: String,
    prefix: String,
    old_version: String,
    new_version: String,
    bump_type: String,
    commit_messages: Vec<String>,
}

pub fn run(
    bump: Option<String>,
    no_tui: bool,
    ci: bool,
    push: bool,
    github_release: bool,
    project: Option<Vec<String>>,
) -> Result<i32> {
    info!(
        "preparing release with clikd version {}",
        env!("CARGO_PKG_VERSION")
    );

    if ci {
        return run_ci_mode(push, github_release);
    }

    if let Some(ref projects) = project {
        return run_per_project_mode(projects);
    }

    let use_auto_mode = no_tui || bump.as_deref() == Some("auto");

    if use_auto_mode {
        return run_auto_mode(bump);
    }

    if bump.is_none() || bump.as_deref() == Some("manual") {
        return run_tui_wizard();
    }

    let bump_scheme_text = bump.as_deref().unwrap_or("patch");
    info!("version bump scheme: {}", bump_scheme_text);

    let mut sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    if let Some(dirty) = atry!(
        sess.repo.check_if_dirty(&[]);
        ["failed to check repository for modified files"]
    ) {
        warn!(
            "preparing release with uncommitted changes in the repository (e.g.: `{}`)",
            dirty.escaped()
        );
    }

    let q = GraphQueryBuilder::default();
    let idents = sess.graph().query(q).context("could not select projects")?;

    if idents.is_empty() {
        info!("no projects found in repository");
        return Ok(0);
    }

    let histories = atry!(
        sess.analyze_histories();
        ["failed to analyze project histories"]
    );

    let mut n_prepared = 0;
    let mut n_skipped = 0;

    for ident in &idents {
        let proj = sess.graph().lookup(*ident);
        let history = histories.lookup(*ident);
        let n_commits = history.n_commits();

        if n_commits == 0 {
            info!(
                "{}: no changes since last release, skipping",
                proj.user_facing_name
            );
            n_skipped += 1;
            continue;
        }

        let bump_scheme = proj
            .version
            .parse_bump_scheme(bump_scheme_text)
            .with_context(|| {
                format!(
                    "invalid bump scheme \"{}\" for project {}",
                    bump_scheme_text, proj.user_facing_name
                )
            })?;

        let proj_mut = sess.graph_mut().lookup_mut(*ident);
        let old_version = proj_mut.version.clone();

        atry!(
            bump_scheme.apply(&mut proj_mut.version);
            ["failed to apply version bump to {}", proj_mut.user_facing_name]
        );

        info!(
            "{}: {} -> {} ({} commit{})",
            proj_mut.user_facing_name,
            old_version,
            proj_mut.version,
            n_commits,
            if n_commits == 1 { "" } else { "s" }
        );

        n_prepared += 1;
    }

    if n_prepared == 0 {
        info!("no projects needed version bumps");
        return Ok(0);
    }

    info!("updating project files with new versions...");

    let changes = atry!(
        sess.rewrite();
        ["failed to update project files"]
    );

    if changes.paths().count() > 0 {
        println!();
        info!("modified files:");
        for path in changes.paths() {
            println!("  {}", path.escaped());
        }
    }

    println!();
    info!(
        "prepared {} project{} for release ({} skipped)",
        n_prepared,
        if n_prepared == 1 { "" } else { "s" },
        n_skipped
    );
    info!("review changes and commit when ready");

    Ok(0)
}

fn run_ci_mode(push: bool, github_release: bool) -> Result<i32> {
    info!("running in CI mode (full automation)");

    let mut sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    if let Some(dirty) = atry!(
        sess.repo.check_if_dirty(&[]);
        ["failed to check repository for modified files"]
    ) {
        return Err(anyhow::anyhow!(
            "CI mode requires a clean working directory. Found uncommitted changes: {}",
            dirty.escaped()
        ));
    }

    let q = GraphQueryBuilder::default();
    let idents = sess.graph().query(q).context("could not select projects")?;

    if idents.is_empty() {
        info!("no projects found in repository");
        return Ok(0);
    }

    let histories = atry!(
        sess.analyze_histories();
        ["failed to analyze project histories"]
    );

    let ai_enabled = sess.changelog_config.ai_enabled;

    let mut prepared_projects: Vec<PreparedProject> = Vec::new();

    for ident in &idents {
        let proj = sess.graph().lookup(*ident);
        let history = histories.lookup(*ident);
        let n_commits = history.n_commits();

        if n_commits == 0 {
            info!(
                "{}: no changes since last release, skipping",
                proj.user_facing_name
            );
            continue;
        }

        let commit_messages: Vec<String> = history
            .commits()
            .into_iter()
            .filter_map(|cid| sess.repo.get_commit_summary(*cid).ok())
            .collect();

        let analysis = atry!(
            commit_analyzer::analyze_commit_messages(&commit_messages);
            ["failed to analyze commit messages for {}", proj.user_facing_name]
        );

        info!("{}: {}", proj.user_facing_name, analysis.summary());

        let bump_scheme_text = analysis.recommendation.as_str();

        if bump_scheme_text == "no bump" {
            info!(
                "{}: no version bump needed based on conventional commits",
                proj.user_facing_name
            );
            continue;
        }

        let bump_scheme = proj
            .version
            .parse_bump_scheme(bump_scheme_text)
            .with_context(|| {
                format!(
                    "invalid bump scheme \"{}\" for project {}",
                    bump_scheme_text, proj.user_facing_name
                )
            })?;

        let old_version = proj.version.to_string();
        let prefix = proj.prefix().escaped();

        let proj_mut = sess.graph_mut().lookup_mut(*ident);

        atry!(
            bump_scheme.apply(&mut proj_mut.version);
            ["failed to apply version bump to {}", proj_mut.user_facing_name]
        );

        let new_version = proj_mut.version.to_string();

        info!(
            "{}: {} -> {} ({})",
            proj_mut.user_facing_name, old_version, new_version, bump_scheme_text
        );

        prepared_projects.push(PreparedProject {
            name: proj_mut.user_facing_name.clone(),
            prefix,
            old_version,
            new_version,
            bump_type: bump_scheme_text.to_string(),
            commit_messages,
        });
    }

    if prepared_projects.is_empty() {
        info!("no projects needed version bumps");
        return Ok(0);
    }

    info!("updating project files with new versions...");

    let changes = atry!(
        sess.rewrite();
        ["failed to update project files"]
    );

    info!("generating changelogs...");

    let mut changelog_paths: Vec<RepoPathBuf> = Vec::new();

    for project in &prepared_projects {
        let categorized = commit_analyzer::categorize_commits(&project.commit_messages);

        if categorized.is_empty() {
            info!(
                "{}: no user-facing changes, skipping changelog",
                project.name
            );
            continue;
        }

        let mut entry = ChangelogEntry::new(project.new_version.clone());
        entry.add_commits(&categorized);

        let draft_changelog = entry.to_markdown();

        let final_changelog_entry = if ai_enabled {
            info!("{}: polishing changelog with AI...", project.name);

            match polish_changelog_with_ai(&draft_changelog, &project.commit_messages) {
                Ok(polished) => polished,
                Err(e) => {
                    warn!(
                        "{}: AI polish failed ({}), using standard changelog",
                        project.name, e
                    );
                    draft_changelog
                }
            }
        } else {
            draft_changelog
        };

        let changelog_rel_path = if project.prefix.is_empty() {
            "CHANGELOG.md".to_string()
        } else {
            format!("{}/CHANGELOG.md", project.prefix)
        };

        let changelog_repo_path = RepoPathBuf::new(changelog_rel_path.as_bytes());
        let changelog_full_path = sess.repo.resolve_workdir(changelog_repo_path.as_ref());

        let existing_content =
            changelog_generator::parse_existing_changelog(&changelog_full_path).unwrap_or_default();

        let full_changelog =
            changelog_generator::generate_changelog(&project.name, &entry, &existing_content);

        let final_content = if ai_enabled && !final_changelog_entry.is_empty() {
            let header = format!(
                "# Changelog\n\n\
                All notable changes to {} will be documented in this file.\n\n\
                The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n\
                and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n",
                project.name
            );
            let ai_entry = if final_changelog_entry.starts_with("## [") {
                final_changelog_entry.clone()
            } else {
                entry.to_markdown()
            };
            format!("{}{}\n{}", header, ai_entry, existing_content)
        } else {
            full_changelog
        };

        if let Some(parent) = changelog_full_path.parent() {
            std::fs::create_dir_all(parent).with_context(|| {
                format!(
                    "failed to create directory for {}",
                    changelog_full_path.display()
                )
            })?;
        }

        std::fs::write(&changelog_full_path, &final_content).with_context(|| {
            format!(
                "failed to write changelog to {}",
                changelog_full_path.display()
            )
        })?;

        changelog_paths.push(changelog_repo_path);
        info!(
            "{}: wrote changelog to {}",
            project.name, changelog_rel_path
        );
    }

    let all_changed_paths: Vec<&crate::core::release::repository::RepoPath> = changes
        .paths()
        .chain(changelog_paths.iter().map(|p| p.as_ref()))
        .collect();

    if !all_changed_paths.is_empty() {
        println!();
        info!("modified files:");
        for path in &all_changed_paths {
            println!("  {}", path.escaped());
        }
    }

    let project_versions: Vec<(String, String)> = prepared_projects
        .iter()
        .map(|p| (p.name.clone(), p.new_version.clone()))
        .collect();

    info!("creating release commit...");

    let commit_message = format_commit_message(&prepared_projects);

    atry!(
        sess.repo.create_commit(&commit_message, &all_changed_paths);
        ["failed to create release commit"]
    );

    info!("creating release tags...");

    atry!(
        sess.repo.create_release_tags(&project_versions);
        ["failed to create release tags"]
    );

    for (name, version) in &project_versions {
        info!("  created tag: {}-v{}", name, version);
    }

    if push {
        info!("pushing to remote...");

        atry!(
            push_to_remote();
            ["failed to push to remote"]
        );

        if github_release {
            info!("creating GitHub releases...");

            for project in &prepared_projects {
                let tag_name = format!("{}-v{}", project.name, project.new_version);

                let changelog_content =
                    get_changelog_for_version(&sess, &project.prefix, &project.new_version)?;

                match create_github_release(
                    &sess,
                    &tag_name,
                    &project.name,
                    &project.new_version,
                    &changelog_content,
                ) {
                    Ok(_) => info!("  created GitHub release: {}", tag_name),
                    Err(e) => warn!("  failed to create GitHub release for {}: {}", tag_name, e),
                }
            }
        }
    }

    println!();
    info!(
        "successfully released {} project{}",
        prepared_projects.len(),
        if prepared_projects.len() == 1 {
            ""
        } else {
            "s"
        }
    );

    for project in &prepared_projects {
        println!(
            "  {} {} -> {} ({})",
            project.name, project.old_version, project.new_version, project.bump_type
        );
    }

    if !push {
        println!();
        info!("run with --push to push commits and tags to remote");
        info!("run with --push --github-release to also create GitHub releases");
    }

    Ok(0)
}

fn format_commit_message(projects: &[PreparedProject]) -> String {
    if projects.len() == 1 {
        let p = &projects[0];
        format!(
            "chore(release): {} v{}\n\n\
            Bump {} from {} to {}",
            p.name, p.new_version, p.name, p.old_version, p.new_version
        )
    } else {
        let mut msg = format!("chore(release): release {} packages\n\n", projects.len());
        for p in projects {
            msg.push_str(&format!(
                "- {}: {} -> {}\n",
                p.name, p.old_version, p.new_version
            ));
        }
        msg
    }
}

fn push_to_remote() -> Result<()> {
    use std::process::Command;

    let output = Command::new("git")
        .args(["push"])
        .output()
        .context("failed to execute git push")?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(anyhow::anyhow!("git push failed: {}", stderr));
    }

    let output = Command::new("git")
        .args(["push", "--tags"])
        .output()
        .context("failed to execute git push --tags")?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(anyhow::anyhow!("git push --tags failed: {}", stderr));
    }

    Ok(())
}

fn get_changelog_for_version(sess: &AppSession, prefix: &str, version: &str) -> Result<String> {
    let changelog_rel_path = if prefix.is_empty() {
        "CHANGELOG.md".to_string()
    } else {
        format!("{}/CHANGELOG.md", prefix)
    };

    let changelog_repo_path = RepoPathBuf::new(changelog_rel_path.as_bytes());
    let changelog_full_path = sess.repo.resolve_workdir(changelog_repo_path.as_ref());

    let content = std::fs::read_to_string(&changelog_full_path).unwrap_or_default();

    let version_header = format!("## [{}]", version);
    let mut in_version_section = false;
    let mut changelog_section = String::new();

    for line in content.lines() {
        if line.starts_with(&version_header) {
            in_version_section = true;
            changelog_section.push_str(line);
            changelog_section.push('\n');
        } else if in_version_section {
            if line.starts_with("## [") {
                break;
            }
            changelog_section.push_str(line);
            changelog_section.push('\n');
        }
    }

    Ok(changelog_section)
}

fn create_github_release(
    sess: &AppSession,
    tag_name: &str,
    package_name: &str,
    version: &str,
    body: &str,
) -> Result<()> {
    use crate::core::release::env::require_var;

    let token = require_var("GITHUB_TOKEN")?;
    let upstream_url = sess.repo.upstream_url()?;

    let upstream_url = git_url_parse::GitUrl::parse(&upstream_url)
        .map_err(|e| anyhow::anyhow!("cannot parse upstream Git URL: {}", e))?;

    let provider: git_url_parse::types::provider::GenericProvider = upstream_url
        .provider_info()
        .map_err(|e| anyhow::anyhow!("cannot extract provider info: {}", e))?;

    let slug = format!("{}/{}", provider.owner(), provider.repo());
    let api_url = format!("https://api.github.com/repos/{}/releases", slug);

    let client = reqwest::blocking::Client::new();

    let release_body = json::object! {
        "tag_name" => tag_name,
        "name" => format!("{} v{}", package_name, version),
        "body" => body,
        "draft" => false,
        "prerelease" => version.contains("-"),
    };

    let response = client
        .post(&api_url)
        .header("Authorization", format!("token {}", token))
        .header("User-Agent", "clikd")
        .header("Accept", "application/vnd.github.v3+json")
        .body(json::stringify(release_body))
        .send()
        .context("failed to send GitHub API request")?;

    if !response.status().is_success() {
        let status = response.status();
        let body = response
            .text()
            .unwrap_or_else(|_| "unknown error".to_string());
        return Err(anyhow::anyhow!(
            "GitHub API request failed ({}): {}",
            status,
            body
        ));
    }

    Ok(())
}

fn polish_changelog_with_ai(draft: &str, commits: &[String]) -> Result<String> {
    use crate::core::ai::changelog::AiChangelogGenerator;

    let rt = tokio::runtime::Runtime::new().context("failed to create async runtime")?;

    rt.block_on(async {
        let generator = AiChangelogGenerator::new()
            .await
            .context("failed to initialize AI changelog generator")?;

        generator
            .polish(draft, commits)
            .await
            .context("failed to polish changelog with AI")
    })
}

fn run_auto_mode(bump: Option<String>) -> Result<i32> {
    info!("running in auto mode (using conventional commits analysis)");

    let mut sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    if let Some(dirty) = atry!(
        sess.repo.check_if_dirty(&[]);
        ["failed to check repository for modified files"]
    ) {
        warn!(
            "preparing release with uncommitted changes in the repository (e.g.: `{}`)",
            dirty.escaped()
        );
    }

    let q = GraphQueryBuilder::default();
    let idents = sess.graph().query(q).context("could not select projects")?;

    if idents.is_empty() {
        info!("no projects found in repository");
        return Ok(0);
    }

    let histories = atry!(
        sess.analyze_histories();
        ["failed to analyze project histories"]
    );

    let mut n_prepared = 0;
    let mut n_skipped = 0;

    for ident in &idents {
        let proj = sess.graph().lookup(*ident);
        let history = histories.lookup(*ident);
        let n_commits = history.n_commits();

        if n_commits == 0 {
            info!(
                "{}: no changes since last release, skipping",
                proj.user_facing_name
            );
            n_skipped += 1;
            continue;
        }

        let commit_messages: Vec<String> = history
            .commits()
            .into_iter()
            .filter_map(|cid| sess.repo.get_commit_summary(*cid).ok())
            .collect();

        let analysis = atry!(
            commit_analyzer::analyze_commit_messages(&commit_messages);
            ["failed to analyze commit messages for {}", proj.user_facing_name]
        );

        info!("{}: {}", proj.user_facing_name, analysis.summary());

        let bump_scheme_text = if let Some(ref explicit_bump) = bump {
            if explicit_bump == "auto" {
                analysis.recommendation.as_str()
            } else {
                explicit_bump.as_str()
            }
        } else {
            analysis.recommendation.as_str()
        };

        if bump_scheme_text == "no bump" {
            info!(
                "{}: no version bump needed based on conventional commits",
                proj.user_facing_name
            );
            n_skipped += 1;
            continue;
        }

        let bump_scheme = proj
            .version
            .parse_bump_scheme(bump_scheme_text)
            .with_context(|| {
                format!(
                    "invalid bump scheme \"{}\" for project {}",
                    bump_scheme_text, proj.user_facing_name
                )
            })?;

        let proj_mut = sess.graph_mut().lookup_mut(*ident);
        let old_version = proj_mut.version.clone();

        atry!(
            bump_scheme.apply(&mut proj_mut.version);
            ["failed to apply version bump to {}", proj_mut.user_facing_name]
        );

        info!(
            "{}: {} -> {} ({} commit{})",
            proj_mut.user_facing_name,
            old_version,
            proj_mut.version,
            n_commits,
            if n_commits == 1 { "" } else { "s" }
        );

        n_prepared += 1;
    }

    if n_prepared == 0 {
        info!("no projects needed version bumps");
        return Ok(0);
    }

    info!("updating project files with new versions...");

    let changes = atry!(
        sess.rewrite();
        ["failed to update project files"]
    );

    if changes.paths().count() > 0 {
        println!();
        info!("modified files:");
        for path in changes.paths() {
            println!("  {}", path.escaped());
        }
    }

    println!();
    info!(
        "prepared {} project{} for release ({} skipped)",
        n_prepared,
        if n_prepared == 1 { "" } else { "s" },
        n_skipped
    );
    info!("review changes and commit when ready");

    Ok(0)
}

fn run_per_project_mode(projects: &[String]) -> Result<i32> {
    use std::collections::HashMap;

    println!("Running in per-project mode");

    let mut bump_specs: HashMap<String, String> = HashMap::new();
    for spec in projects {
        let parts: Vec<&str> = spec.split(':').collect();
        if parts.len() != 2 {
            return Err(anyhow::anyhow!(
                "invalid project spec '{}': expected format 'project:bump' (e.g., gate:major)",
                spec
            ));
        }
        bump_specs.insert(parts[0].to_string(), parts[1].to_string());
    }

    let mut sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    if let Some(dirty) = atry!(
        sess.repo.check_if_dirty(&[]);
        ["failed to check repository for modified files"]
    ) {
        warn!(
            "preparing release with uncommitted changes in the repository (e.g.: `{}`)",
            dirty.escaped()
        );
    }

    let q = GraphQueryBuilder::default();
    let idents = sess.graph().query(q).context("could not select projects")?;

    if idents.is_empty() {
        info!("no projects found in repository");
        return Ok(0);
    }

    let histories = atry!(
        sess.analyze_histories();
        ["failed to analyze project histories"]
    );

    let mut n_prepared = 0;
    let mut n_skipped = 0;

    for ident in &idents {
        let proj = sess.graph().lookup(*ident);
        let history = histories.lookup(*ident);
        let n_commits = history.n_commits();

        let bump_scheme_text = match bump_specs.get(&proj.user_facing_name) {
            Some(bump) => bump.as_str(),
            None => {
                if n_commits == 0 {
                    println!(
                        "{}: no changes and no explicit bump, skipping",
                        proj.user_facing_name
                    );
                } else {
                    println!(
                        "{}: no explicit bump specified, skipping ({} commit{})",
                        proj.user_facing_name,
                        n_commits,
                        if n_commits == 1 { "" } else { "s" }
                    );
                }
                n_skipped += 1;
                continue;
            }
        };

        let bump_scheme = proj
            .version
            .parse_bump_scheme(bump_scheme_text)
            .with_context(|| {
                format!(
                    "invalid bump scheme \"{}\" for project {}",
                    bump_scheme_text, proj.user_facing_name
                )
            })?;

        let proj_mut = sess.graph_mut().lookup_mut(*ident);
        let old_version = proj_mut.version.clone();

        atry!(
            bump_scheme.apply(&mut proj_mut.version);
            ["failed to apply version bump to {}", proj_mut.user_facing_name]
        );

        println!(
            "{}: {} -> {} ({})",
            proj_mut.user_facing_name, old_version, proj_mut.version, bump_scheme_text
        );

        n_prepared += 1;
    }

    if n_prepared == 0 {
        println!("No projects matched the specified bumps");
        return Ok(0);
    }

    println!("Updating project files with new versions...");

    let changes = atry!(
        sess.rewrite();
        ["failed to update project files"]
    );

    if changes.paths().count() > 0 {
        println!();
        println!("Modified files:");
        for path in changes.paths() {
            println!("  {}", path.escaped());
        }
    }

    println!();
    println!(
        "Prepared {} project{} for release ({} skipped)",
        n_prepared,
        if n_prepared == 1 { "" } else { "s" },
        n_skipped
    );
    println!("Review changes and commit when ready");

    Ok(0)
}

fn run_tui_wizard() -> Result<i32> {
    wizard::run()
}
