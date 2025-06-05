package changelog

import (
	"os"

	"clikd/pkg/config"
	"clikd/pkg/internal/changelog"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
)

// AI-bezogene Flags
var (
	aiEnableFlag             bool
	aiModelFlag              string
	aiEnhanceMessagesFlag    bool
	aiGenerateSummariesFlag  bool
	aiCategorizeCommitsFlag  bool
	aiSuggestVersionBumpFlag bool
)

// AddAIFlags fügt die KI-bezogenen Flags zu einem Kommando hinzu
func AddAIFlags(cmd *cobra.Command) {
	// Diese Flag bleibt für kommandospezifische Überschreibungen
	cmd.Flags().BoolVar(&aiEnableFlag, "ai", false, "Override global AI setting for this command only")

	// Modellauswahl
	cmd.Flags().StringVar(&aiModelFlag, "ai-model", "", "Specify AI model to use (overrides config default)")

	// Detaillierte Steuerung der AI-Funktionen
	cmd.Flags().BoolVar(&aiEnhanceMessagesFlag, "ai-enhance-messages", true, "Use AI to enhance commit messages (requires AI enabled)")
	cmd.Flags().BoolVar(&aiGenerateSummariesFlag, "ai-generate-summaries", true, "Use AI to generate summaries for changes (requires AI enabled)")
	cmd.Flags().BoolVar(&aiCategorizeCommitsFlag, "ai-categorize-commits", true, "Use AI to categorize commit messages (requires AI enabled)")
	cmd.Flags().BoolVar(&aiSuggestVersionBumpFlag, "ai-suggest-version", true, "Use AI to suggest version bump type (requires AI enabled)")

	// Flags als Hidden markieren, um die CLI-Oberfläche sauber zu halten
	// Sie können immer noch verwendet werden, aber erscheinen nicht in der Basis-Hilfe
	cmd.Flags().MarkHidden("ai-model")
	cmd.Flags().MarkHidden("ai-enhance-messages")
	cmd.Flags().MarkHidden("ai-generate-summaries")
	cmd.Flags().MarkHidden("ai-categorize-commits")
	cmd.Flags().MarkHidden("ai-suggest-version")
}

// InitializeAI initialisiert die KI-Funktionalität basierend auf den Flags
func InitializeAI() error {
	logger := utils.NewLogger("info", true)

	// Globale Konfiguration abrufen
	cfg, err := config.EnsureInitialized()
	if err != nil {
		return err
	}

	// Prüfe auf globale KI-Aktivierung aus der Konfiguration
	globalAIEnabled := cfg.AI.Enable

	// Standardwerte aus der Konfiguration verwenden
	aiEnableFlag = globalAIEnabled

	// Check if the flag was explicitly set (override the default)
	flagExplicitlySet := false
	// Prüfe, ob das Flag explizit gesetzt wurde (funktioniert nur während der Ausführung des Befehls)
	// Diese Zeile ist hier hauptsächlich als Platzhalter für die Logik
	// In der Praxis wird das Flag über den cobra.Command-Parameter übergeben
	// oder über eine globale Variable geprüft

	// Wenn AI-Flag explizit gesetzt wurde, verwende diesen Wert anstelle der Konfiguration
	if flagExplicitlySet {
		// Flags haben Vorrang vor der Konfiguration
		logger.Debug("AI flag wurde explizit gesetzt, überschreibe Konfiguration")
	} else {
		// Wenn keine Flag gesetzt wurde, verwende die Konfiguration
		logger.Debug("Verwende KI-Einstellungen aus der Konfiguration: %v", globalAIEnabled)
	}

	// Prüfe auf AI-Flags für spezifische Funktionen
	if !aiEnhanceMessagesFlag && !aiGenerateSummariesFlag &&
		!aiCategorizeCommitsFlag && !aiSuggestVersionBumpFlag {
		// Wenn keine spezifischen Flags gesetzt wurden, setze alle entsprechend der globalen Einstellung
		aiEnhanceMessagesFlag = aiEnableFlag
		aiGenerateSummariesFlag = aiEnableFlag
		aiCategorizeCommitsFlag = aiEnableFlag
		aiSuggestVersionBumpFlag = aiEnableFlag
	} else {
		// Wenn mindestens ein Feature-Flag gesetzt wurde, stelle sicher, dass die Haupt-Flag aktiviert ist
		if !aiEnableFlag && (aiEnhanceMessagesFlag || aiGenerateSummariesFlag ||
			aiCategorizeCommitsFlag || aiSuggestVersionBumpFlag) {
			aiEnableFlag = true
		}
	}

	// Wenn KI deaktiviert ist, brechen wir hier ab
	if !aiEnableFlag {
		return nil
	}

	// Setze eine Umgebungsvariable, um in GetAPIKey zu zeigen, dass KI explizit aktiviert wurde
	// Dies führt zu besseren Fehlermeldungen, wenn API-Keys fehlen
	if aiEnableFlag {
		os.Setenv("CLIKD_AI_EXPLICITLY_ENABLED", "true")
	}

	// Verwende das Standard-Modell aus der Konfiguration, wenn kein spezifisches angegeben wurde
	if aiModelFlag == "" {
		aiModelFlag = cfg.AI.DefaultModel
	}

	logger.Info("Initialisiere KI-Subsystem mit Modell: %s", aiModelFlag)

	// Hole die Modell-Konfiguration
	var modelConfig config.ModelConfig
	if cfg.AI.Models != nil {
		if model, ok := cfg.AI.Models[aiModelFlag]; ok {
			modelConfig = model
		}
	}

	// Prüfe, ob die benötigten API-Keys vorhanden sind
	localConfigExists := utils.IsLocalConfigPresent()
	var providerInfo utils.ProviderKeyInfo

	switch modelConfig.Provider {
	case "mistral":
		providerInfo = utils.MistralProvider
	case "openai":
		providerInfo = utils.OpenAIProvider
	case "azure-openai":
		providerInfo = utils.ProviderKeyInfo{
			Name:            "Azure OpenAI",
			ConfigKey:       "ai.models.azure-openai.api_key",
			EnvVarName:      "CLIKD_AZURE_OPENAI_API_KEY",
			EnvVarNameShort: "AZURE_OPENAI_API_KEY",
			Required:        true,
		}
	}

	// API-Key benötigt und prüfen
	if modelConfig.Provider != "local" {
		providerInfo.Required = true
		_, err := utils.GetAPIKey(providerInfo, localConfigExists)
		if err != nil {
			// Ausführliche Warnung, wenn KI explizit aktiviert wurde, aber kein API-Key vorhanden ist
			logger.Error("KI ist aktiviert, aber der API-Key konnte nicht geladen werden.")
			logger.Error("Der Changelog wird ohne KI-Funktionen generiert.")
			logger.Error("%v", err)
			// Deaktiviere KI-Funktionalität
			aiEnableFlag = false
			aiEnhanceMessagesFlag = false
			aiGenerateSummariesFlag = false
			aiCategorizeCommitsFlag = false
			aiSuggestVersionBumpFlag = false

			// Umgebungsvariable zurücksetzen
			os.Unsetenv("CLIKD_AI_EXPLICITLY_ENABLED")

			return nil
		}
	}

	// KI initialisieren
	return changelog.InitAI(modelConfig, changelog.AIOptions{
		EnableAI:              aiEnableFlag,
		ModelName:             aiModelFlag,
		EnhanceCommitMessages: aiEnhanceMessagesFlag,
		GenerateSummaries:     aiGenerateSummariesFlag,
		CategorizeCommits:     aiCategorizeCommitsFlag,
		SuggestVersionBump:    aiSuggestVersionBumpFlag,
	})
}

