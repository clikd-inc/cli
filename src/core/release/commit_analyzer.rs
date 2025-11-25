use anyhow::Result;
use git_conventional::{Commit, Type};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BumpRecommendation {
    Major,
    Minor,
    Patch,
    None,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum ChangelogCategory {
    Added,
    Changed,
    Deprecated,
    Removed,
    Fixed,
    Security,
}

impl ChangelogCategory {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Added => "Added",
            Self::Changed => "Changed",
            Self::Deprecated => "Deprecated",
            Self::Removed => "Removed",
            Self::Fixed => "Fixed",
            Self::Security => "Security",
        }
    }

    pub fn from_conventional_type(commit_type: &Type) -> Option<Self> {
        match commit_type.as_str() {
            "feat" => Some(Self::Added),
            "fix" => Some(Self::Fixed),
            "perf" => Some(Self::Changed),
            "refactor" => Some(Self::Changed),
            "docs" | "chore" | "ci" | "test" | "style" | "build" => None,
            _ => None,
        }
    }
}

#[derive(Debug, Clone)]
pub struct CategorizedCommit {
    pub category: ChangelogCategory,
    pub message: String,
    pub scope: Option<String>,
    pub breaking: bool,
    pub original: String,
}

impl CategorizedCommit {
    pub fn format_for_changelog(&self) -> String {
        let scope_part = self.scope.as_ref().map(|s| format!("**{}**: ", s)).unwrap_or_default();
        let breaking_mark = if self.breaking { " [BREAKING]" } else { "" };
        format!("- {}{}{}", scope_part, self.message, breaking_mark)
    }
}

impl BumpRecommendation {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Major => "major bump",
            Self::Minor => "minor bump",
            Self::Patch => "micro bump",
            Self::None => "no bump",
        }
    }

    pub fn merge(self, other: Self) -> Self {
        match (self, other) {
            (Self::Major, _) | (_, Self::Major) => Self::Major,
            (Self::Minor, _) | (_, Self::Minor) => Self::Minor,
            (Self::Patch, _) | (_, Self::Patch) => Self::Patch,
            (Self::None, Self::None) => Self::None,
        }
    }
}

#[derive(Debug)]
pub struct CommitAnalysis {
    pub recommendation: BumpRecommendation,
    pub total_commits: usize,
    pub feat_count: usize,
    pub fix_count: usize,
    pub breaking_count: usize,
    pub other_count: usize,
}

impl Default for CommitAnalysis {
    fn default() -> Self {
        Self {
            recommendation: BumpRecommendation::None,
            total_commits: 0,
            feat_count: 0,
            fix_count: 0,
            breaking_count: 0,
            other_count: 0,
        }
    }
}

impl CommitAnalysis {
    pub fn summary(&self) -> String {
        if self.total_commits == 0 {
            return "No commits to analyze".to_string();
        }

        let mut parts = Vec::new();

        if self.feat_count > 0 {
            parts.push(format!("{} feat", self.feat_count));
        }
        if self.fix_count > 0 {
            parts.push(format!("{} fix", self.fix_count));
        }
        if self.breaking_count > 0 {
            parts.push(format!("{} BREAKING", self.breaking_count));
        }
        if self.other_count > 0 {
            parts.push(format!("{} other", self.other_count));
        }

        let commits_text = parts.join(", ");

        format!(
            "{} commit{}: {} → suggests {}",
            self.total_commits,
            if self.total_commits == 1 { "" } else { "s" },
            commits_text,
            match self.recommendation {
                BumpRecommendation::Major => "MAJOR",
                BumpRecommendation::Minor => "MINOR",
                BumpRecommendation::Patch => "PATCH",
                BumpRecommendation::None => "NO BUMP",
            }
        )
    }
}

