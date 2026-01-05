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
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runContainer(args);
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

var (
    memLimit  string
    cpuLimit  int
    pidsLimit int
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&memLimit, "mem", "100M", "memory limit (e.g. 100M, 1G)")
	runCmd.Flags().IntVar(&cpuLimit, "cpu", 20, "cpu limit percentage")
	runCmd.Flags().IntVar(&pidsLimit, "pids", 32, "max number of processes")

}

func runContainer(args []string) error {
	if cpuLimit > 100 {
        return fmt.Errorf("--cpu cannot be greater than 100")
    }
    if pidsLimit > 100 {
        return fmt.Errorf("--pids cannot be greater than 100")
    }

	cfg := container.CgroupConfig{
    Memory: memLimit,
    CPU:    cpuLimit,
    Pids:   pidsLimit,
	}

	return container.Run(args, cfg)
}
