use anyhow::{Result, Context};
use git2::Repository;
use std::path::Path;

pub struct GitInfo {
    pub branch: String,
    pub commit_hash: String,
    pub commit_message: String,
    pub is_clean: bool,
    pub repo_root: String,
}

pub fn detect_git_info<P: AsRef<Path>>(path: P) -> Result<GitInfo> {
    let repo = Repository::discover(path)
        .context("Failed to find git repository")?;

    let head = repo.head()
        .context("Failed to get HEAD reference")?;

    let branch = if head.is_branch() {
        head.shorthand()
            .unwrap_or("unknown")
            .to_string()
    } else {
        "detached-head".to_string()
    };

    let commit = head.peel_to_commit()
        .context("Failed to get commit from HEAD")?;

    let commit_hash = commit.id().to_string();
    let commit_hash_short = &commit_hash[..8];

    let commit_message = commit.summary()
        .unwrap_or("No commit message")
        .to_string();

    let is_clean = is_working_tree_clean(&repo)?;

    let repo_root = repo.workdir()
        .context("Repository has no working directory")?
        .to_string_lossy()
        .to_string();

    Ok(GitInfo {
        branch,
        commit_hash: commit_hash_short.to_string(),
        commit_message,
        is_clean,
        repo_root,
    })
}

pub fn get_branch_name<P: AsRef<Path>>(path: P) -> Result<String> {
    let git_info = detect_git_info(path)?;
    Ok(git_info.branch)
}

pub fn sanitize_branch_name(branch: &str) -> String {
    branch
        .chars()
        .map(|c| match c {
            '/' | '\\' | ':' | '*' | '?' | '"' | '<' | '>' | '|' => '_',
            c if c.is_ascii_alphanumeric() || c == '-' || c == '_' => c,
            _ => '_',
        })
        .collect()
}

fn is_working_tree_clean(repo: &Repository) -> Result<bool> {
    let statuses = repo.statuses(None)
        .context("Failed to get repository status")?;

    Ok(statuses.is_empty())
}

pub fn get_main_branch<P: AsRef<Path>>(path: P) -> Result<String> {
    let repo = Repository::discover(path)
        .context("Failed to find git repository")?;

    let branches = repo.branches(Some(git2::BranchType::Local))
        .context("Failed to list branches")?;

    for branch_result in branches {
        let (branch, _) = branch_result.context("Failed to get branch")?;
        if let Some(name) = branch.name()? {
            if name == "main" || name == "master" {
                return Ok(name.to_string());
            }
        }
    }

    Ok("main".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sanitize_branch_name() {
        assert_eq!(sanitize_branch_name("feat/user-profiles"), "feat_user-profiles");
        assert_eq!(sanitize_branch_name("fix/auth:issue"), "fix_auth_issue");
        assert_eq!(sanitize_branch_name("main"), "main");
        assert_eq!(sanitize_branch_name("feature/special-chars!@#"), "feature_special-chars___");
    }
}