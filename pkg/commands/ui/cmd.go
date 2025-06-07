package ui

import (
	"clikd/pkg/ui/demo"

	"github.com/spf13/cobra"
)

// NewUICmd erstellt einen neuen Befehl zum Anzeigen von UI-Demos
func NewUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Zeigt UI-Komponenten-Demos an",
		Long: `Zeigt interaktive Demos der verfügbaren UI-Komponenten an.
Diese Demos veranschaulichen die verschiedenen UI-Elemente, die in der CLI verwendet werden.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			demo.RunDemo()
			return nil
		},
	}

	return cmd
}
