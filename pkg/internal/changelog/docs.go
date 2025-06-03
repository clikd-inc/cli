// Package changelog implements main logic for the CHANGELOG generate.
package changelog

/*
# Template-Funktionen Dokumentation

Diese Dokumentation beschreibt alle verfügbaren Template-Funktionen, die in den CHANGELOG-Templates verwendet werden können.

## Benutzerdefinierte Funktionen

### `contains`

Prüft, ob ein String einen anderen enthält.

```
{{ contains "substring" "string" }}
```

Beispiel:
```
{{ if contains "fix" .Type }}Dies ist ein Bugfix{{ end }}
```

### `datetime`

Formatiert ein Datum nach einem bestimmten Layout.

```
{{ datetime "layout" .Date }}
```

Das Layout folgt dem Go-Zeitformat, wobei das Referenzdatum verwendet wird:
- `2006-01-02` - YYYY-MM-DD Format
- `Jan 2, 2006` - Monat Tag, Jahr Format
- `2006-01-02 15:04:05` - Datum mit Uhrzeit

Beispiel:
```
{{ datetime "2006-01-02" .Tag.Date }}
```

### `hasPrefix`

Prüft, ob ein String mit einem bestimmten Präfix beginnt.

```
{{ hasPrefix "prefix" "string" }}
```

Beispiel:
```
{{ if hasPrefix "v" .Tag.Name }}Dies ist ein Release-Tag{{ end }}
```

### `hasSuffix`

Prüft, ob ein String mit einem bestimmten Suffix endet.

```
{{ hasSuffix "suffix" "string" }}
```

Beispiel:
```
{{ if hasSuffix "-beta" .Tag.Name }}Dies ist ein Beta-Release{{ end }}
```

### `indent`

Rückt alle Zeilen eines Strings um eine bestimmte Anzahl von Leerzeichen ein.

```
{{ indent "string" n }}
```

Beispiel:
```
{{ indent .Body 4 }}
```

### `replace`

Ersetzt alle Vorkommen eines Substrings in einem String durch einen anderen.

```
{{ replace "string" "old" "new" count }}
```

Dabei ist `count` die maximale Anzahl von Ersetzungen. -1 bedeutet alle ersetzen.

Beispiel:
```
{{ replace .Body "\n" " " -1 }}
```

### `upperFirst`

Konvertiert den ersten Buchstaben eines Strings in einen Großbuchstaben.

```
{{ upperFirst "string" }}
```

Beispiel:
```
{{ upperFirst .Type }}
```

## Sprig-Funktionen

Zusätzlich zu den oben genannten benutzerdefinierten Funktionen unterstützt das Template-System alle Funktionen aus der Sprig-Bibliothek.
Einige häufig verwendete Sprig-Funktionen sind:

### Strings

- `trim` - Entfernt Leerzeichen am Anfang und Ende
- `trimPrefix` - Entfernt ein Präfix
- `trimSuffix` - Entfernt ein Suffix
- `lower` - Konvertiert zu Kleinbuchstaben
- `upper` - Konvertiert zu Großbuchstaben
- `title` - Konvertiert zu Title Case
- `repeat` - Wiederholt einen String n-mal
- `substr` - Extrahiert einen Substring
- `trunc` - Kürzt einen String auf eine bestimmte Länge

### Listen und Datenstrukturen

- `first` - Erstes Element einer Liste
- `last` - Letztes Element einer Liste
- `rest` - Alle Elemente außer dem ersten
- `initial` - Alle Elemente außer dem letzten
- `append` - Fügt ein Element hinzu
- `prepend` - Fügt ein Element am Anfang hinzu
- `reverse` - Kehrt die Reihenfolge um
- `uniq` - Entfernt Duplikate
- `without` - Entfernt bestimmte Elemente

### Bedingungen

- `eq` - Gleich
- `ne` - Ungleich
- `lt` - Kleiner als
- `le` - Kleiner oder gleich
- `gt` - Größer als
- `ge` - Größer oder gleich
- `default` - Standardwert, wenn leer

### Mathematik

- `add` - Addition
- `sub` - Subtraktion
- `mul` - Multiplikation
- `div` - Division
- `mod` - Modulo

Eine vollständige Liste aller Sprig-Funktionen finden Sie unter:
http://masterminds.github.io/sprig/

## Template-Kontext

In den Templates stehen die folgenden Objekte zur Verfügung:

### Info

- `.Info.Title` - Titel des Changelogs
- `.Info.RepositoryURL` - URL des Repositories

### Unreleased

- `.Unreleased.CommitGroups` - Gruppen von nicht freigegebenen Commits
- `.Unreleased.Commits` - Nicht freigegebene Commits
- `.Unreleased.MergeCommits` - Nicht freigegebene Merge-Commits
- `.Unreleased.RevertCommits` - Nicht freigegebene Revert-Commits
- `.Unreleased.NoteGroups` - Gruppen von nicht freigegebenen Notizen

### Versions

Eine Liste aller Versionen, wobei jede Version die folgenden Felder hat:

- `.Versions[i].Tag` - Tag-Information (Name, Date, etc.)
- `.Versions[i].CommitGroups` - Commit-Gruppen für diese Version
- `.Versions[i].Commits` - Commits für diese Version
- `.Versions[i].MergeCommits` - Merge-Commits für diese Version
- `.Versions[i].RevertCommits` - Revert-Commits für diese Version
- `.Versions[i].NoteGroups` - Notizgruppen für diese Version

### Commit

Ein Commit hat die folgenden Felder:

- `.Hash` - Hash des Commits (Long, Short)
- `.Author` - Autor (Name, Email, Date)
- `.Committer` - Committer (Name, Email, Date)
- `.Merge` - Merge-Informationen (falls ein Merge-Commit)
- `.Revert` - Revert-Informationen (falls ein Revert-Commit)
- `.Refs` - Referenzen (Issues, PRs, etc.)
- `.Notes` - Notizen (wie BREAKING CHANGE)
- `.Mentions` - Erwähnte Benutzer
- `.JiraIssue` - Jira-Issue (falls vorhanden)
- `.Header` - Header des Commits
- `.Type` - Typ des Commits (feat, fix, etc.)
- `.Scope` - Scope des Commits
- `.Subject` - Betreff des Commits
- `.JiraIssueID` - ID des Jira-Issues (falls vorhanden)
- `.Body` - Body des Commits
- `.TrimmedBody` - Gekürzter Body ohne Notizen

### JiraIssue

Ein JiraIssue hat die folgenden Felder:

- `.Type` - Typ des Issues (Story, Bug, etc.)
- `.Summary` - Zusammenfassung des Issues
- `.Description` - Beschreibung des Issues
- `.Labels` - Labels des Issues

## Beispiele

### Einfaches Changelog

```
# Changelog

{{ range .Versions }}
## {{ .Tag.Name }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ range .CommitGroups }}
### {{ .Title }}

{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end }}

{{ range .NoteGroups }}
### {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end }}
{{ end }}
```

### Changelog mit Links

```
# Changelog

{{ range .Versions }}
## [{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/releases/tag/{{ .Tag.Name }})

{{ range .CommitGroups }}
### {{ .Title }}

{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }} ([{{ .Hash.Short }}]({{ $.Info.RepositoryURL }}/commit/{{ .Hash.Long }}))
{{ end }}
{{ end }}
{{ end }}
```

### Changelog mit Jira-Integration

```
# Changelog

{{ range .Versions }}
## {{ .Tag.Name }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ range .CommitGroups }}
### {{ .Title }}

{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
  {{- if .JiraIssue }}
  [{{ .JiraIssue.Key }}] {{ .JiraIssue.Summary }}
  {{- end }}
{{ end }}
{{ end }}
{{ end }}
```
*/
