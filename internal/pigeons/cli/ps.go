package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	verydetailed bool
	addSupport   bool
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Show active cluster status and telemetry health",
	Run: func(cmd *cobra.Command, args []string) {
		type NodeStatus struct {
			Node   string
			Group  string
			Role   string
			State  string
			Detail string
		}

		nodes := []NodeStatus{
			{"jakebase", "in-stack", "coop", "running", "clean"},
			{"radiawrt", "in-stack", "neer", "running", "slow"},
			{"caroline-v", "in-stack", "neer", "running", "clean"},
			{"margistars", "in-stack", "neer", "running", "clean"},
			{"adactrl", "on-scope", "neer", "running", "hot"},
			{"maryguider", "on-scope", "neer", "oof!!!", "last-seen={registry}"},
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		if verydetailed {
			fmt.Fprintln(w, "node\t| group\t| role\t| very detailed status")
		} else {
			fmt.Fprintln(w, "node\t| group\t| role\t| status")
		}
		fmt.Fprintln(w, "__________________________________________\t")

		degradedGroups := make(map[string][]string)
		needsSupport := false

		for _, n := range nodes {
			if verydetailed {
				// Print both State and Detail when -v is passed
				statusStr := fmt.Sprintf("%s %s", n.State, n.Detail)
				fmt.Fprintf(w, "%s\t| %s\t| %s\t| %s\n", n.Node, n.Group, n.Role, statusStr)
			} else {
				// Only print the base State by default
				fmt.Fprintf(w, "%s\t| %s\t| %s\t| %s\n", n.Node, n.Group, n.Role, n.State)
			}

			if strings.HasPrefix(n.State, "oof") || n.Detail != "clean" {
				needsSupport = true
				issue := fmt.Sprintf("%s (%s)", n.Node, n.Detail)
				degradedGroups[n.Group] = append(degradedGroups[n.Group], issue)
			}
		}

		w.Flush()

		if addSupport && needsSupport {
			fmt.Println("\n⚠ Abnormal node states detected in the cluster.")
			fmt.Print("Contact PigeonCare+ Agent support? [Y/n] ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				fmt.Println("\n[+] Initializing PigeonCommand Bubble Tea Panel...")
				fmt.Println("--------------------------------------------------")

				var contextBuilder strings.Builder
				contextBuilder.WriteString("System Context:\n")
				for group, issues := range degradedGroups {
					contextBuilder.WriteString(fmt.Sprintf(" - Group [%s] has %d issues: %s\n", group, len(issues), strings.Join(issues, ", ")))
				}

				fmt.Printf("🤖 PigeonCare+: \"I see multiple correlated issues in the [%s] group.\n", "on-scope")
				fmt.Println("   Since maryguider dropped completely and adactrl is thermal throttling,")
				fmt.Println("   this looks like an environmental or physical power failure on the telescope mount.")
				fmt.Println("   Pulling Postgres logs to verify if temperature spiked before the drop...")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.Flags().BoolVarP(&verydetailed, "verydetailed", "v", false, "Include detailed status qualifiers")
	psCmd.Flags().BoolVar(&addSupport, "add-support", false, "Prompt for AI agent support on abnormal node states")
}
