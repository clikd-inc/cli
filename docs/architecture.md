# Clikd Architecture

## Overview

Clikd is a release management CLI that supports multiple package ecosystems within monorepos. The architecture follows a modular loader pattern that discovers projects, builds a dependency graph, and analyzes git histories.

## Core Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AppBuilder                                      │
│                    (Session initialization & orchestration)                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Ecosystem Loaders                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │  Cargo   │ │   NPM    │ │   PyPA   │ │   Go     │ │  Elixir  │  ...     │
│  │ Loader   │ │  Loader  │ │  Loader  │ │  Loader  │ │  Loader  │          │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘          │
│       │            │            │            │            │                 │
│       └────────────┴────────────┴────────────┴────────────┘                 │
│                                 │                                            │
│                    process_index_item()                                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ProjectGraphBuilder                                  │
│                                                                              │
│   • Registers projects via try_add_project()                                │
│   • Collects internal dependencies                                           │
│   • Resolves qualified names to unique user-facing names                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                          complete_loading()
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            ProjectGraph                                      │
│                                                                              │
│   • DAG structure via petgraph                                              │
│   • Topologically sorted project iteration                                   │
│   • Dependency cycle detection                                               │
│   • Path matcher disjointness for sub-projects                              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                        analyze_histories()
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Repository                                        │
│                                                                              │
│   • Git repository wrapper (libgit2)                                        │
│   • Commit history analysis per project                                     │
│   • Path relevance matching                                                  │
│   • Release tag detection                                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Ecosystem Loader Flow

Each ecosystem loader follows the same pattern:

```
1. Repository Index Scan
   └── For each file in git index:
       └── process_index_item(repopath, dirname, basename)
           │
           ├── Match manifest file? (Cargo.toml, package.json, etc.)
           │   └── No  → return Ok(())
           │   └── Yes → Continue
           │
           ├── Parse manifest file
           │
           ├── Extract project metadata:
           │   • name (qualified names for disambiguation)
           │   • version
           │   • prefix (directory in repo)
           │   • dependencies (internal to repo)
           │
           ├── Register with graph.try_add_project(qnames, pconfig)
           │   └── Returns ProjectId or None (if ignored)
           │
           ├── Configure rewriters for version updates
           │
           └── Track for dependency resolution
```

## Supported Ecosystems

| Ecosystem | Manifest File      | Version Location            | Workspace Support |
|-----------|--------------------|-----------------------------|-------------------|
| Cargo     | Cargo.toml         | `[package].version`         | Yes               |
| NPM       | package.json       | `version` field             | Yes (Lerna)       |
| PyPA      | setup.cfg/pyproject.toml | `[metadata].version`  | Limited           |
| Go        | go.mod             | Module path                 | No                |
| Elixir    | mix.exs            | `@version` attribute        | Yes (umbrella)    |
| C#        | *.csproj           | `<Version>` element         | Yes               |

## Project Registration

Projects are identified by qualified names that enable disambiguation:

```
Qualified Names: ["my-package", "npm"]
                 ["my-package", "cargo"]

User-Facing Names (computed):
  • "npm:my-package"
  • "cargo:my-package"
```

The naming algorithm progressively adds qualifiers until all names are unique.

## Dependency Graph

Internal dependencies form a DAG (Directed Acyclic Graph):

```
         ┌─────────┐
         │ common  │ ◄── Leaf node (no dependencies)
         └────┬────┘
              │
    ┌─────────┴─────────┐
    ▼                   ▼
┌─────────┐       ┌─────────┐
│   api   │       │   web   │
└────┬────┘       └────┬────┘
     │                 │
     └────────┬────────┘
              ▼
         ┌─────────┐
         │   app   │ ◄── Root node (depends on others)
         └─────────┘

Topological sort: common → api → web → app
```

## Version Requirements

Dependencies can specify requirements in three ways:

1. **Commit-based** (`DepRequirement::Commit`): Version after a specific commit
2. **Manual** (`DepRequirement::Manual`): User-specified semver requirement
3. **Unavailable** (`DepRequirement::Unavailable`): Missing Clikd metadata

## Path Matching

Each project has a `PathMatcher` that determines which repository paths affect it:

```rust
Project "common" → prefix: "packages/common/"
Project "api"    → prefix: "packages/api/"

File "packages/common/src/lib.rs" → affects "common"
File "packages/api/src/main.rs"   → affects "api"
```

Sub-projects are automatically excluded from parent matchers:

```
packages/
├── parent/          ← parent project
│   ├── Cargo.toml
│   └── child/       ← child project
│       └── Cargo.toml

Changes in packages/parent/child/ affect "child", NOT "parent"
```

## Commit Analysis

The repository analyzer processes git history:

```
1. Walk commits from HEAD to last release tag
2. For each commit:
   a. Parse conventional commit message
   b. Determine affected projects via path matching
   c. Categorize change type (feat, fix, breaking, etc.)
   d. Compute version bump recommendation
3. Generate per-project changelogs
```

## Data Flow Summary

```
Git Repository
     │
     ▼
┌──────────────────┐
│  Index Scan      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐      ┌──────────────────┐
│  Cargo Loader    │─────►│                  │
├──────────────────┤      │                  │
│  NPM Loader      │─────►│  ProjectGraph    │
├──────────────────┤      │     Builder      │
│  PyPA Loader     │─────►│                  │
├──────────────────┤      │                  │
│  Go Loader       │─────►│                  │
└──────────────────┘      └────────┬─────────┘
                                   │
                          complete_loading()
                                   │
                                   ▼
                          ┌──────────────────┐
                          │                  │
                          │  ProjectGraph    │
                          │     (DAG)        │
                          │                  │
                          └────────┬─────────┘
                                   │
                          analyze_histories()
                                   │
                                   ▼
                          ┌──────────────────┐
                          │  RepoHistories   │
                          │  (per-project    │
                          │   commit data)   │
                          └──────────────────┘
```

## Key Files

| File | Purpose |
|------|---------|
| `src/core/release/session.rs` | AppBuilder, AppSession initialization |
| `src/core/release/graph.rs` | ProjectGraph, DAG management |
| `src/core/release/project.rs` | Project, Dependency structures |
| `src/core/release/repository.rs` | Git operations, history analysis |
| `src/core/ecosystem/cargo.rs` | Rust/Cargo project loader |
| `src/core/ecosystem/npm.rs` | NPM/Node project loader |
| `src/core/ecosystem/pypa.rs` | Python project loader |
| `src/core/release/commit_analyzer.rs` | Conventional commit parsing |
| `src/core/release/changelog_generator.rs` | Changelog generation |
