# KI-Funktionalität der clikd CLI

Die clikd CLI bietet KI-unterstützte Funktionen, die Ihre Arbeit erleichtern. Diese Anleitung erklärt, wie Sie die KI-Funktionalität einrichten und verwenden können.

## Schnellstart

1. Legen Sie Ihre API-Keys in einer `.env`-Datei im Projektverzeichnis fest:
   ```
   MISTRAL_API_KEY=your_mistral_api_key
   OPENAI_API_KEY=your_openai_api_key
   ```

2. Aktivieren Sie die KI-Funktionalität mit der Umgebungsvariable oder dem Flag:
   ```
   # Per Umgebungsvariable
   export CLIKD_ENABLE_AI=true
   
   # Oder per Flag bei jedem Befehl
   clikd --ai [befehl]
   ```

3. Wählen Sie Ihr bevorzugtes Modell (optional):
   ```
   clikd config set ai.default_model mistral-medium
   ```

## Verfügbare Modelle

Die clikd CLI unterstützt standardmäßig folgende KI-Modelle:

- `mistral-medium` (Standard): Ausgeglichenes Modell für die meisten Aufgaben
- `mistral-small`: Schnelleres, kleineres Modell
- `gpt-3.5-turbo`: OpenAI's GPT-3.5 Modell
- `gpt-4`: OpenAI's GPT-4 Modell (erfordert entsprechenden API-Zugang)

## API-Keys einrichten

Für die Nutzung der KI-Funktionalität benötigen Sie API-Keys von den entsprechenden Anbietern. Es gibt mehrere Möglichkeiten, diese zu konfigurieren:

### Option 1: .env-Datei (empfohlen)

Erstellen Sie eine `.env`-Datei im Hauptverzeichnis Ihres Projekts:

```
MISTRAL_API_KEY=your_mistral_api_key
OPENAI_API_KEY=your_openai_api_key
```

### Option 2: Umgebungsvariablen

Setzen Sie die API-Keys direkt als Umgebungsvariablen:

```bash
export MISTRAL_API_KEY=your_mistral_api_key
export OPENAI_API_KEY=your_openai_api_key
```

### Option 3: Konfigurationsdatei

Sie können die API-Keys auch in der Konfigurationsdatei speichern:

```bash
clikd config set ai.models.mistral-medium.api_key your_mistral_api_key
clikd config set ai.models.gpt-4.api_key your_openai_api_key
```

**Hinweis**: Diese Methode speichert Ihre API-Keys im Klartext in der Konfigurationsdatei. Verwenden Sie diese Option nur, wenn die Sicherheit kein Problem darstellt.

## KI-Einstellungen anpassen

### Bevorzugtes Modell ändern

```bash
clikd config set ai.default_model gpt-4
```

### KI-Funktionalität aktivieren/deaktivieren

```bash
# Aktivieren
clikd config set ai.enable_ai true

# Deaktivieren
clikd config set ai.enable_ai false
```

### Temporär ein anderes Modell verwenden

```bash
clikd --model=gpt-4 [befehl]
```

## Konfigurationsübersicht anzeigen

Um Ihre aktuelle KI-Konfiguration anzuzeigen:

```bash
clikd config get ai
```

## Fehlerbehebung

### API-Key-Problem

Wenn Sie die Fehlermeldung "API key not configured" erhalten:

1. Überprüfen Sie, ob der entsprechende API-Key in Ihrer `.env`-Datei oder als Umgebungsvariable korrekt gesetzt ist
2. Verwenden Sie `clikd config get ai` um zu prüfen, ob die Konfiguration korrekt geladen wurde
3. Versuchen Sie, ein anderes Modell zu verwenden, für das Sie einen API-Key haben

### Modell nicht verfügbar

Wenn ein Modell nicht verfügbar ist:

1. Stellen Sie sicher, dass Sie den richtigen API-Key konfiguriert haben
2. Prüfen Sie, ob Ihr Konto Zugang zu diesem Modell hat (besonders bei GPT-4)
3. Versuchen Sie, ein alternatives Modell zu verwenden

## Unterstützte Befehle mit KI-Funktionalität

Die folgenden Befehle unterstützen KI-Funktionalität:

- `clikd changelog`: Generierung und Verbesserung von Changelogs
- Weitere Befehle folgen in zukünftigen Versionen 