// ShowAIStatus zeigt den Status der KI-Funktionalität an
func ShowAIStatus() {
	logger := utils.NewLogger("info", true)

	if changelog.IsAIEnabled() {
		logger.Info("KI-Funktionalität ist aktiviert")

		// Quelle der Aktivierung anzeigen
		if aiEnableFlag {
			// Prüfen, ob über Flag, Umgebungsvariable oder Konfiguration aktiviert
			cfg, err := config.Get()
			if err == nil && !cfg.AI.Enable && aiEnableFlag {
				logger.Info("KI über Befehlsflag aktiviert (überschreibt Konfiguration)")
			} else if err == nil && cfg.AI.Enable {
				logger.Info("KI über globale Konfiguration aktiviert")
			} else {
				logger.Info("KI über Umgebungsvariable aktiviert")
			}
		}

		// Modell-Informationen anzeigen
		cfg, err := config.Get()
		if err == nil {
			modelName := aiModelFlag
			if modelName == "" {
				modelName = cfg.AI.DefaultModel
			}

			logger.Info("Ausgewähltes KI-Modell: %s", modelName)

			// Zeige Modell-Details, wenn verfügbar
			if model, ok := cfg.AI.Models[modelName]; ok {
				logger.Info("  Provider: %s", model.Provider)
				logger.Info("  Modell-ID: %s", model.ModelID)
				logger.Info("  Max Tokens: %d", model.MaxTokens)
				logger.Info("  Context Window: %d Tokens", model.ContextWindow)
			}
		}

		// Zeige aktivierte Funktionen
		logger.Info("Aktivierte KI-Funktionen:")
		if aiEnhanceMessagesFlag {
			logger.Info("- Commit-Nachrichten verbessern")
		}
		if aiGenerateSummariesFlag {
			logger.Info("- Zusammenfassungen generieren")
		}
		if aiCategorizeCommitsFlag {
			logger.Info("- Commits kategorisieren")
		}
		if aiSuggestVersionBumpFlag {
			logger.Info("- Versionsupdate vorschlagen")
		}

		// Hinweis zur API-Key-Quelle
		if utils.IsLocalConfigPresent() {
			logger.Info("API-Schlüssel wird aus der lokalen .env-Datei geladen")
		} else {
			logger.Info("API-Schlüssel wird aus der globalen Konfiguration geladen")
		}
	} else {
		logger.Info("KI-Funktionalität ist deaktiviert")
		logger.Info("Um KI-Funktionen zu aktivieren, setzen Sie ai.enable=true in der Konfiguration oder verwenden Sie das --ai Flag")
	}
}
