package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "mdocker/internal/container"
)

var unfreezeCmd = &cobra.Command{
    Use:   "unfreeze <pid>",
    Short: "Unfreeze a frozen container",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        if err := container.UnfreezeByPID(args[0]); err != nil {
            fmt.Println("Error:", err)
        }
    },
}
