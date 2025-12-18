use anyhow::{Context, Result};
use once_cell::sync::Lazy;
use regex::Regex;
use std::collections::HashMap;

static DOCKERFILE_IMAGES: Lazy<HashMap<String, String>> =
    Lazy::new(|| parse_dockerfile().unwrap_or_default());

pub fn get_image(service: &str) -> Option<String> {
    DOCKERFILE_IMAGES.get(service).cloned()
}

fn parse_dockerfile() -> Result<HashMap<String, String>> {
    static FROM_PATTERN: Lazy<Regex> = Lazy::new(|| {
        Regex::new(r"(?m)^FROM\s+([^\s]+)\s+AS\s+(\w+)").expect("Invalid regex pattern")
    });

    let dockerfile_content = include_str!("../../../config/images.Dockerfile");

    let mut images = HashMap::new();

    for cap in FROM_PATTERN.captures_iter(dockerfile_content) {
        let image = cap
            .get(1)
            .context("Missing image in FROM statement")?
            .as_str()
            .to_string();
        let alias = cap
            .get(2)
            .context("Missing alias in FROM statement")?
            .as_str()
            .to_string();

        images.insert(alias, image);
    }

    if images.is_empty() {
        anyhow::bail!("No images found in Dockerfile");
    }

    Ok(images)
}

pub fn get_all_images() -> HashMap<String, String> {
    DOCKERFILE_IMAGES.clone()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_dockerfile() {
        let images = parse_dockerfile().expect("BUG: should parse dockerfile");

        assert!(images.contains_key("gate"));
        assert!(images.contains_key("rig"));
        assert!(images.contains_key("studio"));
        assert!(images.contains_key("postgres"));
        assert!(images.contains_key("keydb"));
        assert!(images.contains_key("scylladb"));
        assert!(images.contains_key("minio"));
        assert!(images.contains_key("nats"));
        assert!(images.contains_key("apisix"));

        assert_eq!(images.len(), 9);
    }

    #[test]
    fn test_get_image() {
        let gate_image = get_image("gate");
        assert!(gate_image.is_some());
        assert!(gate_image
            .expect("BUG: gate_image should be Some after assertion")
            .starts_with("ghcr.io/clikd-inc/gate:"));
    }

    #[test]
    fn test_image_format() {
        let images = parse_dockerfile().expect("BUG: should parse dockerfile");

        for (service, image) in &images {
            assert!(
                image.contains(':'),
                "Image {image} for service {service} should contain version tag"
            );
            assert!(
                image.starts_with("ghcr.io/clikd-inc/"),
                "Image {image} should start with ghcr.io/clikd-inc/"
            );
        }
    }
}
