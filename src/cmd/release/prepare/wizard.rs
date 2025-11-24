use anyhow::{Context, Result};
use crossterm::{
    event::{self, Event, KeyCode, KeyEvent, KeyEventKind},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::CrosstermBackend,
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span, Text},
    widgets::{Block, Borders, List, ListItem, ListState, Padding, Paragraph, Wrap},
    Frame, Terminal,
};
use std::io;
use tracing::info;

use crate::{
    atry,
    core::{
        release::{
            commit_analyzer::{self, BumpRecommendation},
            graph::GraphQueryBuilder,
            project::ProjectId,
            session::AppSession,
        },
        ui::markdown,
    },
};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum WizardStep {
    ProjectSelection,
    BumpStrategy,
    ChangelogPreview,
    Confirmation,
}

impl WizardStep {
    fn title(&self) -> &'static str {
        match self {
            Self::ProjectSelection => "Step 1/4: Select Projects",
            Self::BumpStrategy => "Step 2/4: Choose Bump Strategy",
            Self::ChangelogPreview => "Step 3/4: Preview Changelog",
            Self::Confirmation => "Step 4/4: Confirm Changes",
        }
    }

    fn next(&self) -> Option<Self> {
        match self {
            Self::ProjectSelection => Some(Self::BumpStrategy),
            Self::BumpStrategy => Some(Self::ChangelogPreview),
            Self::ChangelogPreview => Some(Self::Confirmation),
            Self::Confirmation => None,
        }
    }

    fn prev(&self) -> Option<Self> {
        match self {
            Self::ProjectSelection => None,
            Self::BumpStrategy => Some(Self::ProjectSelection),
            Self::ChangelogPreview => Some(Self::BumpStrategy),
            Self::Confirmation => Some(Self::ChangelogPreview),
        }
    }
}

struct ProjectItem {
    ident: ProjectId,
    name: String,
    selected: bool,
    commit_count: usize,
    suggested_bump: BumpRecommendation,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum BumpStrategy {
    Auto,
    Major,
    Minor,
    Patch,
}

impl BumpStrategy {
    fn as_str(&self) -> &'static str {
        match self {
            Self::Auto => "auto",
            Self::Major => "major bump",
            Self::Minor => "minor bump",
            Self::Patch => "micro bump",
        }
    }

    fn description(&self) -> &'static str {
        match self {
            Self::Auto => "Use conventional commits to determine bump type automatically",
            Self::Major => "Breaking changes (1.0.0 → 2.0.0)",
            Self::Minor => "New features (1.0.0 → 1.1.0)",
            Self::Patch => "Bug fixes (1.0.0 → 1.0.1)",
        }
    }

    fn all() -> Vec<Self> {
        vec![Self::Auto, Self::Major, Self::Minor, Self::Patch]
    }
}

struct WizardState {
    step: WizardStep,
    projects: Vec<ProjectItem>,
    project_list_state: ListState,
    selected_bump: BumpStrategy,
    bump_list_state: ListState,
    show_help: bool,
}

impl WizardState {
    fn new(projects: Vec<ProjectItem>) -> Self {
        let mut project_list_state = ListState::default();
        if !projects.is_empty() {
            project_list_state.select(Some(0));
        }

        let mut bump_list_state = ListState::default();
        bump_list_state.select(Some(0));

        Self {
            step: WizardStep::ProjectSelection,
            projects,
            project_list_state,
            selected_bump: BumpStrategy::Auto,
            bump_list_state,
            show_help: false,
        }
    }

    fn toggle_help(&mut self) {
        self.show_help = !self.show_help;
    }

    fn next_step(&mut self) -> bool {
        if let Some(next) = self.step.next() {
            self.step = next;
            true
        } else {
            false
        }
    }

    fn prev_step(&mut self) -> bool {
        if let Some(prev) = self.step.prev() {
            self.step = prev;
            true
        } else {
            false
        }
    }

    fn selected_projects(&self) -> Vec<&ProjectItem> {
        self.projects.iter().filter(|p| p.selected).collect()
    }

