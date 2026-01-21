package tui

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui"
)

// InspectCmd represents the inspect command
var InspectCmd = &cobra.Command{
	Use:       "inspect <tui-id>",
	Short:     L("Inspect a TUI configuration"),
	Long:      L("Inspect a TUI configuration and show detailed information"),
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"<tui-id>"},
	Run:       runInspect,
}

func runInspect(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("usage: yao tui inspect <tui-name>\n")
		return
	}

	tuiName := args[0]
	log.Info("Inspecting TUI: %s", tuiName)

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

	// Get the specific TUI
	cfg := tui.Get(tuiName)
	if cfg == nil {
		fmt.Printf("TUI not found: %s\n", tuiName)
		return
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("TUI Inspection Report\n")
	fmt.Printf("File:        %s.tui.yao\n", tuiName)
	if cfg.Name != "" {
		fmt.Printf("Name:        %s\n", cfg.Name)
	}
	fmt.Printf("ID:          %s\n", tuiName)
	fmt.Printf("%s\n", strings.Repeat("-", 70))

	// Print configuration details
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  ID:        %s\n", cfg.ID)
	fmt.Printf("  Name:      %s\n", cfg.Name)
	if cfg.LogLevel != "" {
		fmt.Printf("  LogLevel:  %s\n", cfg.LogLevel)
	}
	fmt.Printf("  Direction: %s\n", cfg.Layout.Direction)
	fmt.Printf("  Children:  %d\n", len(cfg.Layout.Children))
	if len(cfg.Bindings) > 0 {
		fmt.Printf("  Bindings:  %d\n", len(cfg.Bindings))
	}

	if len(cfg.Data) > 0 {
		fmt.Printf("  Data keys: %d\n", len(cfg.Data))
	}

	fmt.Printf("\n%s\n", strings.Repeat("-", 70))
}
