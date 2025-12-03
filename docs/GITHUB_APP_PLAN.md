# Clikd GitHub App - Implementierungsplan

## Übersicht

Eine GitHub App die automatisch Release-PRs erstellt und GitHub Releases publiziert.
Nutzt den gemeinsamen `clikd-core` Code für konsistentes Verhalten mit der CLI.

## Technologie-Stack

| Komponente | Technologie | Begründung |
|------------|-------------|------------|
| **Backend** | Rust + Axum | Shared code mit CLI |
| **Hosting** | Shuttle.rs | Native Rust, managed, einfach |
| **Database** | PostgreSQL (Shuttle) | State, Logs, Analytics |
| **Frontend** | SvelteKit | Schnell, modern, SSR |
| **Auth** | GitHub OAuth | Native Integration |

## Architektur

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIKD WORKSPACE                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  crates/                                                         │
│  ├── clikd-core/           ← Shared Library                     │
│  │   ├── analysis/         - Commit Analysis                    │
│  │   ├── changelog/        - Changelog Generation               │
│  │   ├── ai/               - Claude Integration                 │
│  │   ├── version/          - Version Parsing & Bumping          │
│  │   ├── ecosystem/        - Cargo, NPM, PyPA, Go, etc.         │
│  │   └── github/           - GitHub API Client                  │
│  │                                                               │
│  ├── clikd-cli/            ← CLI Binary                         │
│  │   └── (uses clikd-core)                                      │
│  │                                                               │
│  └── clikd-app/            ← GitHub App (Shuttle)               │
│      ├── src/                                                    │
│      │   ├── main.rs       - Shuttle entrypoint                 │
│      │   ├── webhooks/     - GitHub webhook handlers            │
│      │   ├── api/          - REST API for dashboard             │
│      │   ├── jobs/         - Background job processing          │
│      │   └── db/           - Database models                    │
│      └── web/              - SvelteKit Dashboard                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Datenbank Schema

```sql
-- Installierte Repositories
CREATE TABLE installations (
    id SERIAL PRIMARY KEY,
    github_installation_id BIGINT UNIQUE NOT NULL,
    github_account_login TEXT NOT NULL,
    github_account_type TEXT NOT NULL, -- 'User' or 'Organization'
    created_at TIMESTAMP DEFAULT NOW(),
    settings JSONB DEFAULT '{}'
);

-- Repositories unter einer Installation
CREATE TABLE repositories (
    id SERIAL PRIMARY KEY,
    installation_id INT REFERENCES installations(id),
    github_repo_id BIGINT UNIQUE NOT NULL,
    full_name TEXT NOT NULL, -- 'owner/repo'
    default_branch TEXT DEFAULT 'main',
    config JSONB DEFAULT '{}',
    last_analyzed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Release PRs
CREATE TABLE release_prs (
    id SERIAL PRIMARY KEY,
    repository_id INT REFERENCES repositories(id),
    pr_number INT NOT NULL,
    status TEXT DEFAULT 'open', -- 'open', 'merged', 'closed'
    packages JSONB NOT NULL, -- [{name, from, to, bump_type}]
    changelog_content TEXT,
    ai_enhanced BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP,
    UNIQUE(repository_id, pr_number)
);

-- Veröffentlichte Releases
CREATE TABLE releases (
    id SERIAL PRIMARY KEY,
    repository_id INT REFERENCES repositories(id),
    release_pr_id INT REFERENCES release_prs(id),
    package_name TEXT NOT NULL,
    version TEXT NOT NULL,
    github_release_id BIGINT,
    tag_name TEXT NOT NULL,
    published_at TIMESTAMP DEFAULT NOW()
);

-- Audit Log
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    repository_id INT REFERENCES repositories(id),
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Webhook Events

| Event | Action | App Response |
|-------|--------|--------------|
| `installation` | created | Store installation, scan repos |
| `installation` | deleted | Cleanup data |
| `installation_repositories` | added/removed | Update repo list |
| `push` | (to default branch) | Analyze & create/update Release PR |
| `pull_request` | closed + merged | If Release PR → publish releases |
| `pull_request` | edited | If Release PR → update internal state |

## API Endpoints

### Public API (für Dashboard)

```
GET  /api/installations           - List user's installations
GET  /api/repos                   - List repos for installation
GET  /api/repos/:owner/:repo      - Get repo details + pending releases
POST /api/repos/:owner/:repo/analyze  - Trigger manual analysis
GET  /api/repos/:owner/:repo/releases - Release history
PATCH /api/repos/:owner/:repo/config  - Update repo config
```

### Webhook Endpoint

```
POST /webhooks/github             - GitHub webhook receiver
```

### OAuth Endpoints

```
GET  /auth/github                 - Start GitHub OAuth
GET  /auth/github/callback        - OAuth callback
POST /auth/logout                 - Logout
GET  /auth/me                     - Current user
```

## Shuttle.rs Setup

```rust
// clikd-app/src/main.rs
use axum::{routing::{get, post}, Router};
use shuttle_axum::ShuttleAxum;
use shuttle_runtime::SecretStore;
use shuttle_shared_db::Postgres;
use sqlx::PgPool;