    fn handle_key_project_selection(&mut self, key: KeyCode) -> bool {
        match key {
            KeyCode::Up => {
                if let Some(selected) = self.project_list_state.selected() {
                    if selected > 0 {
                        self.project_list_state.select(Some(selected - 1));
                    }
                }
            }
            KeyCode::Down => {
                if let Some(selected) = self.project_list_state.selected() {
                    if selected < self.projects.len() - 1 {
                        self.project_list_state.select(Some(selected + 1));
                    }
                }
            }
            KeyCode::Char(' ') => {
                if let Some(selected) = self.project_list_state.selected() {
                    self.projects[selected].selected = !self.projects[selected].selected;
                }
            }
            KeyCode::Char('a') => {
                let all_selected = self.projects.iter().all(|p| p.selected);
                for project in &mut self.projects {
                    project.selected = !all_selected;
                }
            }
            KeyCode::Enter => {
                if self.selected_projects().is_empty() {
                    return false;
                }
                return self.next_step();
            }
            _ => {}
        }
        false
    }

    fn handle_key_bump_strategy(&mut self, key: KeyCode) -> bool {
        match key {
            KeyCode::Up => {
                if let Some(selected) = self.bump_list_state.selected() {
                    if selected > 0 {
                        self.bump_list_state.select(Some(selected - 1));
                        self.selected_bump = BumpStrategy::all()[selected - 1];
                    }
                }
            }
            KeyCode::Down => {
                if let Some(selected) = self.bump_list_state.selected() {
                    let strategies = BumpStrategy::all();
                    if selected < strategies.len() - 1 {
                        self.bump_list_state.select(Some(selected + 1));
                        self.selected_bump = strategies[selected + 1];
                    }
                }
            }
            KeyCode::Enter => {
                return self.next_step();
            }
            KeyCode::Backspace | KeyCode::Esc => {
                return self.prev_step();
            }
            _ => {}
        }
        false
    }

    fn handle_key_changelog(&mut self, key: KeyCode) -> bool {
        match key {
            KeyCode::Enter => self.next_step(),
            KeyCode::Backspace | KeyCode::Esc => self.prev_step(),
            _ => false,
        }
    }

    fn handle_key_confirmation(&mut self, key: KeyCode) -> (bool, bool) {
        match key {
            KeyCode::Enter => (false, true),
            KeyCode::Backspace | KeyCode::Esc => (self.prev_step(), false),
            _ => (false, false),
        }
    }
}

