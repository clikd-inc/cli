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
