package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered nodes, their groups, and roles",
	Run: func(cmd *cobra.Command, args []string) {
		nodes := []struct{ Node, Group, Role string }{
			{"jakebase", "in-stack", "coop"},
			{"radiawrt", "in-stack", "pigeoneer"},
			{"caroline-v", "in-stack", "pigeoneer"},
			{"margistars", "in-stack", "pigeoneer"},
			{"adactrl", "on-scope", "pigeoneer"},
			{"maryguider", "on-scope", "pigeoneer"},
		}

		// tabwriter config: minwidth=0, tabwidth=0, padding=3, padchar=' '
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		fmt.Fprintln(w, "node\t| group\t| role")
		fmt.Fprintln(w, "__________________________________________\t")

		for _, n := range nodes {
			fmt.Fprintf(w, "%s\t| %s\t| %s\n", n.Node, n.Group, n.Role)
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