pub fn run() -> Result<i32> {
    info!("starting interactive TUI wizard for release preparation");

    let mut sess = AppSession::initialize_default()
        .context("could not initialize app and project graph")?;

    if let Some(dirty) = sess
        .repo
        .check_if_dirty(&[])
        .context("failed to check repository for modified files")?
    {
        info!(
            "preparing release with uncommitted changes in the repository (e.g.: `{}`)",
            dirty.escaped()
        );
    }

    let q = GraphQueryBuilder::default();
    let idents = sess
        .graph()
        .query(q)
        .context("could not select projects")?;

    if idents.is_empty() {
        info!("no projects found in repository");
        return Ok(0);
    }

    let histories = sess
        .analyze_histories()
        .context("failed to analyze project histories")?;

    let mut projects = Vec::new();
    for ident in &idents {
        let proj = sess.graph().lookup(*ident);
        let history = histories.lookup(*ident);
        let n_commits = history.n_commits();

        if n_commits == 0 {
            continue;
        }

        let commit_messages: Vec<String> = history
            .commits()
            .into_iter()
            .filter_map(|cid| sess.repo.get_commit_summary(*cid).ok())
            .collect();

        let analysis = commit_analyzer::analyze_commit_messages(&commit_messages)
            .context("failed to analyze commit messages")?;

        projects.push(ProjectItem {
            ident: *ident,
            name: proj.user_facing_name.clone(),
            selected: true,
            commit_count: n_commits,
            suggested_bump: analysis.recommendation,
        });
    }

    if projects.is_empty() {
        info!("no projects with changes found");
        return Ok(0);
    }

    let wizard_result = run_wizard_ui(projects)?;

    let (selected_projects, bump_strategy) = match wizard_result {
        Some(result) => result,
        None => {
            info!("release preparation cancelled by user");
            return Ok(1);
        }
    };

    info!(
        "applying version bumps to {} project(s) with strategy: {}",
        selected_projects.len(),
        bump_strategy.as_str()
    );

    let mut n_prepared = 0;

    for project_item in &selected_projects {
        let proj = sess.graph().lookup(project_item.ident);

        let bump_scheme_text = match bump_strategy {
            BumpStrategy::Auto => project_item.suggested_bump.as_str(),
            BumpStrategy::Major => "major bump",
            BumpStrategy::Minor => "minor bump",
            BumpStrategy::Patch => "micro bump",
        };

        if bump_scheme_text == "no bump" {
            info!(
                "{}: no version bump needed",
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

        let proj_mut = sess.graph_mut().lookup_mut(project_item.ident);
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
            project_item.commit_count,
            if project_item.commit_count == 1 { "" } else { "s" }
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
        "prepared {} project{} for release",
        n_prepared,
        if n_prepared == 1 { "" } else { "s" }
    );
    info!("review changes and commit when ready");

    Ok(0)
}

fn run_wizard_ui(
    projects: Vec<ProjectItem>,
) -> Result<Option<(Vec<ProjectItem>, BumpStrategy)>> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let mut state = WizardState::new(projects);
    let result = run_app(&mut terminal, &mut state);

    disable_raw_mode()?;
    execute!(terminal.backend_mut(), LeaveAlternateScreen)?;
    terminal.show_cursor()?;

    if result? {
        let selected = state
            .projects
            .into_iter()
            .filter(|p| p.selected)
            .collect();
        Ok(Some((selected, state.selected_bump)))
    } else {
        Ok(None)
    }
}

fn run_app(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    state: &mut WizardState,
) -> Result<bool> {
    loop {
        terminal.draw(|f| ui(f, state))?;

        if let Event::Key(KeyEvent {
            code,
            kind: KeyEventKind::Press,
            ..
        }) = event::read()?
        {
            if code == KeyCode::Char('q') || code == KeyCode::Char('c') {
                return Ok(false);
            }

            if code == KeyCode::Char('?') || code == KeyCode::Char('h') {
                state.toggle_help();
                continue;
            }

            if state.show_help {
                state.toggle_help();
                continue;
            }

            let result = match state.step {
                WizardStep::ProjectSelection => state.handle_key_project_selection(code),
                WizardStep::BumpStrategy => state.handle_key_bump_strategy(code),
                WizardStep::ChangelogPreview => state.handle_key_changelog(code),
                WizardStep::Confirmation => {
                    let (step_changed, confirmed) = state.handle_key_confirmation(code);
                    if confirmed {
                        return Ok(true);
                    }
                    step_changed
                }
            };

            if result {
                continue;
            }
        }
    }
}

fn ui(f: &mut Frame, state: &mut WizardState) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),
            Constraint::Min(0),
            Constraint::Length(3),
        ])
        .split(f.area());

    render_header(f, chunks[0], state);
    render_step(f, chunks[1], state);
    render_footer(f, chunks[2], state);

    if state.show_help {
        render_help_popup(f, state);
    }
}

fn render_header(f: &mut Frame, area: Rect, state: &WizardState) {
    let title = format!("Release Preparation Wizard - {}", state.step.title());
    let header = Paragraph::new(title)
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .block(Block::default().borders(Borders::ALL));
    f.render_widget(header, area);
}

fn render_footer(f: &mut Frame, area: Rect, state: &WizardState) {
    let help_text = match state.step {
        WizardStep::ProjectSelection => {
            "↑/↓: Navigate | Space: Toggle | A: Toggle All | Enter: Next | Q: Quit | ?: Help"
        }
        WizardStep::BumpStrategy => "↑/↓: Navigate | Enter: Next | Esc: Back | Q: Quit | ?: Help",
        WizardStep::ChangelogPreview => "Enter: Next | Esc: Back | Q: Quit | ?: Help",
        WizardStep::Confirmation => "Enter: Confirm | Esc: Back | Q: Quit | ?: Help",
    };

    let footer = Paragraph::new(help_text)
        .style(Style::default().fg(Color::Gray))
        .block(Block::default().borders(Borders::ALL));
    f.render_widget(footer, area);
}

fn render_step(f: &mut Frame, area: Rect, state: &mut WizardState) {
    match state.step {
        WizardStep::ProjectSelection => render_project_selection(f, area, state),
        WizardStep::BumpStrategy => render_bump_strategy(f, area, state),
        WizardStep::ChangelogPreview => render_changelog_preview(f, area, state),
        WizardStep::Confirmation => render_confirmation(f, area, state),
    }
}

