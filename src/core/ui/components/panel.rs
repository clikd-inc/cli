use ratatui::{
    layout::Rect,
    style::Style,
    widgets::{Block, Borders, Paragraph},
    Frame,
};

pub(crate) struct Panel<'a> {
    title: &'a str,
    content: Paragraph<'a>,
    border_style: Style,
    selected: bool,
}

impl<'a> Panel<'a> {
    pub fn new(title: &'a str, content: Paragraph<'a>) -> Self {
        Self {
            title,
            content,
            border_style: Style::default(),
            selected: false,
        }
    }

    pub fn border_style(mut self, style: Style) -> Self {
        self.border_style = style;
        self
    }

    pub fn selected(mut self, selected: bool) -> Self {
        self.selected = selected;
        self
    }

    pub fn render(&self, frame: &mut Frame, area: Rect) {
        let block = Block::default()
            .title(self.title)
            .borders(Borders::ALL)
            .border_style(self.border_style);

        let paragraph = self.content.clone().block(block);
        frame.render_widget(paragraph, area);
    }
}
