package tui

import (
	"github.com/spf13/cobra"
)

// HelpCmd represents the help command
var HelpCmd = &cobra.Command{
	Use:   "help",
	Short: L("Show help for TUI command"),
	Run:   runHelp,
}

func runHelp(cmd *cobra.Command, args []string) {
	// Show help for the parent tui command
	cmd.Parent().Help()
}
