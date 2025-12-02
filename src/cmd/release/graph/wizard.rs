use anyhow::Result;
use crossterm::{
    event::{self, Event, KeyCode, KeyEventKind},
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
    ExecutableCommand,
};
use ratatui::{
    prelude::*,
    widgets::{Block, Borders, List, ListItem, ListState, Paragraph, Wrap},
};
use std::io::stdout;

use crate::{
    atry,
    core::release::{graph::GraphQueryBuilder, session::AppSession},
};

struct ProjectInfo {
    name: String,
    version: String,
    deps: Vec<DependencyInfo>,
    dependents: Vec<String>,
}

struct DependencyInfo {
    name: String,
    version: String,
}

struct App {
    projects: Vec<ProjectInfo>,
    release_order: Vec<String>,
    list_state: ListState,
    focus: Focus,
}

#[derive(Clone, Copy, PartialEq)]
enum Focus {
    ProjectList,
    Details,
}

impl App {
    fn new(sess: &AppSession, idents: &[usize]) -> Self {
        let mut projects = Vec::new();

        for &ident in idents {
            let proj = sess.graph().lookup(ident);
            let deps: Vec<DependencyInfo> = proj
                .internal_deps
                .iter()
                .map(|d| {
                    let dep_proj = sess.graph().lookup(d.ident);
                    DependencyInfo {
                        name: dep_proj.user_facing_name.clone(),
                        version: dep_proj.version.to_string(),
                    }
                })
                .collect();

            let dependents: Vec<String> = idents
                .iter()
                .filter_map(|&other_ident| {
                    if other_ident == ident {
                        return None;
                    }
                    let other_proj = sess.graph().lookup(other_ident);
                    if other_proj.internal_deps.iter().any(|d| d.ident == ident) {
                        Some(other_proj.user_facing_name.clone())
                    } else {
                        None
                    }
                })
                .collect();

            projects.push(ProjectInfo {
                name: proj.user_facing_name.clone(),
                version: proj.version.to_string(),
                deps,
                dependents,
            });
        }

        let release_order: Vec<String> = sess
            .graph()
            .toposorted()
            .map(|id| sess.graph().lookup(id).user_facing_name.clone())
            .collect();

        let mut list_state = ListState::default();
        if !projects.is_empty() {
            list_state.select(Some(0));
        }

        Self {
            projects,
            release_order,
            list_state,
            focus: Focus::ProjectList,
        }
    }

    fn selected_project(&self) -> Option<&ProjectInfo> {
        self.list_state.selected().and_then(|i| self.projects.get(i))
    }

    fn next(&mut self) {
        if self.projects.is_empty() {
            return;
        }
        let i = match self.list_state.selected() {
            Some(i) => (i + 1) % self.projects.len(),
            None => 0,
        };
        self.list_state.select(Some(i));
    }

    fn previous(&mut self) {
        if self.projects.is_empty() {
            return;
        }
        let i = match self.list_state.selected() {
            Some(i) => {
                if i == 0 {
                    self.projects.len() - 1
                } else {
                    i - 1
                }
            }
            None => 0,
        };
        self.list_state.select(Some(i));
    }

    fn toggle_focus(&mut self) {
        self.focus = match self.focus {
            Focus::ProjectList => Focus::Details,
            Focus::Details => Focus::ProjectList,
        };
    }
}

pub fn run() -> Result<i32> {
    let sess = atry!(
        AppSession::initialize_default();
        ["could not initialize app and project graph"]
    );

    let q = GraphQueryBuilder::default();
    let idents = sess
        .graph()
        .query(q)
        .map_err(|e| anyhow::anyhow!("could not select projects: {}", e))?;

    if idents.is_empty() {
        println!("No projects found in repository");
        return Ok(0);
    }

    let mut app = App::new(&sess, &idents);

    enable_raw_mode()?;
    stdout().execute(EnterAlternateScreen)?;
    let mut terminal = Terminal::new(CrosstermBackend::new(stdout()))?;

    let result = run_app(&mut terminal, &mut app);

    disable_raw_mode()?;
    stdout().execute(LeaveAlternateScreen)?;

    result
}

fn run_app<B: Backend>(terminal: &mut Terminal<B>, app: &mut App) -> Result<i32> {
    loop {
        terminal.draw(|f| ui(f, app))?;

        if let Event::Key(key) = event::read()? {
            if key.kind != KeyEventKind::Press {
                continue;
            }

            match key.code {
                KeyCode::Char('q') | KeyCode::Esc => return Ok(0),
                KeyCode::Down | KeyCode::Char('j') => {
                    if app.focus == Focus::ProjectList {
                        app.next();
                    }
                }
                KeyCode::Up | KeyCode::Char('k') => {
                    if app.focus == Focus::ProjectList {
                        app.previous();
                    }
                }
                KeyCode::Tab => app.toggle_focus(),
                _ => {}
            }
        }
    }
}