#[shuttle_runtime::main]
async fn main(
    #[shuttle_shared_db::Postgres] pool: PgPool,
    #[shuttle_runtime::Secrets] secrets: SecretStore,
) -> ShuttleAxum {
    // Run migrations
    sqlx::migrate!().run(&pool).await?;

    // Get secrets
    let github_app_id = secrets.get("GITHUB_APP_ID").unwrap();
    let github_private_key = secrets.get("GITHUB_PRIVATE_KEY").unwrap();
    let github_webhook_secret = secrets.get("GITHUB_WEBHOOK_SECRET").unwrap();
    let anthropic_api_key = secrets.get("ANTHROPIC_API_KEY").ok();

    // Build app state
    let state = AppState::new(pool, github_app_id, github_private_key, anthropic_api_key);

    // Build router
    let app = Router::new()
        // Webhooks
        .route("/webhooks/github", post(webhooks::github_handler))
        // API
        .route("/api/installations", get(api::list_installations))
        .route("/api/repos", get(api::list_repos))
        .route("/api/repos/:owner/:repo", get(api::get_repo))
        .route("/api/repos/:owner/:repo/analyze", post(api::trigger_analysis))
        .route("/api/repos/:owner/:repo/releases", get(api::list_releases))
        .route("/api/repos/:owner/:repo/config", patch(api::update_config))
        // Auth
        .route("/auth/github", get(auth::start_oauth))
        .route("/auth/github/callback", get(auth::oauth_callback))
        .route("/auth/logout", post(auth::logout))
        .route("/auth/me", get(auth::current_user))
        // Static files (SvelteKit build)
        .nest_service("/", ServeDir::new("web/build"))
        .with_state(state);

    Ok(app.into())
}
```

## Webhook Handler Flow

```rust
// clikd-app/src/webhooks/push.rs
pub async fn handle_push(
    state: &AppState,
    payload: PushPayload,
) -> Result<(), AppError> {
    // 1. Check if push is to default branch
    if !payload.is_default_branch() {
        return Ok(());
    }

    // 2. Get repository config
    let repo = state.db.get_repository(payload.repository.id).await?;
    let config = repo.config.merge_with_defaults();

    // 3. Clone repo (shallow) to temp dir
    let temp_dir = tempfile::tempdir()?;
    let token = state.github.get_installation_token(repo.installation_id).await?;
    clone_repo(&payload.repository.clone_url, &temp_dir, &token).await?;

    // 4. Use clikd-core for analysis
    let analysis = clikd_core::analyze_repository(&temp_dir, &config).await?;

    if analysis.packages_to_release.is_empty() {
        return Ok(()); // Nothing to release
    }

    // 5. Generate changelogs (with AI if enabled)
    let changelogs = if config.changelog.ai_enabled {
        clikd_core::generate_ai_changelogs(&analysis, &state.anthropic_client).await?
    } else {
        clikd_core::generate_changelogs(&analysis)?
    };

    // 6. Check for existing Release PR
    let existing_pr = state.db.get_open_release_pr(repo.id).await?;

    // 7. Create or update Release PR
    let pr = if let Some(pr) = existing_pr {
        update_release_pr(&state.github, &repo, &pr, &analysis, &changelogs).await?
    } else {
        create_release_pr(&state.github, &repo, &analysis, &changelogs).await?
    };

    // 8. Store in database
    state.db.upsert_release_pr(&pr).await?;

    // 9. Log event
    state.db.log_event(repo.id, "release_pr_updated", &pr).await?;

    Ok(())
}
```

## Release PR Erstellung

```rust
// clikd-app/src/github/pr.rs
pub async fn create_release_pr(
    github: &GitHubClient,
    repo: &Repository,
    analysis: &ReleaseAnalysis,
    changelogs: &HashMap<String, String>,
) -> Result<ReleasePr, AppError> {
    // 1. Create branch
    let branch_name = format!("clikd/release-{}", chrono::Utc::now().format("%Y%m%d"));
    github.create_branch(&repo.full_name, &branch_name, &analysis.base_sha).await?;

    // 2. Apply changes to branch
    for package in &analysis.packages_to_release {
        // Update version files
        let version_changes = clikd_core::generate_version_changes(package)?;
        for (path, content) in version_changes {
            github.update_file(&repo.full_name, &branch_name, &path, &content).await?;
        }

        // Update CHANGELOG.md
        let changelog_path = format!("{}/CHANGELOG.md", package.prefix);
        let changelog_content = changelogs.get(&package.name).unwrap();
        github.update_file(&repo.full_name, &branch_name, &changelog_path, changelog_content).await?;
    }

    // 3. Create PR
    let title = format_pr_title(&analysis.packages_to_release);
    let body = format_pr_body(&analysis, changelogs);

    let pr = github.create_pull_request(&repo.full_name, CreatePrRequest {
        title,
        body,
        head: branch_name,
        base: repo.default_branch.clone(),
        labels: vec!["release".into(), "automated".into()],
    }).await?;

    Ok(ReleasePr {
        pr_number: pr.number,
        packages: analysis.packages_to_release.clone(),
        changelog_content: body,
        ai_enhanced: changelogs.values().any(|c| c.contains("AI")),
    })
}

