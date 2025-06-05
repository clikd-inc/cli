# Changelog-Integration für clikd

Die Changelog-Integration ermöglicht es, automatisch CHANGELOG.md-Dateien aus Git-Commit-Nachrichten zu generieren. Sie basiert auf [git-chglog](https://github.com/git-chglog/git-chglog) und bietet eine vollständige Integration in die `clikd`-CLI.

## Funktionen

- Generieren von Changelogs basierend auf semantischen Versionierungskonventionen
- Unterstützung verschiedener Stile (GitHub, GitLab, Bitbucket)
- Anpassbare Templates für benutzerdefinierte Formatierung
- Jira-Integration für das Abrufen von Ticket-Informationen
- Pfad-Filterung zur Einschränkung des Changelogs auf bestimmte Dateien/Verzeichnisse
- Sortierung nach semantischer Version oder Datum
- Unterstützung für Next-Tag-Generierung
- Anpassbare Ausgabe mit Emoji und Farben
- KI-Integration für verbesserte Commit-Kategorisierung und Zusammenfassungen

## Installation

Die Changelog-Funktionalität ist bereits in `clikd` integriert. Keine zusätzliche Installation erforderlich.

## Konfiguration

Bei der ersten Verwendung wird automatisch eine Konfigurationsdatei erstellt. Alternativ kann die Konfiguration mit dem `--init`-Flag initialisiert werden:

```bash
clikd changelog --init
```

Die Konfiguration wird in `.chglog/config.yml` gespeichert und kann nach Bedarf angepasst werden.

## Verwendung

### Grundlegende Nutzung

Generieren eines Changelogs für alle Tags:

```bash
clikd changelog
```

Ausgabe in eine Datei:

```bash
clikd changelog -o CHANGELOG.md
```

### Filtern nach Tags

Für einen bestimmten Tag:

```bash
clikd changelog v1.0.0
```

Für einen Bereich von Tags:

```bash
clikd changelog v1.0.0..v2.0.0
```

Vom ersten Tag bis zu einem bestimmten Tag:

```bash
clikd changelog ..v1.0.0
```

Von einem bestimmten Tag bis zum neuesten:

```bash
clikd changelog v1.0.0..
```

### Optionen und Flags

| Flag | Beschreibung |
| --- | --- |
| `--config, -c` | Pfad zur Konfigurationsdatei (Standard: `.chglog/config.yml`) |
| `--template, -t` | Pfad zur Template-Datei (Standard: `.chglog/CHANGELOG.tpl.md`) |
| `--output, -o` | Ausgabepfad (Standard: stdout) |
| `--init` | Initialisiert die Konfiguration mit einem Assistenten |
| `--next-tag` | Behandelt nicht veröffentlichte Änderungen als den angegebenen Tag |
| `--path` | Filtern nach Pfaden (kommagetrennt) |
| `--sort` | Sortierung der Tags (`date` oder `semver`, Standard: `date`) |
| `--tag-filter-pattern` | Regulärer Ausdruck zum Filtern von Tags |
| `--no-case` | Case-insensitive Filterung von Commits |
| `--no-emoji` | Keine Emojis in der Ausgabe verwenden |
| `--no-color` | Keine Farben in der Ausgabe verwenden |
| `--silent` | Keine Warnungen oder Informationen ausgeben |
| `--verbose` | Ausführliche Logging-Informationen ausgeben |
| `--ai` | KI-Funktionen aktivieren |
| `--ai-model` | Spezifisches KI-Modell verwenden (z.B. `gpt-4`, `mistral-medium`) |
| `--ai-enhance-messages` | Commit-Nachrichten mit KI verbessern |
| `--ai-categorize-commits` | Commits mit KI kategorisieren |
| `--ai-generate-summaries` | Zusammenfassungen für Änderungen generieren |
| `--ai-suggest-version` | Versionsupgrade mit KI vorschlagen (Major, Minor, Patch) |

## Jira-Integration

Die Jira-Integration ermöglicht das Abrufen von Ticket-Informationen aus Jira, wenn Ticket-IDs in Commit-Nachrichten gefunden werden.

### Konfiguration

Die Jira-Integration kann über Umgebungsvariablen oder die Konfigurationsdatei konfiguriert werden:

```yaml
# In .chglog/config.yml
jira:
  url: "https://your-jira-instance.atlassian.net"
  username: "your-username"
  token: "your-api-token"
  type_maps:
    Story: "feat"
    Bug: "fix"
    Task: "chore"
```

Alternativ können Umgebungsvariablen verwendet werden:

```bash
export JIRA_URL="https://your-jira-instance.atlassian.net"
export JIRA_USERNAME="your-username"
export JIRA_TOKEN="your-api-token"
```

### Commit-Format für Jira-Integration

Commit-Nachrichten sollten die Jira-Ticket-ID im Header enthalten:

```
feat(core): Implementiere neue Funktion [JIRA-123]
```

oder

```
[JIRA-123] feat(core): Implementiere neue Funktion
```

## Template-Funktionen

Das Template-System unterstützt verschiedene Funktionen zur Anpassung der Ausgabe:

| Funktion | Beschreibung | Beispiel |
| --- | --- | --- |
| `contains` | Prüft, ob ein String einen anderen enthält | `{{ contains "string" "str" }}` |
| `datetime` | Formatiert Datum/Zeit | `{{ datetime "2006-01-02" .Tag.Date }}` |
| `hasPrefix` | Prüft, ob ein String mit einem Präfix beginnt | `{{ hasPrefix "string" "str" }}` |
| `hasSuffix` | Prüft, ob ein String mit einem Suffix endet | `{{ hasSuffix "string" "ing" }}` |
| `indent` | Einrückt alle Zeilen eines Strings | `{{ indent "line1\nline2" 4 }}` |
| `replace` | Ersetzt Teile eines Strings | `{{ replace .Body "\n" " " -1 }}` |
| `upperFirst` | Konvertiert den ersten Buchstaben in Großbuchstaben | `{{ upperFirst "string" }}` |

Zusätzlich werden alle [Sprig-Funktionen](http://masterminds.github.io/sprig/) unterstützt.

## Beispiele

### Standardmäßiger Changelog

```bash
clikd changelog -o CHANGELOG.md
```

### Changelog für einen bestimmten Bereich mit Pfadfilterung

```bash
clikd changelog v1.0.0..v2.0.0 --path="pkg/,cmd/" -o CHANGELOG.md
```

### Changelog mit Next-Tag

```bash
clikd changelog --next-tag v2.0.0 -o CHANGELOG.md
```

### Changelog mit Jira-Integration

```bash
export JIRA_URL="https://your-jira-instance.atlassian.net"
export JIRA_USERNAME="your-username"
export JIRA_TOKEN="your-api-token"
clikd changelog -o CHANGELOG.md
```

## KI-Integration

Die KI-Integration ermöglicht es, mit Hilfe von Sprachmodellen wie OpenAI's GPT-4 oder Mistral-AI, die Qualität und Konsistenz von Changelogs zu verbessern.

### Aktivierung

Die KI-Funktionalität kann über Kommandozeilen-Flags aktiviert werden:

```bash
clikd changelog --ai -o CHANGELOG.md
```

Oder durch Aktivierung spezifischer Funktionen:

```bash
clikd changelog --ai-enhance-messages --ai-categorize-commits -o CHANGELOG.md
```

### Konfiguration

Die KI-Konfiguration kann in einer YAML-Datei gespeichert werden (standardmäßig `.taskmasterconfig.yaml`):

```yaml
ai:
  enable_ai: true
  default_provider: "mistral"
  default_model: "mistral-medium"
  
  # Modellkonfigurationen
  models:
    mistral-medium:
      provider: "mistral"
      model_id: "mistral-medium"
      max_tokens: 1024
      temperature: 0.7
    
    gpt-4:
      provider: "openai"
      model_id: "gpt-4"
      max_tokens: 1024
      temperature: 0.7
```

Die API-Schlüssel können über Umgebungsvariablen gesetzt werden:

```bash
export MISTRAL_API_KEY="your-mistral-api-key"
export OPENAI_API_KEY="your-openai-api-key"
```

### Verfügbare KI-Funktionen

Die KI-Integration bietet folgende Funktionalitäten:

1. **Commit-Nachrichtenverbesserung**: Macht Commit-Nachrichten klarer und konsistenter
2. **Commit-Kategorisierung**: Ordnet Commits automatisch in die richtigen Kategorien ein
3. **Zusammenfassungsgenerierung**: Erstellt übersichtliche Zusammenfassungen für Versionen
4. **Versionsvorschläge**: Schlägt basierend auf den Änderungen ein Versionsupgrade vor (Major, Minor, Patch)

### Unterstützte KI-Provider

- **Mistral AI**: Standard-Provider mit guter Balance aus Leistung und Kosten
- **OpenAI**: Hohe Qualität für anspruchsvollere Aufgaben
- **Azure OpenAI**: Für Unternehmensumgebungen mit Azure-Integration
- **Lokale Modelle**: Unterstützung für Ollama und andere lokale LLM-Server

### Beispiel

```bash
# Mit Mistral AI für Commit-Kategorisierung und Zusammenfassungen
clikd changelog --ai --ai-model mistral-medium --ai-categorize-commits --ai-generate-summaries -o CHANGELOG.md

# Mit OpenAI für alle KI-Funktionen
clikd changelog --ai --ai-model gpt-4 -o CHANGELOG.md
``` 
