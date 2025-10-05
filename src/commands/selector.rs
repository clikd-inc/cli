use anyhow::Result;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::CrosstermBackend,
    layout::{Constraint, Direction, Layout},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, List, ListItem, Paragraph},
    Frame, Terminal,
};
use std::io;

#[derive(Clone)]
struct CommandItem {
    name: &'static str,
    description: &'static str,
    command: &'static str,
}

struct App {
    commands: Vec<CommandItem>,
    selected: usize,
    should_quit: bool,
    should_execute: Option<String>,
}

impl App {
    fn new() -> Self {
        let commands = vec![
            CommandItem {
                name: "start",
                description: "Start development services with interactive dashboard",
                command: "start",
            },
            CommandItem {
                name: "stop",
                description: "Stop running development services",
                command: "stop",
            },
            CommandItem {
                name: "status",
                description: "Monitor service status and health",
                command: "status",
            },
            CommandItem {
                name: "logs",
                description: "View and filter service logs in real-time",
                command: "logs",
            },
            CommandItem {
                name: "switch",
                description: "Switch between development environments",
                command: "switch",
            },
            CommandItem {
                name: "db",
                description: "Database management operations",
                command: "db",
            },
            CommandItem {
                name: "gen",
                description: "Generate client SDK code",
                command: "gen",
            },
            CommandItem {
                name: "deploy",
                description: "Deploy to target environment",
                command: "deploy",
            },
            CommandItem {
                name: "tui",
                description: "Launch unified TUI dashboard",
                command: "tui",
            },
        ];

        Self {
            commands,
            selected: 0,
            should_quit: false,
            should_execute: None,
        }
    }

    fn next(&mut self) {
        self.selected = (self.selected + 1) % self.commands.len();
    }

    fn previous(&mut self) {
        if self.selected > 0 {
            self.selected -= 1;
        } else {
            self.selected = self.commands.len() - 1;
        }
    }

    fn execute_selected(&mut self) {
        if let Some(cmd) = self.commands.get(self.selected) {
            self.should_execute = Some(cmd.command.to_string());
        }
    }

    fn quit(&mut self) {
        self.should_quit = true;
    }
}

pub async fn run_interactive() -> Result<()> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let mut app = App::new();
    let result = run_app(&mut terminal, &mut app);

    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    if let Err(err) = result {
        eprintln!("Error: {:?}", err);
        return Err(err);
    }

    if let Some(command) = app.should_execute {
        execute_command(&command).await?;
    }

    Ok(())
}

fn run_app<B: ratatui::backend::Backend>(
    terminal: &mut Terminal<B>,
    app: &mut App,
) -> Result<()> {
    loop {
        terminal.draw(|f| ui(f, app))?;

        if event::poll(std::time::Duration::from_millis(100))? {
            if let Event::Key(key) = event::read()? {
                match key.code {
                    KeyCode::Char('q') | KeyCode::Esc => app.quit(),
                    KeyCode::Down | KeyCode::Char('j') => app.next(),
                    KeyCode::Up | KeyCode::Char('k') => app.previous(),
                    KeyCode::Enter => {
                        app.execute_selected();
                        app.quit();
                    }
                    _ => {}
                }
            }
        }

        if app.should_quit {
            break;
        }
    }

    Ok(())
}

fn ui(f: &mut Frame, app: &App) {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .margin(2)
        .constraints([
            Constraint::Length(3),
            Constraint::Min(0),
            Constraint::Length(3),
        ])
        .split(f.area());

    let title = Paragraph::new("Clikd Development CLI")
        .style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
        .block(Block::default().borders(Borders::ALL));
    f.render_widget(title, chunks[0]);

    let items: Vec<ListItem> = app
        .commands
        .iter()
        .enumerate()
        .map(|(i, cmd)| {
            let is_selected = i == app.selected;
            let prefix = if is_selected { "▶ " } else { "  " };

            let style = if is_selected {
                Style::default()
                    .fg(Color::Black)
                    .bg(Color::Cyan)
                    .add_modifier(Modifier::BOLD)
            } else {
                Style::default().fg(Color::White)
            };

            let line = Line::from(vec![
                Span::styled(prefix, style),
                Span::styled(format!("{:<12}", cmd.name), style),
                Span::styled(cmd.description, style),
            ]);

            ListItem::new(line)
        })
        .collect();

    let list = List::new(items)
        .block(Block::default().borders(Borders::ALL).title("Commands"));
    f.render_widget(list, chunks[1]);

    let help_text = "Navigation: ↑↓ or j/k  |  Select: Enter  |  Quit: q or Esc";
    let help = Paragraph::new(help_text)
        .style(Style::default().fg(Color::DarkGray))
        .block(Block::default().borders(Borders::ALL));
    f.render_widget(help, chunks[2]);
}

async fn execute_command(command: &str) -> Result<()> {
    match command {
        "start" => crate::commands::start::run_interactive(None).await,
        "stop" => crate::commands::stop::run_interactive(false).await,
        "status" => crate::commands::status::run_tui().await,
        "logs" => crate::commands::logs::run_tui(None).await,
        "switch" => crate::commands::switch::run_interactive().await,
        "db" => crate::commands::db::run_main_tui().await,
        "gen" => crate::commands::gen::run_main_tui().await,
        "deploy" => crate::commands::deploy::run_interactive(None).await,
        "tui" => crate::commands::tui::run_main_app().await,
        _ => {
            eprintln!("Unknown command: {}", command);
            Ok(())
        }
    }
}