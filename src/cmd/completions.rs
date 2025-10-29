use clap_complete::{generate as gen_completions, Shell};
use crate::cli::Cli;
use clap::CommandFactory;

pub fn generate(shell: Shell) {
    let mut cmd = Cli::command();
    let bin_name = cmd.get_name().to_string();
    gen_completions(shell, &mut cmd, bin_name, &mut std::io::stdout());
}
