package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mdocker/internal/container"
)

var runCmd = &cobra.Command{
	Use:   "run [command] [args...]",
	Short: "Run a command in an isolated container",
	Long:  "Run a command inside a container using Linux namespaces",
	DisableFlagParsing: true,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runContainer(args);
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runContainer(args []string) error {
	return container.Run(args)
}
