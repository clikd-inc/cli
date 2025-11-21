use anyhow::anyhow;
use tracing::warn;
use std::{
    collections::HashMap,
    fs::File,
    io::{BufRead, BufReader, Read, Write},
};

use crate::{
    atry,
    core::release::{
        session::{AppBuilder, AppSession},
        config::ProjectConfiguration,
        errors::Result,
        project::ProjectId,
        repository::{ChangeList, RepoPath, RepoPathBuf},
        rewriters::Rewriter,
        version::Version,
    },
};

#[derive(Debug, Default)]
pub struct GoLoader {
    go_mod_paths: Vec<RepoPathBuf>,
}

impl GoLoader {
    pub fn process_index_item(&mut self, dirname: &RepoPath, basename: &RepoPath) {
        if basename.as_ref() != b"go.mod" {
            return;
        }

        let mut path = dirname.to_owned();
        path.push(basename);
        self.go_mod_paths.push(path);
    }

    pub fn finalize(
        self,
        app: &mut AppBuilder,
        pconfig: &HashMap<String, ProjectConfiguration>,
    ) -> Result<()> {
        for go_mod_path in self.go_mod_paths {
            let (prefix, _) = go_mod_path.split_basename();
            let fs_path = app.repo.resolve_workdir(&go_mod_path);

            let f = atry!(
                File::open(&fs_path);
                ["failed to open go.mod file `{}`", fs_path.display()]
            );

            let reader = BufReader::new(f);
            let mut module_name = None;

            for line_result in reader.lines() {
                let line = line_result?;
                let trimmed = line.trim();

                if trimmed.starts_with("module ") {
                    module_name = Some(trimmed[7..].trim().to_string());
                    break;
                }
            }

            let module_name = atry!(
                module_name.ok_or_else(|| anyhow!("no module declaration found"));
                ["failed to parse module name from `{}`", fs_path.display()]
            );

            let qnames = vec![module_name, "go".to_owned()];

            if let Some(ident) = app.graph.try_add_project(qnames, pconfig) {
                let proj = app.graph.lookup_mut(ident);
                proj.version = Some(Version::Semver(semver::Version::new(0, 0, 0)));
                proj.prefix = Some(prefix.to_owned());

                let go_rewrite = GoModRewriter::new(ident, go_mod_path);
                proj.rewriters.push(Box::new(go_rewrite));
            }
        }

        Ok(())
    }
}

#[derive(Debug)]
pub struct GoModRewriter {
    proj_id: ProjectId,
    repo_path: RepoPathBuf,
}

impl GoModRewriter {
    pub fn new(proj_id: ProjectId, repo_path: RepoPathBuf) -> Self {
        GoModRewriter { proj_id, repo_path }
    }
}

impl Rewriter for GoModRewriter {
    fn rewrite(&self, app: &AppSession, changes: &mut ChangeList) -> Result<()> {
        let fs_path = app.repo.resolve_workdir(&self.repo_path);

        let f = atry!(
            File::open(&fs_path);
            ["failed to open go.mod file `{}`", fs_path.display()]
        );

        let reader = BufReader::new(f);
        let mut lines = Vec::new();

        for line_result in reader.lines() {
            lines.push(line_result?);
        }

        let new_af = atomicwrites::AtomicFile::new(
            &fs_path,
            atomicwrites::OverwriteBehavior::AllowOverwrite,
        );

        let r = new_af.write(|new_f| {
            for line in &lines {
                writeln!(new_f, "{}", line)?;
            }
            Ok(())
        });

        changes.add_path(&self.repo_path);

        match r {
            Err(atomicwrites::Error::Internal(e)) => Err(e.into()),
            Err(atomicwrites::Error::User(e)) => Err(e),
            Ok(()) => Ok(()),
        }
    }
}
