# clikd-bot: GitHub App + Dashboard

## Übersicht

Die GitHub App automatisiert Releases nach PR-Merge und bietet ein Dashboard für Release-Metriken.

```
┌─────────────────────────────────────────────────────────────────┐
│                         clikd-bot                               │
│                     (Shuttle.rs hosted)                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  🔗 GitHub App                                                  │
│  ├── Webhooks: pull_request.closed (merged)                    │
│  ├── Permissions: Contents, PRs, Issues, Releases              │
│  └── OAuth: GitHub Login für Dashboard                         │
│                                                                 │
│  🎯 Core Automation                                             │
│  ├── Release nach Merge (Tags + GitHub Releases)               │
│  ├── Changelog Preview auf PRs                                 │
│  ├── Impact Analysis für Monorepos                             │
│  └── Auto-Labeling                                             │
│                                                                 │
│  📊 Dashboard (Dioxus Fullstack)                                │
│  ├── Release Overview & Metrics                                │
│  ├── Pending Releases                                          │
│  ├── Contributor Stats                                         │
│  └── Settings pro Repo                                         │
│                                                                 │
│  🔌 Integrations                                                │
│  ├── Slack/Discord Webhooks                                    │
│  ├── clikd CLI (sendet Manifest-Daten)                         │
│  └── GitHub Actions (Fallback)                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technologie | Beschreibung |
|-------|-------------|--------------|
| **Fullstack** | Dioxus 0.6+ | Frontend + Backend in einem (Axum intern) |
| **Hosting** | Shuttle.rs | Rust-native, einfaches Deployment |
| **Database** | Turso | SQLite Edge, global repliziert |
| **Auth** | GitHub OAuth | Via GitHub App Installation |

### Warum Dioxus Fullstack?

- **Ein Framework** für alles (kein separates Axum Setup nötig)
- **Server Functions** werden automatisch zu Axum Handlers
- **Built-in**: WebSockets, SSE, Streaming, SSR, Forms, Hot-Reload
- **Type-safe RPC** zwischen Frontend und Backend
- **Cross-Platform** möglich (Web, Desktop, Mobile)

```rust
// Server Function - läuft auf dem Server
#[server]
async fn get_pending_releases(org: String) -> Result<Vec<Release>, ServerFnError> {
    let releases = db::fetch_pending_releases(&org).await?;
    Ok(releases)
}

// Frontend - ruft Server Function direkt auf
fn PendingReleases(org: String) -> Element {
    let releases = use_server_future(move || get_pending_releases(org.clone()))?;

    rsx! {
        for release in releases() {
            ReleaseCard { release }
        }
    }
}
```

---

## Projekt-Struktur

```
clikd-bot/
├── Cargo.toml
├── Shuttle.toml
├── src/
│   ├── main.rs                 # Shuttle + Dioxus Entry
│   ├── lib.rs
│   │
│   ├── app/                    # Dioxus Frontend
│   │   ├── mod.rs
│   │   ├── routes/
│   │   │   ├── dashboard.rs
│   │   │   ├── releases.rs
│   │   │   ├── settings.rs
│   │   │   └── login.rs
│   │   └── components/
│   │       ├── release_card.rs
│   │       ├── metrics_chart.rs
│   │       ├── pending_list.rs
│   │       └── nav.rs
│   │
│   ├── server/                 # Server Functions + Webhooks
│   │   ├── mod.rs
│   │   ├── webhooks/
│   │   │   ├── mod.rs
│   │   │   ├── pull_request.rs
│   │   │   └── installation.rs
│   │   ├── functions/
│   │   │   ├── releases.rs
│   │   │   ├── metrics.rs
│   │   │   └── settings.rs
│   │   └── github/
│   │       ├── client.rs
│   │       ├── auth.rs
│   │       └── api.rs
│   │
│   ├── db/                     # Turso/SQLite
│   │   ├── mod.rs
│   │   ├── schema.rs
│   │   ├── releases.rs
│   │   └── installations.rs
│   │
│   └── models/
│       ├── mod.rs
│       ├── release.rs
│       ├── manifest.rs
│       └── installation.rs
│
├── migrations/
│   └── 001_initial.sql
│
└── assets/
    └── styles.css
