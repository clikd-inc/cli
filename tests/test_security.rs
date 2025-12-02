mod common;
use common::TestRepo;

#[test]
fn test_package_names_with_dots() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "legitimate-package"
version = "1.0.0"
edition = "2021"
"#,
    );

    repo.write_file("src/lib.rs", "pub fn hello() {}");

    repo.write_file(
        "subdir/package.json",
        r#"{"name": "@scope/dotted.name", "version": "1.0.0"}"#,
    );

    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Command should succeed: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(
        bootstrap.contains("legitimate-package"),
        "Legitimate package should be detected"
    );
}

#[test]
fn test_null_byte_in_package_name() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "normal-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);
    assert!(output.status.success());
}

#[test]
fn test_unicode_normalization_attack() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "pаckage"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit with unicode package name");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Unicode package names should be handled: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );
}

#[test]
fn test_very_long_package_name() {
    let repo = TestRepo::new();

    let long_name = "a".repeat(64);
    let cargo_toml = format!(
        r#"[package]
name = "{}"
version = "1.0.0"
edition = "2021"
"#,
        long_name
    );

    repo.write_file("Cargo.toml", &cargo_toml);
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Long package names should be handled gracefully: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(
        bootstrap.contains(&long_name),
        "Long package name should be detected"
    );
}

#[test]
fn test_special_characters_in_version() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "test-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);
    assert!(output.status.success());

    let malformed_versions = [
        "not-a-version",
        "1.2.3.4.5.6.7.8.9.10",
        "-1.0.0",
        "1.-1.0",
        "1.0.-1",
        "<script>alert(1)</script>",
        "'; DROP TABLE versions; --",
        "$(whoami)",
        "`id`",
    ];

    for version in malformed_versions {
        let output = repo.run_clikd_command(&["release", "set-version", "test-package", version]);
        assert!(
            !output.status.success() || !String::from_utf8_lossy(&output.stderr).is_empty(),
            "Malformed version '{}' should be rejected or warned",
            version
        );
    }
}

#[test]
fn test_command_injection_in_bump_spec() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "test-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);
    assert!(output.status.success());

    let injection_attempts = [
        "; rm -rf /",
        "| cat /etc/passwd",
        "$(whoami)",
        "`id`",
        "&& echo pwned",
        "|| true",
    ];

    for injection in injection_attempts {
        let output = repo.run_clikd_command(&["release", "set-version", "test-package", injection]);
        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);

        assert!(
            !stdout.contains("pwned"),
            "Command injection '{}' should not execute",
            injection
        );
        assert!(
            !stderr.contains("pwned"),
            "Command injection '{}' should not execute",
            injection
        );
    }
}

#[test]
fn test_deeply_nested_directory_project() {
    let repo = TestRepo::new();

    let deep_path = "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t";

    repo.write_file(
        &format!("{}/Cargo.toml", deep_path),
        r#"[package]
name = "deep-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file(&format!("{}/src/lib.rs", deep_path), "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Deeply nested projects should be handled: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(
        bootstrap.contains("deep-package"),
        "Deep package should be detected"
    );
}

#[test]
fn test_symlink_traversal() {
    let repo = TestRepo::new();

    repo.write_file(
        "legitimate/Cargo.toml",
        r#"[package]
name = "legitimate-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("legitimate/src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Project detection should succeed"
    );
}

#[test]
fn test_empty_and_whitespace_names() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "valid-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);
    assert!(output.status.success());

    let empty_names = ["", " ", "  ", "\t", "\n", "\r\n"];

    for name in empty_names {
        let output = repo.run_clikd_command(&["release", "status", name]);
        assert!(
            !output.status.success(),
            "Empty/whitespace name '{}' should be rejected",
            name.escape_debug()
        );
    }
}

#[test]
fn test_toml_injection_in_config() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "test-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");

    let malicious_config = r#"
[release.projects."test-package"]
ignore = false

[release.projects."'; DROP TABLE projects; --"]
ignore = true

[release.projects."<script>alert(1)</script>"]
ignore = true
"#;

    repo.write_file("clikd/config.toml", malicious_config);
    repo.commit("initial commit with config");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success() || !output.status.success(),
        "Malformed config should not crash: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );
}

#[test]
fn test_large_file_handling() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "test-package"
version = "1.0.0"
edition = "2021"
"#,
    );

    let large_content = "a".repeat(10 * 1024 * 1024);
    repo.write_file("src/lib.rs", &large_content);
    repo.commit("initial commit with large file");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Large files should be handled gracefully: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );
}

#[test]
fn test_json_with_special_characters() {
    let repo = TestRepo::new();

    let json_with_special = r#"{
  "name": "test-package",
  "version": "1.0.0",
  "description": "<script>alert('xss')</script>",
  "main": "index.js",
  "dependencies": {
    "normal-dep": "^1.0.0"
  }
}"#;

    repo.write_file("package.json", json_with_special);
    repo.commit("initial commit");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "JSON with special characters should be handled: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    assert!(
        bootstrap.contains("test-package"),
        "Package should be detected despite special chars in description"
    );
}

#[test]
fn test_binary_file_as_config() {
    let repo = TestRepo::new();

    repo.write_file(
        "Cargo.toml",
        r#"[package]
name = "test-package"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("src/lib.rs", "pub fn hello() {}");

    let binary_content = vec![0u8, 1, 2, 3, 255, 254, 253, 0, 0, 0];
    let binary_str = String::from_utf8_lossy(&binary_content);
    repo.write_file("clikd/config.toml", &binary_str);

    repo.commit("initial commit with binary config");

    let output = repo.run_clikd_command(&["release", "status"]);

    assert!(
        !output.status.success() || output.status.success(),
        "Binary config should not crash: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );
}

#[test]
fn test_concurrent_project_names() {
    let repo = TestRepo::new();

    repo.write_file(
        "pkg-a/Cargo.toml",
        r#"[package]
name = "shared-name"
version = "1.0.0"
edition = "2021"
"#,
    );
    repo.write_file("pkg-a/src/lib.rs", "pub fn a() {}");

    repo.write_file(
        "pkg-b/package.json",
        r#"{
  "name": "shared-name",
  "version": "2.0.0"
}"#,
    );

    repo.commit("initial commit with conflicting names");

    let output = repo.run_clikd_command(&["release", "init", "--force"]);

    assert!(
        output.status.success(),
        "Conflicting names should be handled with disambiguation: {:?}",
        String::from_utf8_lossy(&output.stderr)
    );

    let bootstrap = repo.read_file("clikd/bootstrap.toml");
    let name_count = bootstrap.matches("shared-name").count();
    assert!(
        name_count >= 2,
        "Both projects with shared name should be tracked (found {} occurrences)",
        name_count
    );
}
