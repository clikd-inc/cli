mod common;
use common::{TestRepo, create_rust_project, create_go_project, create_elixir_project, create_npm_project, create_python_project};

#[test]
fn test_detect_single_rust_project() {
    let repo = TestRepo::new();

    create_rust_project(&repo, ".", "my-crate", "0.1.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-crate"), "Rust project not detected");
    assert!(bootstrap.contains("0.1.0"), "Version not detected");
}

#[test]
fn test_detect_single_go_project() {
    let repo = TestRepo::new();

    create_go_project(&repo, ".", "github.com/test/myapp");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("myapp"), "Go project not detected");
}

#[test]
fn test_detect_single_elixir_project() {
    let repo = TestRepo::new();

    create_elixir_project(&repo, ".", "my_app", "1.0.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("my_app"), "Elixir project not detected");
    assert!(bootstrap.contains("1.0.0"), "Version not detected");
}

#[test]
fn test_detect_single_npm_project() {
    let repo = TestRepo::new();

    create_npm_project(&repo, ".", "my-package", "2.0.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-package"), "NPM project not detected");
    assert!(bootstrap.contains("2.0.0"), "Version not detected");
}

#[test]
fn test_detect_single_python_project() {
    let repo = TestRepo::new();

    create_python_project(&repo, ".", "my-python-pkg", "3.0.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-python-pkg"), "Python project not detected");
    assert!(bootstrap.contains("3.0.0"), "Version not detected");
}

#[test]
fn test_detect_monorepo_multiple_rust_crates() {
    let repo = TestRepo::new();

    create_rust_project(&repo, "crates/web", "web", "0.1.0");
    create_rust_project(&repo, "crates/api", "api", "0.2.0");
    create_rust_project(&repo, "crates/core", "core", "1.0.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("web"), "web crate not detected");
    assert!(bootstrap.contains("api"), "api crate not detected");
    assert!(bootstrap.contains("core"), "core crate not detected");
}

#[test]
fn test_detect_monorepo_mixed_languages() {
    let repo = TestRepo::new();

    create_rust_project(&repo, "backend", "backend", "1.0.0");
    create_npm_project(&repo, "frontend", "frontend", "1.0.0");
    create_python_project(&repo, "scripts", "deployment-tools", "0.5.0");
    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("backend"), "Rust backend not detected");
    assert!(bootstrap.contains("frontend"), "NPM frontend not detected");
    assert!(bootstrap.contains("deployment-tools"), "Python scripts not detected");
}

#[test]
fn test_detect_workspace_with_dependencies() {
    let repo = TestRepo::new();

    repo.write_file("Cargo.toml", r#"[workspace]
members = ["crates/common", "crates/app"]
resolver = "2"
"#);

    repo.write_file("crates/common/Cargo.toml", r#"[package]
name = "common"
version = "0.1.0"
edition = "2021"
"#);

    repo.write_file("crates/app/Cargo.toml", r#"[package]
name = "app"
version = "0.1.0"
edition = "2021"

[dependencies]
common = { path = "../common" }
"#);

    repo.write_file("crates/common/src/lib.rs", "pub fn hello() {}");
    repo.write_file("crates/app/src/lib.rs", "use common::hello;");

    repo.commit("Initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file(".clikd/bootstrap.toml");
    assert!(bootstrap.contains("common"), "common crate not detected");
    assert!(bootstrap.contains("app"), "app crate not detected");
}