fn format_pr_body(analysis: &ReleaseAnalysis, changelogs: &HashMap<String, String>) -> String {
    let mut body = String::new();

    body.push_str("## 🚀 Release\n\n");
    body.push_str("This PR was automatically created by [Clikd](https://clikd.dev).\n\n");

    // Package table
    body.push_str("### 📦 Packages\n\n");
    body.push_str("| Package | Current | Next | Type |\n");
    body.push_str("|---------|---------|------|------|\n");
    for pkg in &analysis.packages_to_release {
        body.push_str(&format!(
            "| {} | {} | {} | {} |\n",
            pkg.name, pkg.current_version, pkg.next_version, pkg.bump_type
        ));
    }
    body.push_str("\n");

    // Changelogs
    body.push_str("### 📝 Changelogs\n\n");
    for (name, changelog) in changelogs {
        body.push_str(&format!("<details>\n<summary>{}</summary>\n\n", name));
        body.push_str(changelog);
        body.push_str("\n</details>\n\n");
    }

    // Footer
    body.push_str("---\n");
    body.push_str("🤖 *Merge this PR to publish releases.*\n");

    body
}
```

## Release Publishing (nach PR Merge)

```rust
// clikd-app/src/webhooks/pull_request.rs
pub async fn handle_pr_merged(
    state: &AppState,
    payload: PullRequestPayload,
) -> Result<(), AppError> {
    // 1. Check if this is a Release PR
    let release_pr = match state.db.get_release_pr_by_number(
        payload.repository.id,
        payload.pull_request.number,
    ).await? {
        Some(pr) => pr,
        None => return Ok(()), // Not a release PR
    };

    // 2. Create GitHub Releases for each package
    for package in &release_pr.packages {
        let tag_name = format!("{}-v{}", package.name, package.next_version);

        // Get changelog for this package
        let changelog = extract_changelog_for_package(&release_pr.changelog_content, &package.name);

        // Create tag
        state.github.create_tag(
            &payload.repository.full_name,
            &tag_name,
            &payload.pull_request.merge_commit_sha,
        ).await?;

        // Create release
        let release = state.github.create_release(
            &payload.repository.full_name,
            CreateReleaseRequest {
                tag_name: tag_name.clone(),
                name: format!("{} v{}", package.name, package.next_version),
                body: changelog,
                draft: false,
                prerelease: package.next_version.contains("-"),
            },
        ).await?;

        // Store in database
        state.db.create_release(Release {
            repository_id: payload.repository.id,
            release_pr_id: release_pr.id,
            package_name: package.name.clone(),
            version: package.next_version.clone(),
            github_release_id: release.id,
            tag_name,
        }).await?;
    }

    // 3. Update Release PR status
    state.db.update_release_pr_status(release_pr.id, "merged").await?;

    // 4. Log event
    state.db.log_event(
        payload.repository.id,
        "releases_published",
        &release_pr.packages,
    ).await?;

    Ok(())
}
```

## Dashboard (SvelteKit)

```
clikd-app/web/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte      - Main layout with nav
│   │   ├── +page.svelte        - Dashboard home
│   │   ├── repos/
│   │   │   ├── +page.svelte    - Repository list
│   │   │   └── [owner]/
│   │   │       └── [repo]/
│   │   │           ├── +page.svelte     - Repo details
│   │   │           ├── releases/
│   │   │           │   └── +page.svelte - Release history
│   │   │           └── settings/
│   │   │               └── +page.svelte - Repo settings
│   │   └── auth/
│   │       └── callback/
│   │           └── +page.svelte
│   ├── lib/
│   │   ├── api.ts              - API client
│   │   ├── components/         - Reusable components
│   │   └── stores/             - Svelte stores
│   └── app.html
├── static/
├── svelte.config.js
└── package.json
```

## Deployment

### Shuttle.rs Konfiguration

```toml
# Shuttle.toml
name = "clikd-app"
assets = ["web/build/*"]

