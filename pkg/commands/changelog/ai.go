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

	// Wenn keine der AI-Flags aktiviert wurde, prüfe globale Einstellung
	if !aiEnableFlag && !aiEnhanceMessagesFlag && !aiGenerateSummariesFlag &&
		!aiCategorizeCommitsFlag && !aiSuggestVersionBumpFlag {

		// Checke Umgebungsvariable
		aiEnv := os.Getenv("CLIKD_CHANGELOG_AI_ENABLED")

		// Wenn weder Flag noch Umgebungsvariable gesetzt ist, verwende die globale Einstellung
		if aiEnv != "true" && aiEnv != "1" && aiEnv != "yes" && !globalAIEnabled {
			return nil
		}

		// Wenn nur die globale Einstellung oder Umgebungsvariable gesetzt ist, aktiviere alle Funktionen
		aiEnableFlag = true
		// In unserer neuen Struktur gibt es keine spezifischen Einstellungen für diese Features,
		// daher setzen wir sie standardmäßig auf true
		aiEnhanceMessagesFlag = true
		aiGenerateSummariesFlag = true
		aiCategorizeCommitsFlag = true
		aiSuggestVersionBumpFlag = true
	}

	// Wenn einzelne Funktionen aktiviert sind, aber nicht die Haupt-Flag, aktiviere diese
	if (aiEnhanceMessagesFlag || aiGenerateSummariesFlag ||
		aiCategorizeCommitsFlag || aiSuggestVersionBumpFlag) && !aiEnableFlag {
		aiEnableFlag = true
	}

	// Verwende das Standard-Modell aus der Konfiguration, wenn kein spezifisches angegeben wurde
	if aiModelFlag == "" {
		aiModelFlag = cfg.AI.DefaultModel
	}

	logger.Info("Initializing AI subsystem with model: %s", aiModelFlag)

	// Hole die Modell-Konfiguration
	var modelConfig config.ModelConfig
	if cfg.AI.Models != nil {
		if model, ok := cfg.AI.Models[aiModelFlag]; ok {
			modelConfig = model
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
		logger.Info("AI functionality is enabled")

		// Quelle der Aktivierung anzeigen
		if aiEnableFlag {
			logger.Info("AI enabled via command flag")
		} else {
			// Prüfen, ob über globale Einstellung aktiviert
			cfg, err := config.Get()
			if err == nil && cfg.AI.Enable {
				logger.Info("AI enabled via global configuration")
			} else {
				logger.Info("AI enabled via environment variable")
			}
		}

		if aiModelFlag != "" {
			logger.Info("Using AI model: %s", aiModelFlag)
		} else {
			logger.Info("Using default AI model from configuration")
		}

		// Zeige aktivierte Funktionen
		logger.Info("Enabled AI features:")
		if aiEnhanceMessagesFlag {
			logger.Info("- Commit message enhancement")
		}
		if aiGenerateSummariesFlag {
			logger.Info("- Summary generation")
		}
		if aiCategorizeCommitsFlag {
			logger.Info("- Commit categorization")
		}
		if aiSuggestVersionBumpFlag {
			logger.Info("- Version bump suggestion")
		}
	} else {
		logger.Info("AI functionality is disabled")
		logger.Info("To enable AI features, use the global --ai flag or set ai.enable=true in config")
	}
}
