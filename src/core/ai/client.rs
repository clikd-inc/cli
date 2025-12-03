use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};

use super::credentials::{resolve_credential, ClaudeCredential};

const API_URL: &str = "https://api.anthropic.com/v1/messages";
const ANTHROPIC_VERSION: &str = "2023-06-01";

#[derive(Serialize)]
struct Message {
    role: String,
    content: String,
}

#[derive(Serialize)]
struct MessagesRequest {
    model: String,
    max_tokens: u32,
    system: String,
    messages: Vec<Message>,
}

#[derive(Deserialize)]
struct ContentBlock {
    text: String,
}

#[derive(Deserialize)]
struct MessagesResponse {
    content: Vec<ContentBlock>,
}

pub struct AnthropicClient {
    credential: ClaudeCredential,
    model: String,
}

impl AnthropicClient {
    pub async fn new() -> Result<Self> {
        let credential = resolve_credential()?;
        Ok(Self {
            credential,
            model: "claude-sonnet-4-5-20250929".to_string(),
        })
    }

    pub async fn complete(&self, system: &str, user: &str) -> Result<String> {
        let client = reqwest::Client::new();

        let request_body = MessagesRequest {
            model: self.model.clone(),
            max_tokens: 4096,
            system: system.to_string(),
            messages: vec![Message {
                role: "user".to_string(),
                content: user.to_string(),
            }],
        };

        let mut request = client
            .post(API_URL)
            .header("Content-Type", "application/json")
            .header("anthropic-version", ANTHROPIC_VERSION);

        request = match &self.credential {
            ClaudeCredential::ApiKey(key) => request.header("x-api-key", key),
            ClaudeCredential::OAuthToken { access_token, .. } => {
                request.header("Authorization", format!("Bearer {}", access_token))
            }
        };

        let response = request
            .json(&request_body)
            .send()
            .await
            .context("Failed to send request to Anthropic API")?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            anyhow::bail!("Anthropic API request failed ({}): {}", status, body);
        }

        let response: MessagesResponse = response
            .json()
            .await
            .context("Failed to parse Anthropic API response")?;

        response
            .content
            .first()
            .map(|block| block.text.clone())
            .ok_or_else(|| anyhow::anyhow!("Empty response from Anthropic API"))
    }
}
