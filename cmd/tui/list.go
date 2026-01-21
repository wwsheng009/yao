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

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: L("List all loaded TUI configurations"),
	Long:  L("List all loaded TUI configurations with details"),
	Run:   runList,
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
