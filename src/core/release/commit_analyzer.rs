use anyhow::Result;
use git_conventional::{Commit, Type};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BumpRecommendation {
    Major,
    Minor,
    Patch,
    None,
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
