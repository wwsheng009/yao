package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui"
)

var Debug bool
var Verbose bool

// Cmd represents the tui command
var Cmd = &cobra.Command{
	Use:   "tui [TUI_ID]",
	Short: L("Run a terminal user interface"),
	Long:  L("Run a terminal user interface defined in .tui.yao files"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Boot()

		tuiID := args[0]

		if Verbose {
			log.Info("Starting TUI: %s", tuiID)
		}

		// Load application engine
		loadWarnings, err := engine.Load(config.Conf, engine.LoadOption{
			Action: "tui",
		}, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to load engine: %v\n", color.RedString("Error:"), err)
			os.Exit(1)
		}

		if Debug && len(loadWarnings) > 0 {
			log.Warn("Load warnings: %v", loadWarnings)
		}

		// Load TUI configurations
		err = tui.Setup(config.Conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to load TUI configurations: %v\n", color.RedString("Error:"), err)
			os.Exit(1)
		}

		if Verbose {
			log.Info("Loaded %d TUI configuration(s)", Count())
		}

		// Get TUI configuration
		cfg := tui.Get(tuiID)
		if cfg == nil {
			fmt.Fprintf(os.Stderr, "%s TUI not found: %s\n", color.RedString("Error:"), tuiID)
			fmt.Fprintf(os.Stderr, "\n%s Available TUIs:\n", color.YellowString("Hint:"))

			availableTUIs := List()
			if len(availableTUIs) == 0 {
				fmt.Fprintf(os.Stderr, "  No TUI configurations found in tuis/ directory\n")
			} else {
				for _, id := range availableTUIs {
					fmt.Fprintf(os.Stderr, "  - %s\n", color.CyanString(id))
				}
			}
			os.Exit(1)
		}

		if Verbose {
			if cfg.Name != "" {
				log.Info("Running TUI: %s (%s)", tuiID, cfg.Name)
			}
		}

		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start goroutine to monitor signals and cancel context
		go func() {
			select {
			case <-sigChan:
				if Debug {
					log.Info("Received interrupt signal, gracefully shutting down...")
				}
				cancel()
			case <-ctx.Done():
				// Context cancelled
			}
		}()

		// Create Bubble Tea program with context
		model := tui.NewModel(nil, nil)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),       // Use alternate screen buffer
			tea.WithMouseCellMotion(), // Enable mouse support
			tea.WithContext(ctx),      // Add context for graceful shutdown
		)

		// Run program
		if _, err := program.Run(); err != nil {
			if Debug {
				log.Error("TUI program error: %v", err)
			}
			os.Exit(1)
		}

		if Verbose {
			log.Info("TUI stopped")
		}
	},
}

func init() {
	Cmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, L("Enable debug mode"))
	Cmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, L("Enable verbose output"))

	// Add subcommands
	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(ValidateCmd)
	Cmd.AddCommand(InspectCmd)
	Cmd.AddCommand(CheckCmd)
	Cmd.AddCommand(DumpCmd)
	Cmd.AddCommand(HelpCmd)
}
