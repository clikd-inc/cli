use std::path::{Path, PathBuf};
use std::process::Command;
use tempfile::TempDir;

pub struct TestRepo {
    pub dir: TempDir,
    pub path: PathBuf,
}

impl TestRepo {
    pub fn new() -> Self {
        let dir = TempDir::new().expect("failed to create temp dir");
        let path = dir.path().to_path_buf();

        Self::init_git(&path);

        TestRepo { dir, path }
    }

    fn init_git(path: &Path) {
        Command::new("git")
            .args(["init"])
            .current_dir(path)
            .output()
            .expect("failed to init git");

        Command::new("git")
            .args(["config", "user.email", "test@example.com"])
            .current_dir(path)
            .output()
            .expect("failed to set git email");

        Command::new("git")
            .args(["config", "user.name", "Test User"])
            .current_dir(path)
            .output()
            .expect("failed to set git name");

        Command::new("git")
            .args(["remote", "add", "origin", "https://github.com/test/repo.git"])
            .current_dir(path)
            .output()
            .expect("failed to add remote");
    }

    pub fn write_file(&self, relative_path: &str, content: &str) {
        let full_path = self.path.join(relative_path);
        if let Some(parent) = full_path.parent() {
            std::fs::create_dir_all(parent).expect("failed to create parent dirs");
        }
        std::fs::write(full_path, content).expect("failed to write file");
    }

    pub fn commit(&self, message: &str) {
        Command::new("git")
            .args(["add", "-A"])
            .current_dir(&self.path)
            .output()
            .expect("failed to git add");

        Command::new("git")
            .args(["commit", "-m", message])
            .current_dir(&self.path)
            .output()
            .expect("failed to git commit");
    }

    pub fn run_clikd_command(&self, args: &[&str]) -> std::process::Output {
        let clikd_bin = env!("CARGO_BIN_EXE_clikd");

        Command::new(clikd_bin)
            .args(args)
            .current_dir(&self.path)
            .output()
            .expect("failed to run clikd command")
    }

    pub fn file_exists(&self, relative_path: &str) -> bool {
        self.path.join(relative_path).exists()
    }

    pub fn read_file(&self, relative_path: &str) -> String {
        std::fs::read_to_string(self.path.join(relative_path))
            .expect("failed to read file")
    }

    pub fn has_config_dir(&self) -> bool {
        self.path.join(".clikd").is_dir()
    }
}

pub fn create_go_project(repo: &TestRepo, dir: &str, module_name: &str) {
    let go_mod = format!(
        "module {}\n\ngo 1.21\n\nrequire (\n\tgithub.com/gin-gonic/gin v1.9.0\n)\n",
        module_name
    );
    repo.write_file(&format!("{}/go.mod", dir), &go_mod);

    let main_go = r#"package main

import "fmt"

func main() {
    fmt.Println("Hello from Go!")
}
"#;
    repo.write_file(&format!("{}/main.go", dir), main_go);
}

pub fn create_elixir_project(repo: &TestRepo, dir: &str, app_name: &str, version: &str) {
    let mix_exs = format!(
        r#"defmodule {}.MixProject do
  use Mix.Project

  def project do
    [
      app: :{},
      version: "{}",
      elixir: "~> 1.14",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  def application do
    [
      extra_applications: [:logger]
    ]
  end

  defp deps do
    [
      {{:phoenix, "~> 1.7"}}
    ]
  end
end
"#,
        app_name.replace('_', ""),
        app_name,
        version
    );
    repo.write_file(&format!("{}/mix.exs", dir), &mix_exs);
}

pub fn create_npm_project(repo: &TestRepo, dir: &str, name: &str, version: &str) {
    let package_json = format!(
        r#"{{
  "name": "{}",
  "version": "{}",
  "description": "Test NPM package",
  "main": "index.js",
  "scripts": {{
    "test": "jest"
  }},
  "dependencies": {{
    "react": "^18.0.0"
  }}
}}
"#,
        name, version
    );
    repo.write_file(&format!("{}/package.json", dir), &package_json);
}

pub fn create_rust_project(repo: &TestRepo, dir: &str, name: &str, version: &str) {
    let cargo_toml = format!(
        r#"[package]
name = "{}"
version = "{}"
edition = "2021"

[dependencies]
serde = {{ version = "1.0", features = ["derive"] }}
"#,
        name, version
    );
    repo.write_file(&format!("{}/Cargo.toml", dir), &cargo_toml);

    repo.write_file(
        &format!("{}/src/lib.rs", dir),
        "pub fn hello() -> String { String::from(\"Hello\") }\n",
    );
}

pub fn create_python_project(repo: &TestRepo, dir: &str, name: &str, version: &str) {
    let setup_py = format!(
        r#"from setuptools import setup, find_packages

setup(
    name="{}",
    version="{}",
    packages=find_packages(),
    install_requires=[
        "requests>=2.28.0",
    ],
)
"#,
        name, version
    );
    repo.write_file(&format!("{}/setup.py", dir), &setup_py);
}

pub fn create_pyproject_toml(repo: &TestRepo, dir: &str, name: &str, version: &str) {
    let pyproject = format!(
        r#"[project]
name = "{}"
version = "{}"
description = "Test Python package"
requires-python = ">=3.8"

dependencies = [
    "requests>=2.28.0",
]
"#,
        name, version
    );
    repo.write_file(&format!("{}/pyproject.toml", dir), &pyproject);
}
