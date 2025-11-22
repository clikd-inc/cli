use std::io::IsTerminal;

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