fn ui(f: &mut Frame, app: &mut App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3),
            Constraint::Min(10),
            Constraint::Length(3),
        ])
        .split(f.area());

    let title = Paragraph::new(" Dependency Graph ")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .block(
            Block::default()
                .borders(Borders::ALL)
                .border_style(Style::default().fg(Color::Cyan)),
        );
    f.render_widget(title, chunks[0]);

    let main_chunks = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([Constraint::Percentage(40), Constraint::Percentage(60)])
        .split(chunks[1]);

    render_project_list(f, app, main_chunks[0]);
    render_details(f, app, main_chunks[1]);

    let help = Paragraph::new(" ↑↓/jk: Navigate | Tab: Switch Panel | q/Esc: Quit ")
        .style(Style::default().fg(Color::DarkGray))
        .block(Block::default().borders(Borders::ALL));
    f.render_widget(help, chunks[2]);
}

fn render_project_list(f: &mut Frame, app: &mut App, area: Rect) {
    let items: Vec<ListItem> = app
        .projects
        .iter()
        .map(|p| {
            let symbol = if p.deps.is_empty() { "○" } else { "●" };
            let content = format!("{} {} @ {}", symbol, p.name, p.version);
            ListItem::new(content)
        })
        .collect();

    let border_style = if app.focus == Focus::ProjectList {
        Style::default().fg(Color::Yellow)
    } else {
        Style::default().fg(Color::White)
    };

    let list = List::new(items)
        .block(
            Block::default()
                .title(" Projects ")
                .borders(Borders::ALL)
                .border_style(border_style),
        )
        .highlight_style(
            Style::default()
                .bg(Color::DarkGray)
                .add_modifier(Modifier::BOLD),
        )
        .highlight_symbol("▶ ");

    f.render_stateful_widget(list, area, &mut app.list_state);
}

fn render_details(f: &mut Frame, app: &mut App, area: Rect) {
    let border_style = if app.focus == Focus::Details {
        Style::default().fg(Color::Yellow)
    } else {
        Style::default().fg(Color::White)
    };

    let block = Block::default()
        .title(" Details ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let inner = block.inner(area);
    f.render_widget(block, area);

    let Some(proj) = app.selected_project() else {
        let no_selection = Paragraph::new("No project selected")
            .style(Style::default().fg(Color::DarkGray));
        f.render_widget(no_selection, inner);
        return;
    };

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(4),
            Constraint::Length(1),
            Constraint::Min(4),
            Constraint::Length(1),
            Constraint::Min(4),
            Constraint::Length(1),
            Constraint::Min(4),
        ])
        .split(inner);

    let info_text = format!(
        "Name: {}\nVersion: {}\nDependencies: {} | Dependents: {}",
        proj.name,
        proj.version,
        proj.deps.len(),
        proj.dependents.len()
    );
    let info = Paragraph::new(info_text)
        .style(Style::default().fg(Color::White))
        .wrap(Wrap { trim: true });
    f.render_widget(info, chunks[0]);

    let deps_title = Paragraph::new("Dependencies:")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD));
    f.render_widget(deps_title, chunks[1]);

    if proj.deps.is_empty() {
        let no_deps = Paragraph::new("  (none)")
            .style(Style::default().fg(Color::DarkGray));
        f.render_widget(no_deps, chunks[2]);
    } else {
        let deps_text: String = proj
            .deps
            .iter()
            .map(|d| format!("  → {} @ {}", d.name, d.version))
            .collect::<Vec<_>>()
            .join("\n");
        let deps = Paragraph::new(deps_text)
            .style(Style::default().fg(Color::Green))
            .wrap(Wrap { trim: true });
        f.render_widget(deps, chunks[2]);
    }

    let dependents_title = Paragraph::new("Depended on by:")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD));
    f.render_widget(dependents_title, chunks[3]);

    if proj.dependents.is_empty() {
        let no_dependents = Paragraph::new("  (none)")
            .style(Style::default().fg(Color::DarkGray));
        f.render_widget(no_dependents, chunks[4]);
    } else {
        let dependents_text: String = proj
            .dependents
            .iter()
            .map(|d| format!("  ← {}", d))
            .collect::<Vec<_>>()
            .join("\n");
        let dependents = Paragraph::new(dependents_text)
            .style(Style::default().fg(Color::Yellow))
            .wrap(Wrap { trim: true });
        f.render_widget(dependents, chunks[4]);
    }

    let order_title = Paragraph::new("Release Order:")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD));
    f.render_widget(order_title, chunks[5]);

    let order_text: String = app
        .release_order
        .iter()
        .enumerate()
        .map(|(i, name)| {
            let marker = if name == &proj.name { "▶" } else { " " };
            format!("{} {}. {}", marker, i + 1, name)
        })
        .collect::<Vec<_>>()
        .join("\n");
    let order = Paragraph::new(order_text)
        .style(Style::default().fg(Color::White))
        .wrap(Wrap { trim: true });
    f.render_widget(order, chunks[6]);
}
