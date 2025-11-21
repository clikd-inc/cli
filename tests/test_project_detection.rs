mod common;
use common::TestRepo;

#[test]
fn test_detect_single_rust_project() {
    let repo = TestRepo::from_fixture("single-rust");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-crate"), "Rust project not detected");
    assert!(bootstrap.contains("0.1.0"), "Version not detected");
}

#[test]
fn test_detect_single_go_project() {
    let repo = TestRepo::from_fixture("single-go");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("myapp"), "Go project not detected");
}

#[test]
fn test_detect_single_elixir_project() {
    let repo = TestRepo::from_fixture("single-elixir");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("my_app"), "Elixir project not detected");
    assert!(bootstrap.contains("1.0.0"), "Version not detected");
}

#[test]
fn test_detect_single_npm_project() {
    let repo = TestRepo::from_fixture("single-npm");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-package"), "NPM project not detected");
    assert!(bootstrap.contains("2.0.0"), "Version not detected");
}

#[test]
fn test_detect_single_python_project() {
    let repo = TestRepo::from_fixture("single-python");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("my-python-pkg"), "Python project not detected");
    assert!(bootstrap.contains("3.0.0"), "Version not detected");
}

#[test]
fn test_detect_monorepo_multiple_rust_crates() {
    let repo = TestRepo::from_fixture("monorepo-rust");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("web"), "web crate not detected");
    assert!(bootstrap.contains("api"), "api crate not detected");
    assert!(bootstrap.contains("core"), "core crate not detected");
}

#[test]
fn test_detect_monorepo_mixed_languages() {
    let repo = TestRepo::from_fixture("monorepo-mixed");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("backend"), "Rust backend not detected");
    assert!(bootstrap.contains("frontend"), "NPM frontend not detected");
    assert!(bootstrap.contains("deployment-tools"), "Python scripts not detected");
}

#[test]
fn test_detect_workspace_with_dependencies() {
    let repo = TestRepo::from_fixture("workspace-rust");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(output.status.success(), "Failed to init: {:?}", String::from_utf8_lossy(&output.stderr));

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(bootstrap.contains("common"), "common crate not detected");
    assert!(bootstrap.contains("app"), "app crate not detected");
}
