use crossterm::event::KeyCode;
use ratatui::{
    layout::{Alignment, Constraint, Direction, Layout},
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, BorderType, Borders, Clear, Paragraph},
    Frame,
};

use crate::core::ui::{
    theme::AppColors,
    utils::{centered_rect, BoxLocation},
};

pub(crate) struct ConfirmDialog<'a> {
    title: &'a str,
    message: &'a str,
    highlighted_text: Option<&'a str>,
    yes_key: KeyCode,
    no_key: KeyCode,
    colors: AppColors,
}

impl<'a> ConfirmDialog<'a> {
    pub fn new(title: &'a str, message: &'a str, colors: AppColors) -> Self {
        Self {
            title,
            message,
            highlighted_text: None,
            yes_key: KeyCode::Char('y'),
            no_key: KeyCode::Char('n'),
            colors,
        }
    }

    pub fn highlighted_text(mut self, text: &'a str) -> Self {
        self.highlighted_text = Some(text);
        self
    }

    pub fn yes_key(mut self, key: KeyCode) -> Self {
        self.yes_key = key;
        self
    }

    pub fn no_key(mut self, key: KeyCode) -> Self {
        self.no_key = key;
        self
    }

    pub fn render(&self, frame: &mut Frame) {
        let block = Block::default()
            .title(format!(" {} ", self.title))
            .border_type(BorderType::Rounded)
            .style(
                Style::default()
                    .bg(self.colors.popup_delete.background)
                    .fg(self.colors.popup_delete.text),
            )
            .title_alignment(Alignment::Center)
            .borders(Borders::ALL);

        let message_line = if let Some(highlighted) = self.highlighted_text {
            Line::from(vec![
                Span::from(self.message),
                Span::styled(
                    highlighted,
                    Style::default()
                        .fg(self.colors.popup_delete.text_highlight)
                        .bg(self.colors.popup_delete.background)
                        .add_modifier(Modifier::BOLD),
                ),
            ])
        } else {
            Line::from(self.message)
        };

        let yes_text = format!("( {} ) yes", self.format_key(self.yes_key));
        let no_text = format!("( {} ) no", self.format_key(self.no_key));

        let max_line_width = u16::try_from(message_line.width()).unwrap_or(64) + 12;
        let lines = 8;

        let message_para = Paragraph::new(message_line).alignment(Alignment::Center);

        let button_block = || {
            Block::default()
                .border_type(BorderType::Rounded)
                .borders(Borders::ALL)
                .style(Style::default().bg(self.colors.popup_delete.background))
        };

        let yes_para = Paragraph::new(yes_text)
            .alignment(Alignment::Center)
            .block(button_block());

        let no_para = Paragraph::new(no_text)
            .alignment(Alignment::Center)
            .block(button_block());

        let area = centered_rect(
            lines,
            max_line_width.into(),
            frame.area(),
            BoxLocation::MiddleCenter,
        );

        let split_popup = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(1),
                Constraint::Length(3),
                Constraint::Length(1),
                Constraint::Length(3),
                Constraint::Length(1),
            ])
            .split(area);

        let split_buttons = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(10),
                Constraint::Percentage(35),
                Constraint::Percentage(10),
                Constraint::Percentage(35),
                Constraint::Percentage(10),
            ])
            .split(split_popup[3]);

        frame.render_widget(Clear, area);
        frame.render_widget(block, area);
        frame.render_widget(message_para, split_popup[1]);
        frame.render_widget(no_para, split_buttons[1]);
        frame.render_widget(yes_para, split_buttons[3]);
    }

    fn format_key(&self, key: KeyCode) -> String {
        match key {
            KeyCode::Char(c) => c.to_string(),
            KeyCode::F(n) => format!("F{}", n),
            KeyCode::Enter => "Enter".to_string(),
            KeyCode::Esc => "Esc".to_string(),
            KeyCode::Backspace => "Backspace".to_string(),
            KeyCode::Left => "←".to_string(),
            KeyCode::Right => "→".to_string(),
            KeyCode::Up => "↑".to_string(),
            KeyCode::Down => "↓".to_string(),
            _ => "?".to_string(),
        }
    }
}
