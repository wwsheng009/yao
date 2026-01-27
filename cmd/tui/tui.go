package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui/tui"
)
var Debug bool
var Verbose bool

// Cmd represents the tui command
var Cmd = &cobra.Command{
	Use:   "tui <tui-name> [args...]",
	Short: L("Run a terminal user interface"),
	Long: L("Run a terminal user interface defined in .tui.yao files") +
		L("\n\n") +
		L("With external data (wrap JSON in quotes with :: prefix):\n") +
		L("  yao tui myapp '::{\"key\":\"value\"}'\n") +
		L("  yao tui myapp '::{\"userId\":123,\"userName\":\"John\"}'\n") +
		L("\n") +
		L("For escaping the :: prefix, use '\\::{' at the start"),
	Args: cobra.MinimumNArgs(1),
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

		// Parse external data arguments (JSON string with :: prefix)
		var externalData map[string]interface{}
		for i, arg := range args {
			if i == 0 {
				continue // Skip tuiID
			}

			// Check if argument is a JSON string with :: prefix
			// Format: '::{"key":"value"}'
			if strings.HasPrefix(arg, "::") {
				// Remove the :: prefix to get the JSON string
				jsonStr := strings.TrimPrefix(arg, "::")

				// Try to parse as JSON
				var v map[string]interface{}
				err := jsoniter.Unmarshal([]byte(jsonStr), &v)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s Failed to parse external data JSON: %v\n", color.RedString("Error:"), err)
					os.Exit(1)
				}

				// Merge external data (later args override previous ones)
				if externalData == nil {
					externalData = v
				} else {
					for k, val := range v {
						externalData[k] = val
					}
				}

				if Verbose {
					fmt.Printf("TUI external data[%d]: parsed %d keys\n", i-1, len(v))
					if Debug {
						for k, val := range v {
							fmt.Printf("  %s: %v\n", k, val)
						}
					}
				}

			} else if strings.HasPrefix(arg, "\\::") {
				// Escaped :: prefix - treat as literal string "::"
				// Remove the backslash and restore the ::
				literalStr := "::" + strings.TrimPrefix(arg, "\\::")
				if externalData == nil {
					externalData = make(map[string]interface{})
				}
				// Store as _args for later access if needed
				if argsKey, exists := externalData["_args"]; !exists {
					externalData["_args"] = []interface{}{literalStr}
				} else {
					externalData["_args"] = append(argsKey.([]interface{}), literalStr)
				}

				if Verbose {
					fmt.Printf("TUI arg[%d]: %s (escaped)\n", i-1, literalStr)
				}
			} else {
				// Regular string argument
				if externalData == nil {
					externalData = make(map[string]interface{})
				}
				// Store as _args for later access if needed
				if argsKey, exists := externalData["_args"]; !exists {
					externalData["_args"] = []interface{}{arg}
				} else {
					externalData["_args"] = append(argsKey.([]interface{}), arg)
				}

				if Verbose {
					fmt.Printf("TUI arg[%d]: %s\n", i-1, arg)
				}
			}
		}

		// Validate and flatten external data before merging
		if externalData != nil {
			var err error
			externalData, err = tui.ValidateAndFlattenExternal(externalData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Failed to validate external data: %v\n", color.RedString("Error:"), err)
				os.Exit(1)
			}
		}

		// Merge external data into config.Data (external data takes precedence)
		if externalData != nil && len(externalData) > 0 {
			if cfg.Data == nil {
				cfg.Data = make(map[string]interface{})
			}
			// Merge external data - external data overrides static data
			for k, v := range externalData {
				cfg.Data[k] = v
			}

			if Verbose {
				fmt.Printf("Merged %d external data keys into TUI configuration\n", len(externalData))
			}
		}

		if Verbose {
			if cfg.Name != "" {
				log.Info("Running TUI: %s (%s)", tuiID, cfg.Name)
			}
		}

		// Load TUI defaults if available (optional, static defaults that can be overridden)
		defaults := tui.LoadTUIDefaults(tuiID)
		if len(defaults) > 0 {
			if cfg.Data == nil {
				cfg.Data = make(map[string]interface{})
			}
			// Merge defaults into config data (external data will override defaults)
			// defaults have lower priority than config data, external has highest
			cfg.Data = tui.MergeData(cfg.Data, defaults, false) // priorityHigher = false
			if Verbose {
				log.Info("Loaded %d default values for TUI: %s", len(defaults), tuiID)
			}
		}

		// Prepare initial state using unified state management
		// This is the ONLY entry point for preparing TUI state data
		// Data flow (priority, highest to lowest):
		//   1. External parameters (jsonStr from command-line args)
		//   2. Static configuration data (from .tui.yao file)
		//   3. TUI defaults (from data/tui/*.json)
		// All data is flattened to support dot-notation access
		tui.PrepareInitialState(cfg, externalData)

		if len(externalData) > 0 {
			log.Info("External data loaded with %d keys", len(externalData))
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
		model := tui.NewModel(cfg, nil)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),       // Use alternate screen buffer
			tea.WithMouseCellMotion(), // Enable mouse support
			tea.WithContext(ctx),      // Add context for graceful shutdown
		)
		// Set the program reference after creation
		model.Program = program

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
