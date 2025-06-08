# Git Service

Der Git-Service stellt eine zentrale Schnittstelle für Git-Operationen bereit, die von verschiedenen Befehlen in der CLI verwendet werden können.

## Verwendung

```go
// Erstelle einen neuen Git-Service für ein Repository
service := git.NewService("/pfad/zum/repo")

// Hole alle Tags im Repository
tags, err := service.GetTags()
if err != nil {
    // Fehlerbehandlung
}

// Wähle Tags basierend auf einer Abfrage aus
selectedTags, firstTag, err := service.SelectTags(tags, "v1.0.0..v2.0.0")
if err != nil {
    // Fehlerbehandlung
}

// Hole den aktuellen Branch-Namen
branch, err := service.GetCurrentBranch()
if err != nil {
    // Fehlerbehandlung
}
```

## Migration aus dem Changelog-Modul

Dieser Service wurde aus dem Changelog-Modul extrahiert und stellt allgemeine Git-Funktionalität bereit, die von verschiedenen Befehlen verwendet werden kann. 

Aktuell unterstützt der Service:
- Tag-Management (Lesen und Filtern)
- Repository-Informationen (Branch)

In zukünftigen Versionen werden weitere Funktionalitäten hinzugefügt, wie:
- Commit-Management
- Branches erstellen und verwalten
- Stash-Operationen
- und mehr

## TODO

- [ ] Commit-Parser und -Extractor anpassen
- [ ] Jira-Integration bereinigen
- [ ] Tests hinzufügen
- [ ] Vollständige Dokumentation 
