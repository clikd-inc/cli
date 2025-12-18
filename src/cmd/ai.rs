use anyhow::{Context, Result};
use dialoguer::{theme::ColorfulTheme, Select};
use owo_colors::OwoColorize;
use std::io::{self, Write};

use crate::core::ai::client::AnthropicClient;
use crate::core::ai::credentials::{
    delete_credentials, load_credentials, now_unix, store_credentials, ClaudeCredential,
};
use crate::core::ai::oauth::OAuthFlow;

pub async fn login() -> Result<()> {
    if let Some(existing) = load_credentials()? {
        println!(
            "{} Already logged in with {}",
            "!".yellow(),
            existing.credential_type()
        );
        print!("Do you want to re-authenticate? [y/N] ");
        io::stdout().flush()?;

        let mut input = String::new();
        io::stdin().read_line(&mut input)?;

        if !input.trim().eq_ignore_ascii_case("y") {
            println!("Login cancelled.");
            return Ok(());
        }
    }

    println!();
    println!("{}", "Claude AI Authentication".bold());
    println!("{}", "========================".dimmed());
    println!();

    let options = vec!["Claude Max/Pro Subscription (OAuth)", "Anthropic API Key"];

    let selection = Select::with_theme(&ColorfulTheme::default())
        .with_prompt("How would you like to authenticate?")
        .items(&options)
        .default(0)
        .interact()?;

    match selection {
        0 => login_oauth().await,
        1 => login_api_key().await,
        _ => unreachable!(),
    }
}

async fn login_oauth() -> Result<()> {
    println!();
    println!("This will authenticate clikd with your Claude Max/Pro subscription.");
    println!();

    let flow = OAuthFlow::new();
    let auth_url = flow.authorization_url();

    println!("{} Opening browser for authentication...", "→".cyan());
    println!();

    if open::that(&auth_url).is_err() {
        println!("{} Could not open browser automatically.", "!".yellow());
        println!();
        println!("Please open this URL manually:");
        println!("{}", auth_url.dimmed());
    }

    println!();
    println!(
        "{} After logging in, you will see a page with an authorization code.",
        "→".cyan()
    );
    println!();
    print!("{} Paste the authorization code here: ", "?".green());
    io::stdout().flush()?;

    let mut code = String::new();
    io::stdin().read_line(&mut code)?;
    let code = code.trim();

    if code.is_empty() {
        anyhow::bail!("No authorization code provided");
    }

    println!();
    println!("{} Exchanging code for tokens...", "→".cyan());

    let tokens = flow.exchange_code(code).await?;

    let credential = ClaudeCredential::OAuthToken {
        access_token: tokens.access_token,
        refresh_token: tokens.refresh_token,
        expires_at: now_unix() + tokens.expires_in,
    };

    store_credentials(&credential)?;

    println!();
    println!(
        "{} Successfully logged in with Claude Max/Pro subscription!",
        "✓".green().bold()
    );
    println!();
    println!(
        "You can now use {} to generate AI-powered changelogs.",
        "clikd release prepare --ai".cyan()
    );

    Ok(())
}

async fn login_api_key() -> Result<()> {
    println!();
    println!("Enter your Anthropic API key.");
    println!(
        "{}",
        "Get one at: https://console.anthropic.com/settings/keys".dimmed()
    );
    println!();

    print!("{} API Key: ", "?".green());
    io::stdout().flush()?;

    let mut api_key = String::new();
    io::stdin().read_line(&mut api_key)?;
    let api_key = api_key.trim();

    if api_key.is_empty() {
        anyhow::bail!("No API key provided");
    }

    if !api_key.starts_with("sk-ant-") {
        println!(
            "{} Warning: API key doesn't start with 'sk-ant-'. Are you sure it's correct?",
            "!".yellow()
        );
        print!("Continue anyway? [y/N] ");
        io::stdout().flush()?;

        let mut confirm = String::new();
        io::stdin().read_line(&mut confirm)?;
        if !confirm.trim().eq_ignore_ascii_case("y") {
            println!("Login cancelled.");
            return Ok(());
        }
    }

    let credential = ClaudeCredential::ApiKey(api_key.to_string());
    store_credentials(&credential)?;

    println!();
    println!("{} Successfully saved API key!", "✓".green().bold());
    println!();
    println!(
        "You can now use {} to generate AI-powered changelogs.",
        "clikd release prepare --ai".cyan()
    );

    Ok(())
}

