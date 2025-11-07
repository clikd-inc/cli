use kdam::{term, tqdm, BarExt, Column, RichProgress, Spinner as KdamSpinner};
use owo_colors::{OwoColorize, Rgb};
use spinoff::{spinners, Color as SpinoffColor, Spinner};

pub const ICON_SUCCESS: &str = "✓";
pub const ICON_ERROR: &str = "✗";
pub const ICON_WARNING: &str = "⚠";
pub const ICON_INFO: &str = "ℹ";
pub const ICON_ARROW: &str = "▶";
pub const ICON_POINTER: &str = "→";
pub const SEPARATOR: &str = "─";
pub const LOGO: &str = "🐱";

pub fn primary() -> Rgb {
    Rgb(114, 227, 173)
}

pub fn error() -> Rgb {
    Rgb(202, 50, 20)
}

pub fn warning() -> Rgb {
    Rgb(245, 158, 11)
}

pub fn info() -> Rgb {
    Rgb(59, 130, 246)
}

pub fn success() -> Rgb {
    Rgb(114, 227, 173)
}

pub fn primary_spinoff() -> SpinoffColor {
    SpinoffColor::TrueColor {
        r: 114,
        g: 227,
        b: 173,
    }
}

pub struct DockerProgressBar {
    progress: RichProgress,
}

impl Default for DockerProgressBar {
    fn default() -> Self {
        Self::new()
    }
}

impl DockerProgressBar {
    pub fn new() -> Self {
        use std::io::{stderr, IsTerminal};

        term::init(stderr().is_terminal());

        let bar = tqdm!(
            total = 100,
            unit_scale = true,
            unit_divisor = 1024,
            unit = "B",
            animation = "fillup",
            ncols = 50
        );

        let progress = RichProgress::new(
            bar,
            vec![
                Column::Spinner(KdamSpinner::new(&["◜", "◠", "◝", "◞", "◡", "◟"], 80.0, 1.0)),
                Column::Text("".to_owned()),
                Column::Animation,
                Column::Percentage(1),
                Column::Text("•".to_owned()),
                Column::CountTotal,
                Column::Text("•".to_owned()),
                Column::Rate,
                Column::Text("•".to_owned()),
                Column::RemainingTime,
            ],
        );

        Self { progress }
    }

    pub fn set_message(&mut self, msg: String) {
        self.progress.replace(1, Column::Text(msg));
    }

    pub fn set_length(&mut self, len: u64) {
        self.progress.pb.total = len as usize;
    }

    pub fn set_position(&mut self, pos: u64) {
        let _ = self.progress.update_to(pos as usize);
    }

    pub fn finish_and_clear(self) {
        drop(self.progress);
        println!();
    }
}

pub fn dimmed(text: &str) -> String {
    format!("{}", text.dimmed())
}

pub fn separator(width: usize) -> String {
    format!("{}", SEPARATOR.repeat(width).dimmed())
}

pub fn header(text: &str) -> String {
    format!(
        "\n{} {}\n{}",
        LOGO,
        text.bold().color(primary()),
        separator(60)
    )
}

pub fn success_icon() -> String {
    format!("{}", ICON_SUCCESS.color(success()).bold())
}

pub fn error_icon() -> String {
    format!("{}", ICON_ERROR.color(error()).bold())
}

pub fn warning_icon() -> String {
    format!("{}", ICON_WARNING.color(warning()).bold())
}

pub fn info_icon() -> String {
    format!("{}", ICON_INFO.color(info()).bold())
}

pub fn arrow_icon() -> String {
    format!("{}", ICON_ARROW.color(primary()))
}

pub fn pointer_icon() -> String {
    format!("{}", ICON_POINTER.color(primary()))
}

pub fn success_message(msg: &str) -> String {
    format!("{} {}", success_icon(), msg)
}

pub fn error_message(msg: &str) -> String {
    format!("{} {}", error_icon(), msg)
}

pub fn warning_message(msg: &str) -> String {
    format!("{} {}", warning_icon(), msg)
}

pub fn info_message(msg: &str) -> String {
    format!("{} {}", info_icon(), msg)
}

pub fn step_message(msg: &str) -> String {
    format!("{} {}", pointer_icon(), msg.dimmed())
}

pub fn highlight(text: &str) -> String {
    format!("{}", text.color(primary()).bold())
}

pub fn url(text: &str) -> String {
    format!("{}", text.color(info()).underline())
}

pub fn code(text: &str) -> String {
    format!("{}", text.color(warning()))
}

pub fn create_progress_bar() -> DockerProgressBar {
    DockerProgressBar::new()
}

pub fn create_spinner(msg: impl Into<String>) -> Spinner {
    Spinner::new(spinners::Arc, msg.into(), Some(primary_spinoff()))
}

pub fn format_docker_status(status: &str) -> String {
    if let Some(pos) = status.find(" for ") {
        let (prefix, suffix) = status.split_at(pos + 5);
        format!("{}{}", dimmed(prefix), highlight(suffix))
    } else {
        dimmed(status)
    }
}
