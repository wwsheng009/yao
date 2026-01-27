package tui

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/engine"
	"github.com/yaoapp/yao/tui/tui"
)

// CheckCmd represents the check command
var CheckCmd = &cobra.Command{
	Use:       "check <tui-id>",
	Short:     L("Check a TUI configuration"),
	Long:      L("Initialize a TUI and check for initialization errors"),
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"<tui-id>"},
	Run:       runCheck,
}

func runCheck(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("usage: yao tui check <tui-name>\n")
		return
	}

	tuiName := args[0]
	log.Info("Checking TUI: %s", tuiName)

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
	_ = tui.Get(tuiName)
	if !tui.IsLoaded(tuiName) {
		fmt.Printf("TUI not found: %s\n", tuiName)
		return
	}

	fmt.Printf("\nChecking TUI: %s\n", tuiName)
	fmt.Println("TUI loaded successfully")

	fmt.Println(" Ready to run with:")
	fmt.Printf("   yao tui %s\n", tuiName)
}