pub fn analyze_commit_messages(messages: &[String]) -> Result<CommitAnalysis> {
    let mut analysis = CommitAnalysis {
        recommendation: BumpRecommendation::None,
        total_commits: messages.len(),
        ..Default::default()
    };

    for message in messages {
        let commit_result = Commit::parse(message);

        let recommendation = match commit_result {
            Ok(commit) => {
                if commit.breaking() {
                    analysis.breaking_count += 1;
                    BumpRecommendation::Major
                } else {
                    let commit_type = commit.type_();
                    if commit_type == Type::FEAT {
                        analysis.feat_count += 1;
                        BumpRecommendation::Minor
                    } else if commit_type == Type::FIX {
                        analysis.fix_count += 1;
                        BumpRecommendation::Patch
                    } else if commit_type == Type::PERF {
                        analysis.fix_count += 1;
                        BumpRecommendation::Patch
                    } else {
                        analysis.other_count += 1;
                        BumpRecommendation::None
                    }
                }
            }
            Err(_) => {
                analysis.other_count += 1;
                BumpRecommendation::None
            }
        };

        analysis.recommendation = analysis.recommendation.merge(recommendation);
    }

    Ok(analysis)
}

pub fn recommend_bump_for_commits(commit_summaries: &[String]) -> Result<BumpRecommendation> {
    let analysis = analyze_commit_messages(commit_summaries)?;
    Ok(analysis.recommendation)
}

pub fn categorize_commits(messages: &[String]) -> Vec<CategorizedCommit> {
    let mut categorized = Vec::new();

    for message in messages {
        let commit_result = Commit::parse(message);

        match commit_result {
            Ok(commit) => {
                let commit_type = commit.type_();
                let breaking = commit.breaking();

                if breaking {
                    categorized.push(CategorizedCommit {
                        category: ChangelogCategory::Changed,
                        message: commit.description().to_string(),
                        scope: commit.scope().map(|s| s.to_string()),
                        breaking: true,
                        original: message.clone(),
                    });
                } else if let Some(category) = ChangelogCategory::from_conventional_type(&commit_type) {
                    categorized.push(CategorizedCommit {
                        category,
                        message: commit.description().to_string(),
                        scope: commit.scope().map(|s| s.to_string()),
                        breaking: false,
                        original: message.clone(),
                    });
                }
            }
            Err(_) => {
            }
        }
    }

    categorized.sort_by_key(|c| c.category);
    categorized
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_feat_recommends_minor() {
        let commits = vec!["feat: add new feature".to_string()];
        let recommendation = recommend_bump_for_commits(&commits).unwrap();
        assert_eq!(recommendation, BumpRecommendation::Minor);
    }

    #[test]
    fn test_fix_recommends_patch() {
        let commits = vec!["fix: correct bug".to_string()];
        let recommendation = recommend_bump_for_commits(&commits).unwrap();
        assert_eq!(recommendation, BumpRecommendation::Patch);
    }

    #[test]
    fn test_breaking_recommends_major() {
        let commits = vec!["feat!: breaking change".to_string()];
        let recommendation = recommend_bump_for_commits(&commits).unwrap();
        assert_eq!(recommendation, BumpRecommendation::Major);
    }

    #[test]
    fn test_breaking_footer_recommends_major() {
        let commits = vec!["feat: add feature\n\nBREAKING CHANGE: breaks API".to_string()];
        let recommendation = recommend_bump_for_commits(&commits).unwrap();
        assert_eq!(recommendation, BumpRecommendation::Major);
    }

    #[test]
    fn test_mixed_commits_picks_highest() {
        let commits = vec![
            "fix: bug fix".to_string(),
            "feat: new feature".to_string(),
            "docs: update README".to_string(),
        ];
        let recommendation = recommend_bump_for_commits(&commits).unwrap();
        assert_eq!(recommendation, BumpRecommendation::Minor);
    }

    #[test]
    fn test_analysis_summary() {
        let commits = vec![
            "feat: add feature".to_string(),
            "fix: fix bug".to_string(),
            "docs: update".to_string(),
        ];
        let analysis = analyze_commit_messages(&commits).unwrap();
        assert_eq!(analysis.feat_count, 1);
        assert_eq!(analysis.fix_count, 1);
        assert_eq!(analysis.other_count, 1);
        assert_eq!(analysis.recommendation, BumpRecommendation::Minor);
    }
}
