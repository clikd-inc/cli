use crate::config::Config;
use crate::core::auth::{github, org_check, token};
use crate::utils::theme::*;
use anyhow::Result;

pub async fn login(no_browser: bool, config: &Config) -> Result<()> {
    let device_response = github::request_device_code(&config.github.oauth_client_id).await?;

    println!("{}", header("GitHub Authentication"));

    if !no_browser {
        println!("\n{}", step_message("Opening browser to:"));
        println!("  {}", url(&device_response.verification_uri));

        if let Err(e) = open::that(&device_response.verification_uri) {
            eprintln!(
                "{}",
                warning_message(&format!("Failed to open browser: {}", e))
            );
            println!("{}", step_message("Please open the URL manually"));
        }
    } else {
        println!("\n{}", step_message("Please visit:"));
        println!("  {}", url(&device_response.verification_uri));
    }

    println!("\n{}", step_message("Enter code:"));
    println!("  {}", code(&device_response.user_code));
    println!();

    let mut sp = create_spinner("Waiting for authorization...");

    let access_token = match github::poll_for_token(
        &config.github.oauth_client_id,
        &device_response.device_code,
        device_response.interval,
        device_response.expires_in,
    )
    .await
    {
        Ok(token) => {
            sp.success("Authorized!");
            token
        }
        Err(e) => {
            sp.fail("Authorization failed");
            return Err(e.into());
        }
    };

    println!("{}", step_message("Getting user info..."));
    let username = github::get_username(&access_token).await?;

    println!("{}", step_message("Verifying organization membership..."));
    org_check::verify_membership(&access_token, &config.github.org_name).await?;

    token::save_token(&access_token)?;

    println!(
        "\n{}",
        success_message(&format!(
            "Successfully authenticated as {}",
            highlight(&username)
        ))
    );
    println!(
        "{}",
        success_message(&format!(
            "Organization: {}",
            highlight(&config.github.org_name)
        ))
    );

    Ok(())
}

pub async fn logout() -> Result<()> {
    match token::load_token() {
        Ok(_) => {
            token::delete_token()?;
            println!("{}", success_message("Successfully logged out"));
        }
        Err(_) => {
            println!("{}", info_message("Not currently logged in"));
        }
    }
    Ok(())
}

pub async fn status() -> Result<()> {
    match token::load_token() {
        Ok(token) => match github::get_username(&token).await {
            Ok(username) => {
                println!(
                    "{}",
                    success_message(&format!("Logged in as {}", highlight(&username)))
                );
            }
            Err(_) => {
                println!("{}", warning_message("Token exists but may be invalid"));
                println!("{}", step_message("Run 'clikd login' to re-authenticate"));
            }
        },
        Err(_) => {
            println!("{}", info_message("Not logged in"));
            println!("{}", step_message("Run 'clikd login' to authenticate"));
        }
    }
    Ok(())
}
