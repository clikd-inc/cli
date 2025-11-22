use ratatui::{
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    style::Style,
    widgets::{Block, Borders, Clear, Paragraph, Wrap},
    Frame,
};

pub(crate) struct Popup<'a> {
    title: &'a str,
    content: &'a str,
    width_percent: u16,
    height_percent: u16,
    style: Style,
}

impl<'a> Popup<'a> {
    pub fn new(title: &'a str, content: &'a str) -> Self {
        Self {
            title,
            content,
            width_percent: 50,
            height_percent: 50,
            style: Style::default(),
        }
    }

    pub fn width_percent(mut self, percent: u16) -> Self {
        self.width_percent = percent;
        self
    }

    pub fn height_percent(mut self, percent: u16) -> Self {
        self.height_percent = percent;
        self
    }

    pub fn style(mut self, style: Style) -> Self {
        self.style = style;
        self
    }

    pub fn render(&self, frame: &mut Frame, area: Rect) {
        let popup_area = self.centered_rect(area);

        frame.render_widget(Clear, popup_area);

        let block = Block::default()
            .title(self.title)
            .borders(Borders::ALL)
            .style(self.style);

        let paragraph = Paragraph::new(self.content)
            .block(block)
            .wrap(Wrap { trim: true })
            .alignment(Alignment::Left);

        frame.render_widget(paragraph, popup_area);
    }

    fn centered_rect(&self, r: Rect) -> Rect {
        let popup_layout = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage((100 - self.height_percent) / 2),
                Constraint::Percentage(self.height_percent),
                Constraint::Percentage((100 - self.height_percent) / 2),
            ])
            .split(r);

        Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage((100 - self.width_percent) / 2),
                Constraint::Percentage(self.width_percent),
                Constraint::Percentage((100 - self.width_percent) / 2),
            ])
            .split(popup_layout[1])[1]
    }
}
