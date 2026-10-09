package cli

import (
	"github.com/spf13/cobra"
)

var runnerCmd = &cobra.Command{
	Use:   "runner [subcommand]",
	Short: "Dispatch remote execution runners across edge nodes",
}

func init() {
	rootCmd.AddCommand(runnerCmd)
}
