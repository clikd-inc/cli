use anyhow::Result;
use crate::config::Config;
use crate::core::auth::{github, token, org_check};
use indicatif::{ProgressBar, ProgressStyle};
use owo_colors::OwoColorize;
use std::time::Duration;

pub async fn login(no_browser: bool, config: &Config) -> Result<()> {
    let device_response = github::request_device_code(&config.github.oauth_client_id).await?;

    println!("\n{}", "GitHub Authentication".bold().cyan());
    println!("{}", "─".repeat(50).dimmed());

    if !no_browser {
        println!("\n{} Opening browser to:", "→".green());
        println!("  {}", device_response.verification_uri.bright_blue().underline());

        if let Err(e) = open::that(&device_response.verification_uri) {
            eprintln!("{} Failed to open browser: {}", "!".yellow(), e);
            println!("{} Please open the URL manually", "→".yellow());
        }
    } else {
        println!("\n{} Please visit:", "→".cyan());
        println!("  {}", device_response.verification_uri.bright_blue().underline());
    }

    println!("\n{} Enter code:", "→".cyan());
    println!("  {}", device_response.user_code.bright_green().bold());
    println!("\n{}", "Waiting for authorization...".dimmed());

    let pb = ProgressBar::new_spinner();
    pb.set_style(
        ProgressStyle::default_spinner()
            .tick_chars("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
            .template("{spinner:.cyan} {msg}")
            .unwrap(),
    );
    pb.set_message("Polling for authorization...");
    pb.enable_steady_tick(Duration::from_millis(100));

    let token_result = github::poll_for_token(
        &config.github.oauth_client_id,
        &device_response.device_code,
        device_response.interval,
        device_response.expires_in,
    )
    .await;

    pb.finish_and_clear();

    let access_token = token_result?;

    println!("{} Getting user info...", "→".cyan());
    let username = github::get_username(&access_token).await?;

    println!("{} Verifying organization membership...", "→".cyan());
    org_check::verify_membership(&access_token, &config.github.org_name).await?;

    token::save_token(&access_token)?;

    println!("\n{} Successfully authenticated as {}",
        "✓".green().bold(),
        username.bright_green().bold()
    );
    println!("{} Organization: {}",
        "✓".green().bold(),
        config.github.org_name.bright_green()
    );

    Ok(())
}

pub async fn logout() -> Result<()> {
    match token::load_token() {
        Ok(_) => {
            token::delete_token()?;
            println!("{} Successfully logged out", "✓".green().bold());
        }
        Err(_) => {
            println!("{} Not currently logged in", "ℹ".blue().bold());
        }
    }
    Ok(())
}

pub async fn status() -> Result<()> {
    match token::load_token() {
        Ok(token) => {
            match github::get_username(&token).await {
                Ok(username) => {
                    println!("{} Logged in as {}",
                        "✓".green().bold(),
                        username.bright_green().bold()
                    );
                }
                Err(_) => {
                    println!("{} Token exists but may be invalid", "⚠".yellow().bold());
                    println!("{} Run 'clikd auth login' to re-authenticate", "→".cyan());
                }
            }
        }
        Err(_) => {
            println!("{} Not logged in", "ℹ".blue().bold());
            println!("{} Run 'clikd auth login' to authenticate", "→".cyan());
        }
    }
    Ok(())
}
