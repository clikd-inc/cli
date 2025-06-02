package hello

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	name     string
	language string
)

// NewHelloCmd creates a new hello command with subcommands
func NewHelloCmd() *cobra.Command {
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Say hello to someone",
		Long:  "A simple command to demonstrate subcommands and flags",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				cmd.Help()
				return
			}

			message := fmt.Sprintf("Hello, %s!", name)
			fmt.Println(message)
		},
	}

	// Add local flags for the hello command
	helloCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the person to greet")
	helloCmd.Flags().StringVarP(&language, "language", "l", "english", "Language for the greeting (english, spanish, german)")

	// Add subcommands
	helloCmd.AddCommand(newHelloWorldCmd())

	return helloCmd
}

// newHelloWorldCmd creates a hello world subcommand
func newHelloWorldCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "world",
		Short: "Say hello to the world",
		Long:  "A simple command that says hello to the world",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, World!")
		},
	}
}