[build]
# Build SvelteKit before Rust
pre_build = "cd web && npm install && npm run build"
```

### Secrets (via Shuttle Console)

```
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----...
GITHUB_WEBHOOK_SECRET=whsec_...
GITHUB_CLIENT_ID=Iv1.abc123
GITHUB_CLIENT_SECRET=...
ANTHROPIC_API_KEY=sk-ant-...  # Optional
```

### Deploy Commands

```bash
# Install Shuttle CLI
cargo install cargo-shuttle

# Login
cargo shuttle login

# Deploy
cd clikd-app
cargo shuttle deploy
```

## GitHub App Registration

1. Gehe zu https://github.com/settings/apps/new
2. Konfiguriere:
   - **Name**: Clikd Release Manager
   - **Homepage URL**: https://clikd.dev
   - **Webhook URL**: https://clikd-app.shuttleapp.rs/webhooks/github
   - **Webhook Secret**: (generieren und in Shuttle Secrets speichern)

3. Permissions:
   - **Repository permissions**:
     - Contents: Read & Write
     - Metadata: Read
     - Pull requests: Read & Write
   - **Organization permissions**:
     - Members: Read (optional, für Team-Features)

4. Events:
   - Installation
   - Push
   - Pull request

5. Nach Erstellung:
   - App ID notieren
   - Private Key generieren und herunterladen
   - In Shuttle Secrets speichern

## Phasen-Plan

### Phase 1: Core Library Extraktion (1 Woche)
- [ ] `clikd-core` Crate erstellen
- [ ] Code aus CLI extrahieren
- [ ] Public API definieren
- [ ] Tests migrieren

### Phase 2: CLI --ci Mode (1 Woche)
- [ ] `--ci` Flag implementieren
- [ ] Changelog in Non-TUI Mode
- [ ] Auto-Commit + Tags
- [ ] GitHub Release Creation
- [ ] Testen

### Phase 3: App Backend Basics (2 Wochen)
- [ ] Shuttle.rs Projekt Setup
- [ ] Database Schema + Migrations
- [ ] GitHub App Registration
- [ ] Webhook Handler (Installation, Push)
- [ ] Basic Analysis Flow

### Phase 4: Release PR Creation (1 Woche)
- [ ] Branch Creation
- [ ] File Updates via GitHub API
- [ ] PR Creation mit Changelog
- [ ] PR Update bei neuen Commits

### Phase 5: Release Publishing (1 Woche)
- [ ] PR Merge Detection
- [ ] Tag Creation
- [ ] GitHub Release Creation
- [ ] Database Updates

### Phase 6: Dashboard UI (2 Wochen)
- [ ] SvelteKit Setup
- [ ] OAuth Flow
- [ ] Repository List
- [ ] Repo Details + Pending Releases
- [ ] Settings Page

### Phase 7: Polish & Launch (1 Woche)
- [ ] Error Handling
- [ ] Rate Limiting
- [ ] Logging & Monitoring
- [ ] Documentation
- [ ] Beta Launch

## Kosten-Schätzung

| Service | Free Tier | Paid |
|---------|-----------|------|
| Shuttle.rs | 3 projects, shared resources | $20/mo pro project |
| PostgreSQL | Included in Shuttle | - |
| GitHub App | Free | Free |
| Claude API | - | ~$0.01 per changelog |

**Für Start: Shuttle Free Tier reicht völlig aus.**

## Unique Selling Points vs Konkurrenz

| Feature | Release Please | Semantic Release | Changesets | **Clikd** |
|---------|---------------|------------------|------------|-----------|
| Multi-Language | ⚠️ Limited | ⚠️ JS-fokussiert | ❌ JS only | ✅ |
| Monorepo | ✅ | ⚠️ Plugin | ✅ | ✅ |
| AI Changelogs | ❌ | ❌ | ❌ | ✅ |
| Local TUI | ❌ | ❌ | ❌ | ✅ |
| Web Dashboard | ❌ | ❌ | ❌ | ✅ |
| GitHub App | ❌ | ❌ | ❌ | ✅ |
| Changelog Editor | ❌ | ❌ | ❌ | ✅ |