pub async fn logout() -> Result<()> {
    match load_credentials()? {
        Some(creds) => {
            delete_credentials()?;
            println!(
                "{} Logged out from {} credentials.",
                "✓".green(),
                creds.credential_type()
            );
        }
        None => {
            println!("{} Not logged in.", "!".yellow());
        }
    }
    Ok(())
}

pub async fn status() -> Result<()> {
    println!();
    println!("{}", "Claude AI Authentication Status".bold());
    println!("{}", "================================".dimmed());
    println!();

    if let Ok(api_key) = std::env::var("ANTHROPIC_API_KEY") {
        let masked = if api_key.len() > 8 {
            format!("{}...{}", &api_key[..4], &api_key[api_key.len() - 4..])
        } else {
            "****".to_string()
        };
        println!("{} Using API Key from environment", "✓".green());
        println!("  Key: {}", masked.dimmed());
        println!();
        println!(
            "{}",
            "Note: ANTHROPIC_API_KEY takes priority over OAuth login.".dimmed()
        );
        return Ok(());
    }

    match load_credentials()? {
        Some(creds) => {
            println!("{} Logged in", "✓".green());
            println!("  Type: {}", creds.credential_type());

            if let ClaudeCredential::OAuthToken { expires_at, .. } = &creds {
                let now = now_unix();
                if *expires_at > now {
                    let remaining = expires_at - now;
                    let hours = remaining / 3600;
                    let minutes = (remaining % 3600) / 60;
                    println!("  Expires in: {}h {}m", hours, minutes);
                } else {
                    println!("  Status: {} (will auto-refresh)", "Expired".yellow());
                }
            }
        }
        None => {
            println!("{} Not logged in", "✗".red());
            println!();
            println!(
                "Run {} to authenticate with Claude Max/Pro.",
                "clikd ai login".cyan()
            );
            println!(
                "Or set {} environment variable for API key auth.",
                "ANTHROPIC_API_KEY".cyan()
            );
        }
    }

    Ok(())
}

pub async fn test() -> Result<()> {
    println!();
    println!("{}", "Claude AI Connection Test".bold());
    println!("{}", "=========================".dimmed());
    println!();

    println!("{} Checking credentials...", "→".cyan());

    let creds = load_credentials()?;
    let cred_source = if std::env::var("ANTHROPIC_API_KEY").is_ok() {
        "ANTHROPIC_API_KEY environment variable"
    } else if creds.is_some() {
        creds.as_ref().map(|c| c.credential_type()).unwrap_or("Unknown")
    } else {
        println!("{} No credentials found", "✗".red());
        println!();
        println!(
            "Run {} to authenticate first.",
            "clikd ai login".cyan()
        );
        return Ok(());
    };

    println!("  Credential source: {}", cred_source.dimmed());
    println!();

    println!("{} Initializing API client...", "→".cyan());

    let client = AnthropicClient::new()
        .await
        .context("failed to initialize Anthropic client")?;

    println!("  Model: {}", "claude-sonnet-4-5-20250929".dimmed());
    println!();

    println!("{} Sending test request...", "→".cyan());

    let start = std::time::Instant::now();
    let response = client
        .complete(
            "You are a helpful assistant. Respond with exactly one short sentence.",
            "Say hello and confirm you're working.",
        )
        .await
        .context("API request failed")?;
    let elapsed = start.elapsed();

    println!();
    println!("{} API connection successful!", "✓".green().bold());
    println!();
    println!("  Response time: {:?}", elapsed);
    println!("  Response: {}", response.trim().dimmed());
    println!();
    println!(
        "{}",
        "AI changelog generation is ready to use with 'clikd release prepare'.".green()
    );

    Ok(())
}