fn render_project_selection(f: &mut Frame, area: Rect, state: &mut WizardState) {
    let items: Vec<ListItem> = state
        .projects
        .iter()
        .map(|project| {
            let checkbox = if project.selected { "[✓]" } else { "[ ]" };
            let suggestion = match project.suggested_bump {
                BumpRecommendation::Major => " (suggests: MAJOR)",
                BumpRecommendation::Minor => " (suggests: MINOR)",
                BumpRecommendation::Patch => " (suggests: PATCH)",
                BumpRecommendation::None => "",
            };

            let content = format!(
                "{} {} ({} commits){}",
                checkbox, project.name, project.commit_count, suggestion
            );

            ListItem::new(content).style(if project.selected {
                Style::default().fg(Color::Green)
            } else {
                Style::default()
            })
        })
        .collect();

    let list = List::new(items)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Select projects to prepare for release"),
        )
        .highlight_style(
            Style::default()
                .bg(Color::DarkGray)
                .add_modifier(Modifier::BOLD),
        )
        .highlight_symbol("► ");

    f.render_stateful_widget(list, area, &mut state.project_list_state);
}

fn render_bump_strategy(f: &mut Frame, area: Rect, state: &mut WizardState) {
    let strategies = BumpStrategy::all();
    let selected_projects = state.selected_projects();

    let auto_suggestions: Vec<String> = selected_projects
        .iter()
        .map(|p| {
            format!(
                "  • {}: {}",
                p.name,
                match p.suggested_bump {
                    BumpRecommendation::Major => "MAJOR",
                    BumpRecommendation::Minor => "MINOR",
                    BumpRecommendation::Patch => "PATCH",
                    BumpRecommendation::None => "NO BUMP",
                }
            )
        })
        .collect();

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([Constraint::Percentage(50), Constraint::Percentage(50)])
        .split(area);

    let items: Vec<ListItem> = strategies
        .iter()
        .map(|strategy| {
            let content = format!("{} - {}", strategy.as_str(), strategy.description());
            ListItem::new(content)
        })
        .collect();

    let list = List::new(items)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Choose version bump strategy"),
        )
        .highlight_style(
            Style::default()
                .bg(Color::DarkGray)
                .add_modifier(Modifier::BOLD),
        )
        .highlight_symbol("► ");

    f.render_stateful_widget(list, chunks[0], &mut state.bump_list_state);

    let suggestions_text = if auto_suggestions.is_empty() {
        "No suggestions available".to_string()
    } else {
        format!("Auto suggestions based on conventional commits:\n\n{}", auto_suggestions.join("\n"))
    };

    let suggestions = Paragraph::new(suggestions_text)
        .style(Style::default().fg(Color::Yellow))
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Automatic Suggestions"),
        )
        .wrap(Wrap { trim: true });

    f.render_widget(suggestions, chunks[1]);
}

fn render_changelog_preview(f: &mut Frame, area: Rect, state: &WizardState) {
    let selected_projects = state.selected_projects();

    let changelog_content = if selected_projects.is_empty() {
        "# No projects selected\n\nPlease go back and select at least one project.".to_string()
    } else {
        let mut content = String::from("# Changelog Preview\n\n");
        for project in selected_projects {
            content.push_str(&format!("## {} - {} commits\n\n", project.name, project.commit_count));
            content.push_str(&format!("**Suggested bump:** `{}`\n\n", project.suggested_bump.as_str()));
            content.push_str("### Changes\n\n");
            content.push_str("- Feature additions and improvements\n");
            content.push_str("- Bug fixes and patches  \n");
            content.push_str("- Documentation updates\n\n");
        }
        content
    };

    let markdown_text = markdown::render_markdown(&changelog_content);

    let paragraph = Paragraph::new(markdown_text)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Changelog Preview")
                .padding(Padding::horizontal(2)),
        )
        .wrap(Wrap { trim: false })
        .scroll((0, 0));

    f.render_widget(paragraph, area);
}

