package cmd

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

var tuiDebug bool
var tuiVerbose bool

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui [TUI_ID]",
	Short: L("Run a terminal user interface"),
	Long:  L("Run a terminal user interface defined in .tui.yao files"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Boot()

		tuiID := args[0]

		if tuiVerbose {
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

		if tuiDebug && len(loadWarnings) > 0 {
			log.Warn("Load warnings: %v", loadWarnings)
		}

		// Load TUI configurations
		err = tui.Load(config.Conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to load TUI configurations: %v\n", color.RedString("Error:"), err)
			os.Exit(1)
		}

		if tuiVerbose {
			log.Info("Loaded %d TUI configuration(s)", tui.Count())
		}

		// Get TUI configuration
		cfg := tui.Get(tuiID)
		if cfg == nil {
			fmt.Fprintf(os.Stderr, "%s TUI not found: %s\n", color.RedString("Error:"), tuiID)
			fmt.Fprintf(os.Stderr, "\n%s Available TUIs:\n", color.YellowString("Hint:"))
			
			availableTUIs := tui.List()
			if len(availableTUIs) == 0 {
				fmt.Fprintf(os.Stderr, "  No TUI configurations found in tuis/ directory\n")
			} else {
				for _, id := range availableTUIs {
					fmt.Fprintf(os.Stderr, "  - %s\n", color.CyanString(id))
				}
			}
			os.Exit(1)
		}

		if tuiVerbose {
			log.Info("Running TUI: %s (%s)", tuiID, cfg.Name)
		}

		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start goroutine to monitor signals and cancel context
		go func() {
			<-sigChan
			if tuiDebug {
				log.Info("Received interrupt signal, gracefully shutting down...")
			}
			cancel()
		}()

		// Create Bubble Tea program with context
		model := tui.NewModel(cfg, nil)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),       // Use alternate screen buffer
			tea.WithMouseCellMotion(), // Enable mouse support
			tea.WithContext(ctx),      // Add context for graceful shutdown
		)

		// Set program reference in model for state updates
		model.Program = program

		// Run program
		if _, err := program.Run(); err != nil {
			if tuiDebug {
				log.Error("TUI program error: %v", err)
			}
			os.Exit(1)
		}

		if tuiVerbose {
			log.Info("TUI stopped")
		}
	},
}

// init initializes the TUI command
func init() {
	tuiCmd.PersistentFlags().BoolVarP(&tuiDebug, "debug", "d", false, L("Enable debug mode"))
	tuiCmd.PersistentFlags().BoolVarP(&tuiVerbose, "verbose", "v", false, L("Enable verbose output"))
}
