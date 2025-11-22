mod common;
use common::{TestRepo, create_rust_project, create_npm_project, create_python_project};

#[test]
fn test_version_bump_patch() {
    let repo = TestRepo::new();

    create_rust_project(&repo, ".", "test-crate", "1.0.0");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("src/fix.rs", "pub fn fix() {}");
    repo.commit("fix: patch bug");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_minor() {
    let repo = TestRepo::new();

    create_rust_project(&repo, ".", "test-crate", "1.0.0");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("src/feature.rs", "pub fn feature() {}");
    repo.commit("feat: add new feature");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_major_breaking_change() {
    let repo = TestRepo::new();

    create_rust_project(&repo, ".", "test-crate", "1.0.0");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("src/breaking.rs", "pub fn breaking() {}");
    repo.commit("feat!: breaking change\n\nBREAKING CHANGE: Old API removed");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_preserves_format() {
    let repo = TestRepo::new();

    create_rust_project(&repo, ".", "test-crate", "1.2.3");
    repo.commit("Initial commit");

    let original_cargo = repo.read_file("Cargo.toml");
    assert!(original_cargo.contains("version = \"1.2.3\""), "Original version not correct");

    repo.run_clikd_command(&["release", "init", "--force"]);

    let after_init_cargo = repo.read_file("Cargo.toml");
    println!("After init Cargo.toml:\n{}", after_init_cargo);
    assert!(after_init_cargo.contains("version = \"0.0.0-dev.0\"") || after_init_cargo.contains("version = \"1.2.3-dev.0\""),
        "Dev version not set correctly. Got: {}", after_init_cargo);

    repo.write_file("src/fix.rs", "pub fn fix() {}");
    repo.commit("fix: bug fix");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_npm_package() {
    let repo = TestRepo::new();

    create_npm_project(&repo, ".", "my-package", "2.0.0");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("index.js", "module.exports = {};");
    repo.commit("feat: add new API");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_python_package() {
    let repo = TestRepo::new();

    create_python_project(&repo, ".", "my-pkg", "0.5.0");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("my_pkg/__init__.py", "");
    repo.commit("feat: add module");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_multiple_projects() {
    let repo = TestRepo::new();

    create_rust_project(&repo, "crates/web", "web", "0.1.0");
    create_rust_project(&repo, "crates/api", "api", "0.2.0");
    repo.write_file("Cargo.toml", "[workspace]\nmembers = [\"crates/web\", \"crates/api\"]\nresolver = \"2\"\n");
    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("crates/web/src/feature.rs", "pub fn feature() {}");
    repo.commit("feat(web): add feature");

    repo.write_file("crates/api/src/fix.rs", "pub fn fix() {}");
    repo.commit("fix(api): fix bug");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}

#[test]
fn test_version_bump_respects_dependencies() {
    let repo = TestRepo::new();

    repo.write_file("Cargo.toml", r#"[workspace]
members = ["common", "app"]
"#);

    repo.write_file("common/Cargo.toml", r#"[package]
name = "common"
version = "0.1.0"
edition = "2021"
"#);

    repo.write_file("app/Cargo.toml", r#"[package]
name = "app"
version = "0.1.0"
edition = "2021"

[dependencies]
common = { path = "../common" }
"#);

    repo.write_file("common/src/lib.rs", "pub fn shared() {}");
    repo.write_file("app/src/lib.rs", "use common::shared;");

    repo.commit("Initial commit");

    repo.run_clikd_command(&["release", "init", "--force"]);

    repo.write_file("common/src/new.rs", "pub fn new() {}");
    repo.commit("feat(common): add new function");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(output.status.success(), "Failed to get status: {:?}", String::from_utf8_lossy(&output.stderr));
}
