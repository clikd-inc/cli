# Konfigurationsstrategie für clikd CLI

Dieses Dokument beschreibt die einheitliche Konfigurationsstrategie für die clikd CLI, die sowohl globale Benutzereinstellungen als auch projektspezifische Konfigurationen unterstützt.

## Konfigurationsebenen

Die Konfiguration erfolgt auf drei Ebenen mit absteigender Priorität:

1. **Temporäre/Kommandozeilen-Konfiguration**: Über Flags und Argumente, überschreibt alle anderen Einstellungen.
2. **Projekt-/Repository-Konfiguration**: Spezifisch für ein Projekt, überschreibt die globale Konfiguration.
3. **Globale Benutzer-Konfiguration**: Für den Benutzer überall verfügbar, unabhängig vom Projekt.
4. **Standardwerte**: Fest in der Anwendung codiert, werden verwendet, wenn keine anderen Konfigurationen vorhanden sind.

## Speicherorte und Dateiformate

### Globale Konfiguration
- **Dateipfad**: `~/.config/clikd/config.toml`
- **Format**: TOML (klar strukturiert und leicht zu bearbeiten)
- **Inhalt**: Allgemeine Einstellungen, die für alle Projekte gelten

### Projekt-Konfiguration
- **Dateipfad**: `clikd/config.toml` im Projektverzeichnis
- **Format**: TOML (identisch zur globalen Konfiguration)
- **Inhalt**: Projektspezifische Einstellungen, die die globale Konfiguration überschreiben

