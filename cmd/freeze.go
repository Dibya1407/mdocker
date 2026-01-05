package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "mdocker/internal/container"
)

var freezeCmd = &cobra.Command{
    Use:   "freeze <pid>",
    Short: "Freeze a running container",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        if err := container.FreezeByPID(args[0]); err != nil {
            fmt.Println("Error:", err)
        }
    },
}
