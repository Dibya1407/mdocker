package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mdocker",
	Short: "A minimal container runtime",
	Long:  "mdocker is a learning container runtime built using Linux namespaces and cgroups",
}

func Execute() {
	err := rootCmd.Execute();
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