### Projektstruktur
- **clikd/**: Hauptverzeichnis für projektspezifische Konfigurationen (nicht versteckt)
  - **config.toml**: Hauptkonfigurationsdatei
  - **templates/**: Verzeichnis für Templates (z.B. changelog.md)
  - **cache/**: Verzeichnis für generierte Dateien
  - **plugins/**: Verzeichnis für Erweiterungen

### Sensible Daten
- **Umgebungsvariablen**: Primärer Weg für API-Keys und andere sensible Daten
- **Lokale .env-Datei**: `.env` im Projektverzeichnis (Standard-Konvention) erstellt keine sondern nutzt die im root die eh da ist.
- **Globale nutzt environments die in zsh oder bash gespeichert sind.

## Konfigurationsstruktur

Alle Konfigurationen werden in einer einheitlichen Hierarchie organisiert:

```toml
# Globale Einstellungen
version = "1.0.0"

# AI-Konfiguration
[ai]
enable = true
default_model = "mistral-medium"
default_provider = "mistral"

[ai.models.mistral-medium]
provider = "mistral"
max_tokens = 1024
temperature = 0.7
top_p = 0.9

[ai.models.gpt-4]
provider = "openai"
max_tokens = 1024
temperature = 0.7

# Changelog-Konfiguration
[changelog]
style = "github"
template = "templates/changelog.md"
jira_integration = false

# Weitere Funktionen folgen dem gleichen Muster
[function_name]
setting1 = "value1"
setting2 = "value2"
```

## Umgebungsvariablen

Umgebungsvariablen haben Vorrang vor Datei-Konfigurationen und folgen einer einheitlichen Namenskonvention:

- **Präfix**: `CLIKD_` für alle clikd-spezifischen Variablen
- **API-Keys**: Behalten ihren üblichen Namen, z.B. `MISTRAL_API_KEY`, `OPENAI_API_KEY`
- **Konfigurationsvariablen**: Entsprechen der Hierarchie, z.B. `CLIKD_AI_ENABLE`, `CLIKD_CHANGELOG_STYLE`

## Konfigurationsmanagement

Die CLI bietet einheitliche Befehle für die Konfigurationsverwaltung:

```bash
# Globale Konfiguration anzeigen
clikd config get [key]

# Globale Konfiguration setzen
clikd config set [key=value]

# Projekt-Konfiguration anzeigen
clikd config get --local [key]

# Projekt-Konfiguration setzen
clikd config set --local [key=value]

# Konfiguration beschreiben (Dokumentation)
clikd config describe [key]

# Konfiguration initialisieren
clikd config init [--local]
```

### Beispiele

```bash
# AI-Modell global setzen
clikd config set ai.default_model=gpt-4

# Changelog-Stil projektspezifisch setzen
clikd config set --local changelog.style=gitlab

# Alle AI-Einstellungen anzeigen
clikd config get ai

# Dokumentation für ein bestimmtes Feature anzeigen
clikd config describe changelog
```

## Prioritätsreihenfolge

Die Konfigurationswerte werden in folgender Reihenfolge ausgewertet (höchste bis niedrigste Priorität):

1. Kommandozeilen-Flags (z.B. `clikd --ai-model=gpt-4 changelog`)
2. Umgebungsvariablen mit `CLIKD_` Präfix (z.B. `CLIKD_AI_MODEL=gpt-4`)
3. Projektspezifische Umgebungsvariablen aus `.env` im Projektverzeichnis
4. Projekt-Konfiguration aus `clikd/config.toml`
5. Globale Umgebungsvariablen aus `~/.config/clikd/.env`
6. Globale Konfiguration aus `~/.config/clikd/config.toml`
7. Standardwerte der Anwendung

## API-Key-Verwaltung

API-Keys und andere sensible Daten werden sicher gehandhabt:

1. **Priorität für Umgebungsvariablen**: 
   - Direkt gesetzt: `MISTRAL_API_KEY=xyz123`
   - Aus .env-Dateien geladen (Projekt > Global)

2. **Referenzierung in Konfiguration**: Anstatt API-Keys direkt zu speichern, können Umgebungsvariablen referenziert werden:
   ```toml
   [ai.models.mistral-medium]
   api_key = "${MISTRAL_API_KEY}"  # Referenziert eine Umgebungsvariable
   ```

3. **Warnungen**: Die CLI warnt Benutzer, wenn sie API-Keys direkt in Konfigurationsdateien speichern.

## Initialisierung und Migration

- **Initialisierung**: `clikd init` erstellt die grundlegende Projektstruktur mit Standardwerten
- **Projektinitialisierung**: `clikd config init --local` erstellt nur die projektspezifische Konfiguration
- **Migration**: Automatische Migration von .chglog/ zu clikd/ wird unterstützt

### Migration von .chglog

Für bestehende Projekte, die .chglog/ verwenden:

1. Die Datei `.chglog/CHANGELOG.tpl.md` wird nach `clikd/templates/changelog.md` migriert
2. Die Einstellungen aus `.chglog/config.yml` werden in den `[changelog]`-Abschnitt der `clikd/config.toml` konvertiert
3. Da es noch keine bestehenden Benutzer gibt, ist keine Abwärtskompatibilitätsschicht erforderlich

## Fehlerbehebung und Validierung

- **Validierung**: Konfigurationswerte werden beim Einlesen validiert
- **Fehlerbehandlung**: Klare Fehlermeldungen mit Vorschlägen zur Behebung
- **Diagnose**: `clikd config diagnose` zeigt die aktuell wirksame Konfiguration und Herkunft der Werte

## Beispiel: Vollständige Konfigurationsdatei

```toml
# clikd Configuration
version = "1.0.0"

# Allgemeine Einstellungen
[general]
log_level = "info"
color = true

# AI-Konfiguration
[ai]
enable = true
default_model = "mistral-medium"
default_provider = "mistral"
verbose = false

[ai.models.mistral-medium]
provider = "mistral"
model_id = "mistral-medium"
max_tokens = 1024
temperature = 0.7
top_p = 0.9
context_window = 8192

[ai.models.gpt-4]
provider = "openai"
model_id = "gpt-4"
max_tokens = 1024
temperature = 0.7
top_p = 0.9
context_window = 8192

# Changelog-Konfiguration
[changelog]
style = "github"
template = "templates/changelog.md"
jira_integration = false
sort = "semver"
tag_filter_pattern = "v*"
path = ""
no_case = false

[changelog.jira]
base_url = ""
username = ""
project_key = ""
issue_pattern = "[A-Z]+-[0-9]+"
``` 
