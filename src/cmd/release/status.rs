use anyhow::{Context, Result};
use tracing::info;

use crate::atry;
use crate::core::release::{
    session::AppSession,
    graph::GraphQueryBuilder,
};

pub fn run() -> Result<i32> {
    info!("checking release status with clikd version {}", env!("CARGO_PKG_VERSION"));

    let sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    let mut q = GraphQueryBuilder::default();
    let idents = sess
        .graph()
        .query(q)
        .context("cannot get requested statuses")?;

    let histories = sess.analyze_histories()?;

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
                        proj.user_facing_name,
                        n,
                        this_info.version
                    );
                }
            } else {
                println!(
                    "{}: no more than {} relevant commit(s) since {} (unable to track in detail)",
                    proj.user_facing_name,
                    n,
                    this_info.version
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

    Ok(0)
}