```

---

## Phasen

### Phase 1: CLI (✅ ABGESCHLOSSEN)

- [x] PR-Based Workflow
- [x] Release Manifest in `clikd/releases/*.json`
- [x] Branch `release/YYYYMMDD-HHMMSS`
- [x] PR via GitHub API
- [x] Changelog Generation

### Phase 2: GitHub App Core

**Ziel:** Automatische Releases nach PR-Merge

```
PR merged → Webhook → Parse Manifest → Create Tags → Create GitHub Releases
```

| Task | Beschreibung |
|------|--------------|
| GitHub App erstellen | App Registration auf github.com |
| Webhook Endpoint | `POST /webhooks/github` |
| Manifest Parser | `clikd/releases/*.json` lesen |
| Tag Creation | Git Tags via GitHub API |
| Release Creation | GitHub Releases mit Changelog |
| Manifest Cleanup | Datei nach Release löschen |

**Webhook Handler:**

```rust
#[server]
async fn handle_pr_webhook(payload: PullRequestEvent) -> Result<(), ServerFnError> {
    if payload.action != "closed" || !payload.pull_request.merged {
        return Ok(());
    }

    let manifests = github::get_release_manifests(&payload.repository).await?;

    for manifest in manifests {
        for release in manifest.releases {
            github::create_tag(&release).await?;
            github::create_release(&release).await?;
        }
        github::delete_manifest_file(&manifest.path).await?;
    }

    Ok(())
}
```

### Phase 3: Changelog Preview + Labels

**Ziel:** Bot kommentiert auf PRs mit Release-Preview

| Task | Beschreibung |
|------|--------------|
| PR Comment Bot | Preview des Changelogs als Kommentar |
| Auto-Labels | `release:major`, `release:minor`, `release:patch` |
| Breaking Change Warning | ⚠️ Alert bei BREAKING CHANGE |
| Impact Analysis | "Dieser PR betrifft: pkg-a, pkg-b" |

### Phase 4: Dashboard Basic

**Ziel:** Web UI für Release-Übersicht

```
┌─────────────────────────────────────────────────────────────────┐
│  🚀 clikd Dashboard                              [org-selector] │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  📊 Release Overview (letzte 30 Tage)                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐              │
│  │   47    │ │   12    │ │  3.2d   │ │   98%   │              │
│  │Releases │ │ Projekte│ │Avg Time │ │ Success │              │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘              │
│                                                                 │
│  📈 Release Timeline                                           │
│  [═══════════════════════════════════════════════════════]     │
│                                                                 │
│  📦 Recent Releases                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ rig v1.2.0        │ 2h ago  │ 🟢 Published │ [View]      │  │
│  │ requip v2.0.0     │ 2h ago  │ 🟢 Published │ [View]      │  │
│  │ mondo v3.1.0      │ 5d ago  │ 🟢 Published │ [View]      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  🔄 Pending Releases (PRs mit Release-Manifests)               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ PR #142: Release gate, jiji │ Awaiting Review │ [View]   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

| Task | Beschreibung |
|------|--------------|
| GitHub OAuth | Login via GitHub |
| Org/Repo Selector | Multi-Org Support |
| Release List | Recent + Pending Releases |
| Basic Metrics | Count, Avg Time, Success Rate |

### Phase 5: Dashboard Advanced

**Ziel:** Metriken, Notifications, Settings

| Feature | Beschreibung |
|---------|--------------|
| DORA Metrics | Lead Time, Deployment Frequency |
| Release Timeline | Visualisierung über Zeit |
| Contributor Stats | Wer hat zu welchen Releases beigetragen |
| Slack/Discord | Notifications bei Release |
| Repo Settings | Per-Repo Konfiguration |

---

## Database Schema (Turso)

```sql
-- GitHub App Installations
CREATE TABLE installations (
    id INTEGER PRIMARY KEY,
    github_installation_id INTEGER UNIQUE NOT NULL,
    account_login TEXT NOT NULL,
    account_type TEXT NOT NULL, -- 'User' or 'Organization'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tracked Repositories
CREATE TABLE repositories (
    id INTEGER PRIMARY KEY,
    installation_id INTEGER REFERENCES installations(id),
    github_repo_id INTEGER UNIQUE NOT NULL,
    full_name TEXT NOT NULL, -- 'owner/repo'
    default_branch TEXT DEFAULT 'main',
    settings JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Releases
CREATE TABLE releases (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER REFERENCES repositories(id),
    package_name TEXT NOT NULL,
    version TEXT NOT NULL,
    bump_type TEXT NOT NULL,
    changelog TEXT,
    tag_name TEXT,
    github_release_id INTEGER,
    pr_number INTEGER,
    created_by TEXT,
    released_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Release Metrics (aggregated)
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER REFERENCES repositories(id),
    date DATE NOT NULL,
    release_count INTEGER DEFAULT 0,
    avg_lead_time_hours REAL,
    success_rate REAL,
    UNIQUE(repository_id, date)
);
```

---

## Cargo.toml

```toml
[package]
name = "clikd-bot"
version = "0.1.0"
edition = "2021"

[dependencies]
dioxus = { version = "0.6", features = ["fullstack", "router"] }
shuttle-runtime = "0.49"
shuttle-turso = "0.49"
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
libsql = "0.6"
hmac = "0.12"
sha2 = "0.10"
octocrab = "0.41"
chrono = { version = "0.4", features = ["serde"] }
tracing = "0.1"

[features]
default = ["web"]
server = ["dioxus/server"]
web = ["dioxus/web"]
```

---

## Nächste Schritte

1. **GitHub App Registration**
   - App auf github.com/settings/apps erstellen
   - Webhook URL: `https://clikd-bot.shuttleapp.rs/webhooks/github`
   - Permissions: Contents (read/write), Pull Requests (read), Issues (read/write)

2. **Projekt Setup**
   ```bash
   cargo shuttle init clikd-bot
   cd clikd-bot
   cargo add dioxus --features fullstack,router
   cargo add shuttle-turso octocrab serde serde_json
   ```

3. **Webhook Handler implementieren**

4. **Dashboard UI bauen**

---

## Links

- [Dioxus Docs](https://dioxuslabs.com/learn/0.6/)
- [Shuttle.rs Docs](https://docs.shuttle.rs/)
- [Turso Docs](https://docs.turso.tech/)
- [GitHub Apps Docs](https://docs.github.com/en/apps)
- [Octocrab (GitHub API Client)](https://docs.rs/octocrab)
