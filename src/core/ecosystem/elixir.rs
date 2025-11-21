use anyhow::anyhow;
use tracing::warn;
use std::{
    collections::HashMap,
    fs::File,
    io::{Read, Write},
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
pub struct ElixirLoader {
    mix_exs_paths: Vec<RepoPathBuf>,
}

impl ElixirLoader {
    pub fn process_index_item(&mut self, dirname: &RepoPath, basename: &RepoPath) {
        if basename.as_ref() != b"mix.exs" {
            return;
        }

        let mut path = dirname.to_owned();
        path.push(basename);
        self.mix_exs_paths.push(path);
    }

    pub fn finalize(
        self,
        app: &mut AppBuilder,
        pconfig: &HashMap<String, ProjectConfiguration>,
    ) -> Result<()> {
        for mix_exs_path in self.mix_exs_paths {
            let (prefix, _) = mix_exs_path.split_basename();
            let fs_path = app.repo.resolve_workdir(&mix_exs_path);

            let mut contents = String::new();
            let mut f = atry!(
                File::open(&fs_path);
                ["failed to open mix.exs file `{}`", fs_path.display()]
            );

            atry!(
                f.read_to_string(&mut contents);
                ["failed to read mix.exs file `{}`", fs_path.display()]
            );

            let app_name = atry!(
                Self::extract_app_name(&contents)
                    .ok_or_else(|| anyhow!("failed to extract app name from mix.exs"));
                ["failed to parse app name from `{}`", fs_path.display()]
            );

            let version_str = Self::extract_version(&contents)
                .unwrap_or_else(|| String::from("0.1.0"));

            let qnames = vec![app_name, "elixir".to_owned()];

            if let Some(ident) = app.graph.try_add_project(qnames, pconfig) {
                let proj = app.graph.lookup_mut(ident);

                let version = match semver::Version::parse(&version_str) {
                    Ok(v) => Version::Semver(v),
                    Err(_) => {
                        warn!("failed to parse version `{}` from mix.exs, using default", version_str);
                        Version::Semver(semver::Version::new(0, 1, 0))
                    }
                };

                proj.version = Some(version);
                proj.prefix = Some(prefix.to_owned());

                let elixir_rewrite = MixExsRewriter::new(ident, mix_exs_path);
                proj.rewriters.push(Box::new(elixir_rewrite));
            }
        }

        Ok(())
    }

    fn extract_app_name(contents: &str) -> Option<String> {
        for line in contents.lines() {
            let trimmed = line.trim();
            if trimmed.starts_with("app:") {
                if let Some(colon_pos) = trimmed.find(':') {
                    let app_part = trimmed[colon_pos + 1..].trim();
                    if let Some(app_part) = app_part.strip_prefix(':') {
                        let name = app_part.trim().trim_end_matches(',');
                        return Some(name.to_string());
                    }
                }
            }
        }
        None
    }

    fn extract_version(contents: &str) -> Option<String> {
        for line in contents.lines() {
            let trimmed = line.trim();
            if trimmed.starts_with("version:") {
                if let Some(colon_pos) = trimmed.find(':') {
                    let version_part = trimmed[colon_pos + 1..].trim();
                    if let Some(version_part) = version_part.strip_prefix('"') {
                        if let Some(end_quote) = version_part.find('"') {
                            return Some(version_part[..end_quote].to_string());
                        }
                    }
                }
            }
        }
        None
    }
}

#[derive(Debug)]
pub struct MixExsRewriter {
    proj_id: ProjectId,
    repo_path: RepoPathBuf,
}

impl MixExsRewriter {
    pub fn new(proj_id: ProjectId, repo_path: RepoPathBuf) -> Self {
        MixExsRewriter { proj_id, repo_path }
    }
}

impl Rewriter for MixExsRewriter {
    fn rewrite(&self, app: &AppSession, changes: &mut ChangeList) -> Result<()> {
        let fs_path = app.repo.resolve_workdir(&self.repo_path);
        let proj = app.graph().lookup(self.proj_id);

        let mut contents = String::new();
        let mut f = atry!(
            File::open(&fs_path);
            ["failed to open mix.exs file `{}`", fs_path.display()]
        );

        atry!(
            f.read_to_string(&mut contents);
            ["failed to read mix.exs file `{}`", fs_path.display()]
        );

        drop(f);

        let new_version = proj.version.to_string();
        let mut new_contents = String::new();

        for line in contents.lines() {
            let trimmed = line.trim();
            if trimmed.starts_with("version:") {
                if let Some(indent) = line.find("version:") {
                    new_contents.push_str(&line[..indent]);
                    new_contents.push_str(&format!("version: \"{}\",\n", new_version));
                    continue;
                }
            }
            new_contents.push_str(line);
            new_contents.push('\n');
        }

        let new_af = atomicwrites::AtomicFile::new(
            &fs_path,
            atomicwrites::OverwriteBehavior::AllowOverwrite,
        );

        let r = new_af.write(|new_f| {
            new_f.write_all(new_contents.as_bytes())?;
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
