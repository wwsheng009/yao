package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui"
)

var useTextOutput bool

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: L("List all loaded TUI configurations"),
	Long:  L("List all TUI configurations with details"),
	Run:   runList,
}

func init() {
	ListCmd.Flags().BoolVarP(&useTextOutput, "text", "t", false, "Use text output instead of TUI interface")
}

func runList(cmd *cobra.Command, args []string) {
	log.Info("Listing all TUI configurations...")

	// Boot Yao if not already done
	Boot()

	// Load application engine (required for application.App to be initialized)
	_, err := engine.Load(config.Conf, engine.LoadOption{
		Action: "tui",
	}, nil)
	if err != nil {
		fmt.Printf("Error loading engine: %v\n", err)
		return
	}

	// Load TUI configurations
	err = tui.Load(config.Conf)
	if err != nil {
		fmt.Printf("Error loading TUI configurations: %v\n", err)
		return
	}

	// Check if tui-list.tui.yao exists and user didn't request text output
	if !useTextOutput {
		if listCfg := tui.Get("__yao.tui-list"); listCfg != nil {
			// Use TUI interface to display the list
			runTUIList(listCfg)
			return
		}
		// Fall through to text output if tui-list.tui.yao doesn't exist
	}

	// Text output (original behavior)
	ids := tui.List()

	if len(ids) == 0 {
		fmt.Println("No TUI configurations found in tuis/")
		return
	}

	fmt.Printf("\nFound %d TUI configuration(s):\n", len(ids))
	fmt.Println(strings.Repeat("=", 60))

	// Iterate through all TUI configurations
	for _, id := range ids {
		cfg := tui.Get(id)

		fmt.Printf("\n  ID:       %s\n", id)
		if cfg != nil && cfg.Name != "" {
			fmt.Printf("  Name:     %s\n", cfg.Name)
		}

		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Println("\nRun a specific TUI:")
	fmt.Printf("  yao tui <tui-name>\n")
}

// runTUIList runs the TUI list interface
func runTUIList(cfg *tui.Config) {
	// Get the list of all TUI configurations
	tuiIDs := tui.List()

	// Prepare TUI list data using PrepareInitialState for consistency
	// This ensures data is flattened and merged using the unified approach
	externalData := prepareTUIListData(tuiIDs)

	// Prepare initial state using unified state management
	tui.PrepareInitialState(cfg, externalData)

	if Verbose {
		log.Info("Loaded %d TUI items for TUI list interface", len(tuiIDs))
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
			if Verbose {
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
}

// prepareTUIListData prepares TUI list data for passing to the TUI list interface
// This data will be merged and flattened using PrepareInitialState
func prepareTUIListData(tuiIDs []string) map[string]interface{} {
	// Use []interface{} to ensure proper type compatibility with list component
	// The list component's ParseListPropsWithBinding expects []interface{}, not []string
	tuiItems := make([]interface{}, 0, len(tuiIDs))
	for _, id := range tuiIDs {
		tuiCfg := tui.Get(id)
		if tuiCfg == nil {
			continue
		}

		item := map[string]interface{}{
			"id":      id,
			"name":    tuiCfg.Name,
			"title":   fmt.Sprintf("%s - %s", id, tuiCfg.Name), // Title for list display
			"command": fmt.Sprintf("yao tui %s", id),
		}

		// Add optional description if available in Data
		if tuiCfg.Data != nil {
			if desc, ok := tuiCfg.Data["description"]; ok {
				item["description"] = desc
			}
		}

		tuiItems = append(tuiItems, item)
	}

	// Return data that will be merged and flattened by PrepareInitialState
	// items is []interface{} containing map[string]interface{} which is required by list component
	externalData := map[string]interface{}{
		"items":      tuiItems,
		"tuiItems":   tuiItems,
		"totalCount": len(tuiItems),
	}

	return externalData
}

// tuiNames converts TUI item map to string array for list component
func tuiNames(items []map[string]interface{}) []string {
	names := make([]string, len(items))
	for i, item := range items {
		if name, ok := item["id"].(string); ok {
			names[i] = name
		} else {
			names[i] = fmt.Sprintf("%v", item)
		}
	}
	return names
}
