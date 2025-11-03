use crate::cli::InitArgs;
use crate::core::git::{branch, gitignore};
use crate::core::ide::{intellij, vscode};
use crate::error::{CliError, Result};
use crate::utils::theme::*;
use dialoguer::Confirm;
use std::env;
use std::fs;

const CONFIG_TEMPLATE: &str = include_str!("../../templates/config.toml");

pub async fn run(args: InitArgs) -> Result<()> {
    println!("{}", header("Initializing Clikd"));

    let project_root = args.workdir.unwrap_or_else(|| env::current_dir().unwrap());

    let config_path = project_root.join("clikd/config.toml");
    if config_path.exists() {
        return Err(CliError::AlreadyInitialized);
    }

    let project_name = project_root
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("clikd-project");

    let project_id = sanitize_project_id(project_name);

    println!("\n{}", step_message("Creating project structure..."));
    fs::create_dir_all(project_root.join("clikd"))?;
    fs::create_dir_all(project_root.join("clikd/.temp"))?;

    println!("{}", step_message("Generating configuration..."));
    let config = CONFIG_TEMPLATE.replace("{{project_id}}", &project_id);
    fs::write(&config_path, config)?;

    println!("{}", step_message("Initializing git branch..."));
    branch::init_current_branch()?;

    println!("{}", step_message("Updating .gitignore..."));
    gitignore::update(&project_root)?;

    if args.vscode {
        vscode::create_settings(&project_root)?;
        println!("{}", success_message(&format!("Generated VS Code settings in {}", highlight(".vscode/settings.json"))));
    } else if args.intellij {
        intellij::create_settings(&project_root)?;
        println!("{}", success_message(&format!("Generated IntelliJ settings in {}", highlight(".idea/clikd.xml"))));
    } else {
        if Confirm::new()
            .with_prompt("Generate VS Code settings?")
            .default(false)
            .interact()?
        {
            vscode::create_settings(&project_root)?;
            println!("{}", success_message(&format!("Generated VS Code settings in {}", highlight(".vscode/settings.json"))));
        }

        if Confirm::new()
            .with_prompt("Generate IntelliJ Settings?")
            .default(false)
            .interact()?
        {
            intellij::create_settings(&project_root)?;
            println!("{}", success_message(&format!("Generated IntelliJ settings in {}", highlight(".idea/clikd.xml"))));
        }
    }

    println!("\n{}\n", success_message(&format!("Finished {} init", highlight("clikd"))));

    Ok(())
}

fn sanitize_project_id(id: &str) -> String {
    id.chars()
        .filter(|c| c.is_alphanumeric() || *c == '_' || *c == '-' || *c == '.')
        .take(40)
        .collect::<String>()
        .trim_start_matches(|c| c == '_' || c == '-' || c == '.')
        .to_string()
}
