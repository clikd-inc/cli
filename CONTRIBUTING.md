# Contributing to Clikd

## Development Setup

### Prerequisites

- Rust 1.75+ (stable)
- Git 2.x

### Building

```bash
cargo build
cargo build --release
```

### Running Tests

```bash
cargo test
cargo test --all-features
```

## Testing Guidelines

### Test Structure

Tests are organized into:
- **Unit tests**: In-file `#[cfg(test)]` modules
- **Integration tests**: `tests/` directory

### Writing Integration Tests

Integration tests use the `TestRepo` helper from `tests/common.rs`:

```rust
mod common;
use common::TestRepo;

#[test]
fn test_my_feature() {
    let repo = TestRepo::new();

    // Create project files dynamically
    repo.write_file("Cargo.toml", r#"
[package]
name = "test-crate"
version = "1.0.0"
edition = "2021"
"#);
    repo.write_file("src/lib.rs", "pub fn hello() {}\n");
    repo.commit("Initial commit");

    // Run clikd commands
    repo.run_clikd_command(&["release", "init", "--force"]);

    // Make changes and verify
    repo.write_file("src/feature.rs", "pub fn feature() {}");
    repo.commit("feat: add new feature");

    let output = repo.run_clikd_command(&["release", "status"]);
    assert!(output.status.success());
}
```

### TestRepo API

| Method | Description |
|--------|-------------|
| `TestRepo::new()` | Create new temp directory with git init |
| `repo.write_file(path, content)` | Write file relative to repo root |
| `repo.read_file(path)` | Read file contents |
| `repo.file_exists(path)` | Check if file exists |
| `repo.commit(message)` | Stage all and commit |
| `repo.run_clikd_command(args)` | Execute clikd binary |
| `repo.has_config_dir()` | Check if `.clikd/` exists |

### Project Type Test Templates

#### Cargo (Rust)

```rust
repo.write_file("Cargo.toml", r#"
[package]
name = "my-crate"
version = "1.0.0"
edition = "2021"
"#);
repo.write_file("src/lib.rs", "");
```

#### NPM (JavaScript/TypeScript)

```rust
repo.write_file("package.json", r#"{
  "name": "my-package",
  "version": "1.0.0",
  "main": "index.js"
}"#);
repo.write_file("index.js", "module.exports = {};");
```

#### Python (PyPA)

```rust
repo.write_file("setup.cfg", r#"
[metadata]
name = my-package
version = 1.0.0
"#);
repo.write_file("setup.py", r#"
from setuptools import setup
version = "1.0.0"  # clikd project-version
setup()
"#);
repo.write_file("my_package/__init__.py", "");
```

#### Cargo Workspace

```rust
repo.write_file("Cargo.toml", r#"
[workspace]
members = ["crates/a", "crates/b"]
resolver = "2"
"#);
repo.write_file("crates/a/Cargo.toml", r#"
[package]
name = "a"
version = "0.1.0"
edition = "2021"
"#);
repo.write_file("crates/a/src/lib.rs", "");
// ... repeat for crates/b
```

### Conventional Commit Messages

Use conventional commits in tests to trigger version bumps:

| Prefix | Version Bump | Example |
|--------|--------------|---------|
| `fix:` | Patch | `fix: correct calculation` |
| `feat:` | Minor | `feat: add export option` |
| `feat!:` | Major | `feat!: breaking API change` |
| `fix(scope):` | Patch (scoped) | `fix(api): handle empty input` |

### Test Naming

Use descriptive test names:

```rust
#[test]
fn test_version_bump_patch() { }

#[test]
fn test_version_bump_major_breaking_change() { }

#[test]
fn test_workspace_with_internal_dependencies() { }
```

### Running Specific Tests

```bash
cargo test test_version_bump
cargo test test_changelog
cargo test --test test_release_init
```

### Test Coverage

Run tests with coverage (requires `cargo-llvm-cov`):

```bash
cargo llvm-cov --all-features
```

## Code Style

### Formatting

```bash
cargo fmt
cargo fmt -- --check  # CI check
```

### Linting

```bash
cargo clippy --all-targets --all-features -- -D warnings
```

### Documentation

```bash
cargo doc --no-deps
```

## Breaking Changes Policy

### What Constitutes a Breaking Change

**CLI Breaking Changes:**
- Removing or renaming commands/subcommands
- Removing or renaming required flags
- Changing default behavior that affects output
- Changing exit codes for error conditions
- Removing output fields in structured formats (JSON)

**Config Breaking Changes:**
- Removing configuration keys
- Changing configuration key semantics
- Changing default values with user-visible impact

**Not Breaking:**
- Adding new commands/flags (additive)
- Adding new optional configuration keys
- Internal refactoring without behavior change
- Bug fixes (even if someone depended on buggy behavior)

### Commit Message Convention

Breaking changes must be marked in commit messages:

```
feat!: rename --output flag to --format

BREAKING CHANGE: The --output flag has been renamed to --format.
Users must update their scripts accordingly.
```

Or with scope:

```
feat(cli)!: change default output format to JSON
```

### Deprecation Process

1. **Announce**: Add deprecation warning in current release
2. **Document**: Update docs with migration path
3. **Grace Period**: Maintain deprecated feature for 2 minor versions
4. **Remove**: Remove in next major version

Example deprecation warning:

```rust
eprintln!(
    "Warning: --output is deprecated and will be removed in v2.0. Use --format instead."
);
```

### Migration Guides

For major version bumps, include migration guide in release notes:

```markdown
## Migration from v1.x to v2.0

### Breaking Changes

1. **`--output` renamed to `--format`**
   - Before: `clikd release status --output json`
   - After: `clikd release status --format json`

2. **Config key `analysis.cache_size` split**
   - Before: `cache_size = 512`
   - After: `commit_cache_size = 512` and `tree_cache_size = 3`
```

## Pull Request Process

1. Fork the repository
2. Create feature branch from `main`
3. Implement changes with tests
4. Ensure CI passes (`cargo test`, `cargo clippy`, `cargo fmt`)
5. Submit PR with descriptive title and description

### PR Title Convention

Use conventional commit format:

- `feat: add release preview command`
- `fix: handle empty changelog gracefully`
- `docs: update README installation section`
- `refactor: simplify version parsing`
- `test: add workspace dependency tests`

### PR Checklist

- [ ] Tests pass locally
- [ ] New functionality has tests
- [ ] Documentation updated (if applicable)
- [ ] Breaking changes documented
- [ ] No new clippy warnings

## Release Process

Releases are automated via GitHub Actions when tags are pushed:

```bash
git tag v0.5.0
git push origin v0.5.0
```

The CI will:
1. Run all tests
2. Build release binaries for all platforms
3. Create GitHub release with binaries
4. Publish to crates.io (if configured)

## Questions?

Open an issue for:
- Bug reports
- Feature requests
- Questions about contributing
