package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui/tui"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:       "validate <tui-id>",
	Short:     L("Validate a TUI configuration"),
	Long:      L("Validate a specific TUI configuration and show validation details"),
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"<tui-id>"},
	Run:       runValidate,
}

func runValidate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("usage: yao tui validate <tui-name>\n")
		return
	}

	tuiName := args[0]
	log.Info("Validating TUI: %s", tuiName)

	// Boot and load Yao configuration
	Boot()

	// Load application engine (required for application.App to be initialized)
	_, err := engine.Load(config.Conf, engine.LoadOption{
		Action: "tui",
	}, nil)
	if err != nil {
		fmt.Printf("Error: Failed to load engine: %v\n", err)
		return
	}

	// Load TUI configurations
	err = tui.Load(config.Conf)
	if err != nil {
		fmt.Printf("Error: Failed to load TUI configurations: %v\n", err)
		return
	}

	// Get the specific TUI configuration
	cfg := tui.Get(tuiName)
	if cfg == nil {
		fmt.Printf("Error: TUI not found: %s\n", tuiName)
		return
	}

	fmt.Printf("\n")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("TUI Validation Report\n")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("File:        %s.tui.yao\n", tuiName)
	fmt.Printf("Name:        %s\n", cfg.Name)
	fmt.Println(strings.Repeat("-", 70))

	// ✅ 使用真正的TUI ConfigValidator
	registry := tui.GetGlobalRegistry()
	validator := tui.NewConfigValidator(cfg, registry)

	if !validator.Validate() {
		fmt.Printf("\n❌ TUI validation FAILED for '%s'\n", cfg.Name)
		fmt.Printf("\nValidation Summary:\n")
		fmt.Println(validator.GetErrorSummary())

		// Show errors and warnings
		errors := validator.GetErrors()
		warnings := validator.GetWarnings()

		if len(errors) > 0 {
			fmt.Printf("\nErrors (%d):\n", len(errors))
			for i, err := range errors {
				fmt.Printf("  %d. %s: %s\n", i+1, err.Path, err.Message)
			}
		}

		if len(warnings) > 0 {
			fmt.Printf("\nWarnings (%d):\n", len(warnings))
			for i, warn := range warnings {
				fmt.Printf("  %d. %s: %s\n", i+1, warn.Path, warn.Message)
			}
		}

		return
	}

	// Show warnings if any
	warnings := validator.GetWarnings()
	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  Valid but with warnings (%d):\n", len(warnings))
		for i, warn := range warnings {
			fmt.Printf("  %d. %s: %s\n", i+1, warn.Path, warn.Message)
		}
	}

	fmt.Printf("\n✅ TUI '%s' is valid\n", cfg.Name)
	fmt.Printf("\nConfiguration Details:\n")
	fmt.Printf("  File:      %s.tui.yao\n", tuiName)
	fmt.Printf("  Direction: %s\n", cfg.Layout.Direction)
	fmt.Printf("  Components: %d\n", len(cfg.Layout.Children))

	// Show component types
	if len(cfg.Layout.Children) > 0 {
		fmt.Printf("\nComponent List:\n")
		for i, child := range cfg.Layout.Children {
			fmt.Printf("  %2d. %-20s", i+1, child.Type)
			if child.ID != "" {
				fmt.Printf(" (ID: %s)", child.ID)
			}
			if child.Bind != "" {
				fmt.Printf("  Bind: %s", child.Bind)
			}
			fmt.Println()
		}
	}

	// Show data bindings
	if len(cfg.Data) > 0 {
		if len(cfg.Data) <= 5 {
			fmt.Printf("\nInitial Data:\n")
			for key, value := range cfg.Data {
				strValue, _ := json.MarshalIndent(value, "", "    ")
				fmt.Printf("  %s: %s\n", key, strValue)
			}
		} else {
			fmt.Printf("\nInitial Data has %d key-value pairs\n", len(cfg.Data))
		}
	}
}