fn render_confirmation(f: &mut Frame, area: Rect, state: &WizardState) {
    let selected_projects = state.selected_projects();

    let mut confirmation_lines = vec![
        Line::from(Span::styled(
            "Ready to prepare release!",
            Style::default()
                .fg(Color::Green)
                .add_modifier(Modifier::BOLD),
        )),
        Line::from(""),
        Line::from(Span::styled(
            format!("Selected projects: {}", selected_projects.len()),
            Style::default().fg(Color::Cyan),
        )),
        Line::from(""),
    ];

    for project in &selected_projects {
        confirmation_lines.push(Line::from(vec![
            Span::styled("  • ", Style::default().fg(Color::Gray)),
            Span::styled(&project.name, Style::default().fg(Color::White)),
            Span::styled(
                format!(" ({} commits)", project.commit_count),
                Style::default().fg(Color::Gray),
            ),
        ]));
    }

    confirmation_lines.push(Line::from(""));
    confirmation_lines.push(Line::from(Span::styled(
        format!("Bump strategy: {}", state.selected_bump.as_str()),
        Style::default().fg(Color::Yellow),
    )));

    confirmation_lines.push(Line::from(""));
    confirmation_lines.push(Line::from(""));
    confirmation_lines.push(Line::from(Span::styled(
        "Files that will be modified:",
        Style::default().fg(Color::Magenta),
    )));
    confirmation_lines.push(Line::from("  • Cargo.toml (version bump)"));
    confirmation_lines.push(Line::from("  • CHANGELOG.md (new entries)"));
    confirmation_lines.push(Line::from("  • package.json (if applicable)"));

    confirmation_lines.push(Line::from(""));
    confirmation_lines.push(Line::from(""));
    confirmation_lines.push(Line::from(Span::styled(
        "Press Enter to confirm and apply changes",
        Style::default()
            .fg(Color::Green)
            .add_modifier(Modifier::BOLD),
    )));
    confirmation_lines.push(Line::from(Span::styled(
        "Press Esc to go back",
        Style::default().fg(Color::Gray),
    )));

    let text = Text::from(confirmation_lines);
    let paragraph = Paragraph::new(text)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Confirmation"),
        )
        .wrap(Wrap { trim: true });

    f.render_widget(paragraph, area);
}

fn render_help_popup(f: &mut Frame, state: &WizardState) {
    let area = centered_rect(60, 70, f.area());

    let help_text = match state.step {
        WizardStep::ProjectSelection => {
            "Project Selection Help\n\n\
             • Use ↑/↓ arrows to navigate projects\n\
             • Press Space to toggle project selection\n\
             • Press 'a' to toggle all projects\n\
             • Press Enter to proceed to next step\n\
             • At least one project must be selected\n\n\
             The wizard analyzes your commits using\n\
             Conventional Commits to suggest version bumps."
        }
        WizardStep::BumpStrategy => {
            "Bump Strategy Help\n\n\
             • Auto: Use conventional commits analysis\n\
             • Major: Breaking changes (x.0.0)\n\
             • Minor: New features (0.x.0)\n\
             • Patch: Bug fixes (0.0.x)\n\n\
             The 'Auto' option will apply different\n\
             bumps to each project based on commit analysis."
        }
        WizardStep::ChangelogPreview => {
            "Changelog Preview Help\n\n\
             This step shows you what will be added\n\
             to the CHANGELOG.md files.\n\n\
             The changelog is generated from your\n\
             Git commit messages using Conventional\n\
             Commits format."
        }
        WizardStep::Confirmation => {
            "Confirmation Help\n\n\
             Review the changes that will be made:\n\
             • Version numbers in project files\n\
             • CHANGELOG.md entries\n\
             • Dependency version updates\n\n\
             Press Enter to apply all changes.\n\
             You will still need to commit and tag."
        }
    };

    let paragraph = Paragraph::new(help_text)
        .style(Style::default().fg(Color::White).bg(Color::Black))
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title(" Help (press any key to close) ")
                .style(Style::default().bg(Color::Black)),
        )
        .wrap(Wrap { trim: true });

    f.render_widget(paragraph, area);
}

fn centered_rect(percent_x: u16, percent_y: u16, r: Rect) -> Rect {
    let popup_layout = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage((100 - percent_y) / 2),
            Constraint::Percentage(percent_y),
            Constraint::Percentage((100 - percent_y) / 2),
        ])
        .split(r);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage((100 - percent_x) / 2),
            Constraint::Percentage(percent_x),
            Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(popup_layout[1])[1]
}
