# CLI Refactoring Plan: PR-Based Release Workflow

## Übersicht

Transformation von `clikd release prepare` von Direct-Commit zu PR-basiertem Workflow.

**Aktueller Flow:**
```
clikd release prepare → bump versions → commit → tags → user pusht
```

**Neuer Flow:**
```
clikd release prepare → branch erstellen → bump versions → manifest erstellen → PR via GitHub API
                                                                              ↓
                                                    GitHub App (nach merge) → tags + GitHub releases
```

---

## Phase 1: Bug Fixes

### 1.1 Dynamische Datei-Anzeige im Confirmation Screen

**Problem:** `src/cmd/release/prepare/wizard.rs:979-981` zeigt hardcodierte Dateiliste.

**Lösung:**
1. `EcosystemType` enum in `src/core/ecosystem/mod.rs` erstellen
2. `project_type: EcosystemType` Feld zu `ProjectItem` hinzufügen
3. `render_confirmation()` dynamisch basierend auf ausgewählten Projekt-Typen

### 1.2 AI Preview im Changelog View

**Problem:** `render_project_changelog()` zeigt nur Standard-Kategorisierung. AI-Polishing passiert erst nach Confirmation.

**Voraussetzung:** AI ist in Config aktiviert:
```toml
[release.changelog]
ai_enabled = true
```

**Lösung:**
1. `ai_changelog_cache: HashMap<ProjectId, String>` zu `WizardState` hinzufügen
2. `ai_enabled: bool` Flag im WizardState (aus Config gelesen)
3. AI-Generierung nur triggern wenn `ai_enabled == true` UND Changelog-View betreten wird
4. Loading-Indikator zeigen während AI arbeitet
5. Bei `ai_enabled == false`: Standard-Kategorisierung wie bisher

---

## Phase 2: PR-Based Workflow

### 2.1 Release Manifest Struktur

**Neue Datei:** `src/core/release/manifest.rs`

```rust
pub struct ReleaseManifest {
    pub schema_version: String,
    pub created_at: DateTime<Utc>,
    pub created_by: String,
    pub base_branch: String,
    pub releases: Vec<ProjectRelease>,
}

pub struct ProjectRelease {
    pub name: String,
    pub ecosystem: String,
    pub previous_version: String,
    pub new_version: String,
    pub bump_type: String,
    pub changelog: String,
    pub tag_name: String,
    pub prefix: String,
}
```

Manifest wird in `clikd/releases/release-YYYYMMDD-HHMMSS.json` gespeichert.

### 2.2 GitHub API Integration (Bestehenden Client erweitern)

**Datei:** `src/core/github/client.rs`

Der bestehende `GitHubInformation` Client nutzt `GITHUB_TOKEN` Environment Variable.
Das Auth-System (`clikd login`) speichert den Token aber im Keyring via `token::save_token()`.

**Lösung:** `GitHubInformation::new()` erweitern:

```rust
impl GitHubInformation {
    fn new(sess: &AppSession) -> Result<Self> {
        // Erst Keyring-Token versuchen, dann Environment Variable als Fallback
        let token = match crate::core::auth::token::load_token() {
            Ok(t) => t,
            Err(_) => require_var("GITHUB_TOKEN")?,
        };
        // ... rest bleibt gleich
    }
}
```

**Neue Methode für PR-Erstellung hinzufügen:**

```rust
impl GitHubInformation {
    fn create_pull_request(
        &self,
        head: &str,
        base: &str,
        title: &str,
        body: &str,
        client: &mut reqwest::blocking::Client,
    ) -> Result<String> {
        let pr_info = object! {
            "title" => title,
            "head" => head,
            "base" => base,
            "body" => body,
        };

        let create_url = self.api_url("pulls");
        let resp = client.post(create_url).body(json::stringify(pr_info)).send()?;

        if resp.status().is_success() {
            let parsed = json::parse(&resp.text()?)?;
            Ok(parsed["html_url"].to_string())
        } else {
            Err(anyhow!("failed to create PR: {}", resp.text()?))
        }
    }
}
```

### 2.3 Neuer Workflow in `run()`

1. Release-Branch erstellen: `release/YYYYMMDD-HHMMSS`
2. Versionen bumpen (wie bisher)
3. Changelogs schreiben (wie bisher)
4. Manifest-Datei erstellen in `clikd/releases/`
5. Commit erstellen (OHNE Tags!)
6. Branch pushen
7. PR via GitHub API erstellen
8. PR-URL ausgeben

---

## Phase 3: Entfernung von Tag/Release-Erstellung

### 3.1 Zu entfernende Funktionen

**`src/cmd/release/prepare/wizard.rs`:**
- Zeilen 599-603: `create_release_tags()` Aufruf entfernen

**`src/cmd/release/prepare.rs`:**
- `create_github_release()` Funktion (Zeilen 556-610) entfernen
- Alle zugehörigen Imports

**`src/core/release/repository.rs`:**
- `create_release_tags()` Methode kann bleiben (für andere Use Cases), aber wird nicht mehr von prepare aufgerufen

---

## Phase 4: Repository Git-Methoden

**Datei:** `src/core/release/repository.rs`

Neue Methoden hinzufügen:
- `current_branch() -> Result<String>`
- `create_branch(name: &str) -> Result<()>`
- `checkout_branch(name: &str) -> Result<()>`
- `push_branch(name: &str) -> Result<()>`

---

## Datei-Änderungen Übersicht

| Datei | Aktion | Beschreibung |
|-------|--------|--------------|
| `src/core/ecosystem/mod.rs` | NEW | `EcosystemType` enum |
| `src/core/release/manifest.rs` | NEW | Release Manifest Strukturen |
| `src/core/github/client.rs` | MODIFY | Token-Fallback + `create_pull_request()` |
| `src/cmd/release/prepare/wizard.rs` | MODIFY | Hauptänderungen (PR Workflow) |
| `src/core/release/repository.rs` | MODIFY | Git Branch-Methoden |
| `src/cmd/release/prepare.rs` | MODIFY | `create_github_release` entfernen |
| `src/lib.rs` | MODIFY | `ecosystem/mod.rs` exportieren |

---

## Abhängigkeiten

✅ `reqwest` ist bereits vorhanden in `Cargo.toml`:
```toml
reqwest = { version = "0.12", features = ["json", "rustls-tls", "blocking"] }
```

Keine neuen Dependencies erforderlich.

---

## Implementierungs-Reihenfolge

1. EcosystemType enum + dynamische Datei-Anzeige
2. AI Preview im Changelog
3. Release Manifest Struktur
4. Repository Git-Methoden
5. GitHub API Client
6. PR Workflow in run()
7. Tag/Release-Erstellung entfernen
8. Tests

---

## Auth-Strategie

Token-Priorität in `GitHubInformation::new()`:
1. **Keyring-Token** (via `clikd login` gespeichert)
2. **Environment Variable** `GITHUB_TOKEN` (Fallback für CI)

Wenn beides fehlt: Fehler mit Hinweis auf `clikd login`.

---

## Branch-Naming

Format: `release/YYYYMMDD-HHMMSS` (Timestamp-basiert)
- Eindeutig ohne Konflikte
- Sortierbar
- Keine User-Konfiguration nötig
