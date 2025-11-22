use std::time::{Duration, Instant};

use ratatui::{
    layout::Alignment,
    style::Style,
    widgets::{Block, Borders, Clear, Paragraph},
    Frame,
};

use crate::core::ui::{
    theme::AppColors,
    utils::{centered_rect, BoxLocation},
};

pub(crate) struct Toast<'a> {
    message: &'a str,
    location: BoxLocation,
    colors: AppColors,
    created_at: Instant,
    duration: Duration,
}

impl<'a> Toast<'a> {
    pub fn new(message: &'a str, colors: AppColors) -> Self {
        Self {
            message,
            location: BoxLocation::BottomRight,
            colors,
            created_at: Instant::now(),
            duration: Duration::from_secs(4),
        }
    }

    pub fn location(mut self, location: BoxLocation) -> Self {
        self.location = location;
        self
    }

    pub fn duration(mut self, duration: Duration) -> Self {
        self.duration = duration;
        self
    }

    pub fn is_expired(&self) -> bool {
        self.created_at.elapsed() > self.duration
    }

    pub fn render(&self, frame: &mut Frame) {
        if self.is_expired() {
            return;
        }

        let block = Block::default()
            .title("")
            .title_alignment(Alignment::Center)
            .style(
                Style::default()
                    .bg(self.colors.popup_info.background)
                    .fg(self.colors.popup_info.text),
            )
            .borders(Borders::NONE);

        let max_line_width = self.message.lines().map(|l| l.len()).max().unwrap_or(0) + 8;
        let lines = self.message.lines().count() + 2;

        let paragraph = Paragraph::new(self.message)
            .block(block)
            .style(
                Style::default()
                    .bg(self.colors.popup_info.background)
                    .fg(self.colors.popup_info.text),
            )
            .alignment(Alignment::Center);

        let area = centered_rect(lines, max_line_width, frame.area(), self.location);
        frame.render_widget(Clear, area);
        frame.render_widget(paragraph, area);
    }
}
