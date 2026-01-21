package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yaoapp/kun/utils"
	"github.com/yaoapp/yao/config"
)

var envShowSystem bool

var envCmd = &cobra.Command{
	Use:   "env",
	Short: L("Show environment variables"),
	Long:  L("Show environment variables"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(L(" OS ENV \"YAO_ROOT\": %s\n"), os.Getenv("YAO_ROOT"))
		// Use appPath if set, otherwise auto-detect from cwd
		var checkPath string
		var root string
		var err error

		if appPath != "" {
			checkPath = appPath
		} else {
			// Auto-detect from current working directory
			cwd, err := os.Getwd()
			if err != nil {
				color.Red(L("Error getting current directory: %s\n"), err.Error())
				return
			}
			checkPath = cwd
		}

		// Find app root from the determined path
		root, err = findAppRootFromPath(checkPath)
		if err == nil {
			// Display the detected root path
			color.Green(L("✓ App root found:"))
			fmt.Printf(L("  Path: %s\n"), root)
			fmt.Println("")

			// Source Analysis
			// Priority: 1. --app flag 2. YAO_ROOT env 3. Auto-detect
			color.Cyan(L("  Root Source Analysis:"))
			envRoot := os.Getenv("YAO_ROOT")
			
			// 1. Command Line --app
			fmt.Print(L("    > 1. Command Line --app: "))
			if appPath != "" {
				color.Green("%s (High Priority)", appPath)
			} else {
				color.Yellow("(Not Set)")
			}

			// 2. Environment YAO_ROOT
			fmt.Print(L("    > 2. Environment YAO_ROOT: "))
			if envRoot != "" {
				color.Green("%s", envRoot)
			} else {
				color.Yellow("(Not Set)")
			}

			// 3. Auto-detect
			fmt.Print(L("    > 3. Auto-detect: "))
			color.White("%s", root)
			fmt.Println("")

			// Check for app configuration files
			color.Cyan(L("  Configuration Files:"))
			for _, appFile := range []string{"app.yao", "app.json", "app.jsonc"} {
				appFilePath := fmt.Sprintf("%s/%s", root, appFile)
				if _, err := os.Stat(appFilePath); err == nil {
					fmt.Print(L("    ✓ Found: "))
					color.Green("%s", appFile)
				}
			}
			fmt.Println("")
		} else {
			color.Yellow(L("! App root not found for path:\n"))
			fmt.Printf(L("  Path: %s"), checkPath)
			fmt.Println("")
		}

		// Show environment variables
		Boot()


		fmt.Printf(L("  YAO_ROOT: %s\n"), config.Conf.Root)
		fmt.Println("")

		if envShowSystem {
			res := map[string]interface{}{
				"config": config.Conf,
				"system": getSystemEnv(),
			}
			utils.Dump(res)
		} else {
			utils.Dump(config.Conf)
		}
	},
	Example: `  yao env                  Show environment variables
  yao env --system          Show environment variables including system env`,
}

func init() {
	envCmd.Flags().BoolVarP(&envShowSystem, "system", "s", false, L("Show system environment variables"))
}

func getSystemEnv() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := splitEnv(e)
		if pair.key != "" {
			env[pair.key] = pair.value
		}
	}
	return env
}

type envPair struct {
	key   string
	value string
}

func splitEnv(envStr string) envPair {
	for i := 0; i < len(envStr); i++ {
		if envStr[i] == '=' {
			return envPair{
				key:   envStr[:i],
				value: envStr[i+1:],
			}
		}
	}
	return envPair{key: envStr, value: ""}
}
