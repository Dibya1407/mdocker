package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "mdocker/internal/container"
)

var killCmd = &cobra.Command{
    Use:   "kill <pid>",
    Short: "Kill a running container",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        if err := container.KillByPID(args[0]); err != nil {
            fmt.Println("Error:", err)
        }
    },
}
