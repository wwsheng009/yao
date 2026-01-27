package tui

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui/tui"
)

// DumpCmd represents the dump command
var DumpCmd = &cobra.Command{
	Use:       "dump <tui-id>",
	Short:     L("Dump TUI configuration as JSON"),
	Long:      L("Dump the raw TUI configuration JSON for debugging purposes"),
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"<tui-id>"},
	Run:       runDump,
}

func runDump(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("usage: yao tui dump <tui-name>\n")
		return
	}

	tuiName := args[0]
	log.Info("Dumping TUI configuration: %s", tuiName)

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

	// Dump as JSON structure
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("failed to serialize configuration: %v\n", err)
		return
	}

	fmt.Printf("Raw Configuration JSON for '%s' (%s.tui.yao):\n\n", tuiName, tuiName)
	fmt.Println(jsonData)
}
