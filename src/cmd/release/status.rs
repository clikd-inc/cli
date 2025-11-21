use anyhow::Result;
use log::info;

use crate::atry;
use crate::core::release::session::AppSession;

pub fn run() -> Result<i32> {
    info!("checking release status with clikd version {}", env!("CARGO_PKG_VERSION"));

    let _repo = atry!(
        crate::core::release::repository::Repository::open_from_env();
        ["clikd is not being run from a Git working directory"]
    );

    let sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    let mut seen_any = false;

    for ident in sess.graph().toposorted() {
        let proj = sess.graph().lookup(ident);

        if !seen_any {
            info!("Projects in repository:");
            println!();
            seen_any = true;
        }

        let loc_desc = {
            let p = proj.prefix();

            if p.len() == 0 {
                "root".to_owned()
            } else {
                format!("`{}`", p.escaped())
            }
        };

        println!(
            "  {} @ {} ({})",
            proj.user_facing_name, proj.version, loc_desc
        );
    }

    if !seen_any {
        info!("No projects detected in repository");
        return Ok(1);
    }

    println!();
    Ok(0)
}
