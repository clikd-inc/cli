use anyhow::{Context, Result};
use tracing::{info, warn};

use crate::{
    atry,
    core::release::{
        commit_analyzer,
        graph::GraphQueryBuilder,
        session::AppSession,
    },
};

#[path = "prepare/wizard.rs"]
mod wizard;

pub fn run(bump: Option<String>, no_tui: bool) -> Result<i32> {
    info!(
        "preparing release with clikd version {}",
        env!("CARGO_PKG_VERSION")
    );

    let use_auto_mode = no_tui || bump.as_deref() == Some("auto");

    if use_auto_mode {
        return run_auto_mode(bump);
    }

    if bump.is_none() || bump.as_deref() == Some("manual") {
        return run_tui_wizard();
    }

    let bump_scheme_text = bump.as_deref().unwrap_or("micro bump");
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
            .filter_map(|cid| {
                sess.repo
                    .get_commit_summary(*cid)
                    .ok()
            })
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

fn run_tui_wizard() -> Result<i32> {
    wizard::run()
}
