use anyhow::{Context, Result};
use tracing::info;

use crate::atry;
use crate::cli::ReleaseOutputFormat;
use crate::core::release::{graph::GraphQueryBuilder, session::AppSession};
use crate::core::ui::utils::is_interactive_terminal;

pub fn run(format: Option<ReleaseOutputFormat>, no_tui: bool) -> Result<i32> {
    info!(
        "checking release status with clikd version {}",
        env!("CARGO_PKG_VERSION")
    );

    let sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    let q = GraphQueryBuilder::default();
    let idents = sess
        .graph()
        .query(q)
        .context("cannot get requested statuses")?;

    let histories = sess.analyze_histories()?;

    let format = format.unwrap_or(ReleaseOutputFormat::Table);
    let use_tui = matches!(format, ReleaseOutputFormat::Table)
        && is_interactive_terminal()
        && !no_tui;

    if use_tui {
        eprintln!("TUI mode not yet implemented. Use --format text or --no-tui for now.");
        eprintln!("Falling back to text mode...\n");
    }

    match format {
        ReleaseOutputFormat::Json => {
            use serde_json::json;

            let mut projects = Vec::new();

            for ident in &idents {
                let proj = sess.graph().lookup(*ident);
                let history = histories.lookup(*ident);
                let n = history.n_commits();
                let rel_info = history.release_info(&sess.repo)?;

                let mut commits = Vec::new();
                for cid in history.commits() {
                    let summary = sess.repo.get_commit_summary(*cid)?;
                    commits.push(summary);
                }

                let project_data = if let Some(this_info) = rel_info.lookup_project(proj) {
                    json!({
                        "name": proj.user_facing_name,
                        "current_version": this_info.version.to_string(),
                        "commits_count": n,
                        "commits": commits,
                        "age": this_info.age,
                    })
                } else {
                    json!({
                        "name": proj.user_facing_name,
                        "current_version": null,
                        "commits_count": n,
                        "commits": commits,
                        "age": null,
                    })
                };

                projects.push(project_data);
            }

            let output = json!({
                "projects": projects
            });

            println!("{}", serde_json::to_string_pretty(&output)?);
        }
        _ => {
            for ident in idents {
                let proj = sess.graph().lookup(ident);
                let history = histories.lookup(ident);
                let n = history.n_commits();
                let rel_info = history.release_info(&sess.repo)?;

                if let Some(this_info) = rel_info.lookup_project(proj) {
                    if this_info.age == 0 {
                        if n == 0 {
                            println!(
                                "{}: no relevant commits since {}",
                                proj.user_facing_name, this_info.version
                            );
                        } else {
                            println!(
                                "{}: {} relevant commit(s) since {}",
                                proj.user_facing_name, n, this_info.version
                            );
                        }
                    } else {
                        println!(
                            "{}: no more than {} relevant commit(s) since {} (unable to track in detail)",
                            proj.user_facing_name, n, this_info.version
                        );
                    }
                } else {
                    println!(
                        "{}: {} relevant commit(s) since start of history (no releases on record)",
                        proj.user_facing_name, n
                    );
                }

                for (idx, cid) in history.commits().into_iter().enumerate() {
                    let summary = sess.repo.get_commit_summary(*cid)?;
                    println!("    {}. {}", idx + 1, summary);
                }

                if n > 0 {
                    println!();
                }
            }
        }
    }

    Ok(0)
}
