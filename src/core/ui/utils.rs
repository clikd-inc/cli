use std::io::IsTerminal;

use ratatui::layout::{Constraint, Direction, Layout, Rect};

const TTY_PATH: &str = "/dev/tty";

pub(crate) fn is_tty() -> bool {
    std::fs::OpenOptions::new()
        .read(true)
        .write(false)
        .open(TTY_PATH)
        .is_ok()
}

pub(crate) fn is_interactive_terminal() -> bool {
    std::io::stdout().is_terminal() && std::io::stdin().is_terminal()
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub(crate) enum OutputMode {
    Tui,
    Text,
    Json,
}

impl OutputMode {
    pub(crate) fn detect() -> Self {
        if is_interactive_terminal() {
            Self::Tui
        } else {
            Self::Text
        }
    }

    pub(crate) fn from_format_and_tty(format: &str) -> Self {
        match format {
            "json" => Self::Json,
            "table" => {
                if is_interactive_terminal() {
                    Self::Tui
                } else {
                    Self::Text
                }
            }
            _ => Self::Text,
        }
    }
}

#[derive(Debug, Clone, Copy)]
pub(crate) enum BoxLocation {
    TopLeft,
    TopCenter,
    TopRight,
    MiddleLeft,
    MiddleCenter,
    MiddleRight,
    BottomLeft,
    BottomCenter,
    BottomRight,
}

impl BoxLocation {
    pub(crate) const fn get_indexes(self) -> (usize, usize) {
        match self {
            Self::TopLeft => (0, 0),
            Self::TopCenter => (0, 1),
            Self::TopRight => (0, 2),
            Self::MiddleLeft => (1, 0),
            Self::MiddleCenter => (1, 1),
            Self::MiddleRight => (1, 2),
            Self::BottomLeft => (2, 0),
            Self::BottomCenter => (2, 1),
            Self::BottomRight => (2, 2),
        }
    }

    pub(crate) const fn get_constraints(
        self,
        blank_horizontal: u16,
        blank_vertical: u16,
        text_lines: u16,
        text_width: u16,
    ) -> ([Constraint; 3], [Constraint; 3]) {
        let h = match self {
            Self::TopLeft | Self::MiddleLeft | Self::BottomLeft => [
                Constraint::Length(text_width),
                Constraint::Min(1),
                Constraint::Length(0),
            ],
            Self::TopCenter | Self::MiddleCenter | Self::BottomCenter => [
                Constraint::Length(blank_horizontal),
                Constraint::Length(text_width),
                Constraint::Length(blank_horizontal),
            ],
            Self::TopRight | Self::MiddleRight | Self::BottomRight => [
                Constraint::Length(0),
                Constraint::Min(1),
                Constraint::Length(text_width),
            ],
        };

        let v = match self {
            Self::TopLeft | Self::TopCenter | Self::TopRight => [
                Constraint::Length(text_lines),
                Constraint::Min(1),
                Constraint::Length(0),
            ],
            Self::MiddleLeft | Self::MiddleCenter | Self::MiddleRight => [
                Constraint::Length(blank_vertical),
                Constraint::Length(text_lines),
                Constraint::Length(blank_vertical),
            ],
            Self::BottomLeft | Self::BottomCenter | Self::BottomRight => [
                Constraint::Length(0),
                Constraint::Min(1),
                Constraint::Length(text_lines),
            ],
        };

        (h, v)
    }
}

pub(crate) fn centered_rect(
    text_lines: usize,
    text_width: usize,
    area: Rect,
    location: BoxLocation,
) -> Rect {
    let calc = |x: u16, y: usize| usize::from(x).saturating_sub(y).saturating_div(2);

    let blank_vertical = calc(area.height, text_lines);
    let blank_horizontal = calc(area.width, text_width);

    let (h_constraints, v_constraints) = location.get_constraints(
        blank_horizontal.try_into().unwrap_or_default(),
        blank_vertical.try_into().unwrap_or_default(),
        text_lines.try_into().unwrap_or_default(),
        text_width.try_into().unwrap_or_default(),
    );

    let indexes = location.get_indexes();

    let popup_layout = Layout::default()
        .direction(Direction::Vertical)
        .constraints(v_constraints)
        .split(area);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints(h_constraints)
        .split(popup_layout[indexes.0])[indexes.1]
}
