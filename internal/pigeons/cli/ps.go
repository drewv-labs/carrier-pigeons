package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeons/adapters"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeons/tui"
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
				fmt.Println("\n[+] Handoff approved. Booting PigeonCare+...")

				var contextBuilder strings.Builder
				for group, issues := range degradedGroups {
					contextBuilder.WriteString(fmt.Sprintf("Group [%s]: %d issues -> %s\n", group, len(issues), strings.Join(issues, ", ")))
				}

				// 1. Connect to Postgres (using your local connection string)
				dbURL := "postgres://drewv:ctd_password@localhost:5432/edge_ledger"
				store, err := ledger.NewStore(context.Background(), dbURL)
				if err != nil {
					fmt.Printf("Failed to connect to ledger: %v\n", err)
					os.Exit(1)
				}
				defer store.Close()

				// 2. Initialize the LLM with the active database connection
				llmCfg := adapters.LLMConfig{
					Provider: "ollama",
					Endpoint: "http://localhost:11434",
					Model:    "qwen2.5-coder",
					Store:    store, // Injected!
				}
				agentLLM := adapters.NewLLMAdapter(llmCfg)

				// 3. Boot Bubble Tea
				p := tea.NewProgram(tui.NewAgentModel(contextBuilder.String(), agentLLM))
				if _, err := p.Run(); err != nil {
					fmt.Printf("Fatal error launching PigeonCare+: %v\n", err)
					os.Exit(1)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.Flags().BoolVarP(&verydetailed, "verydetailed", "v", false, "Include detailed status qualifiers")
	psCmd.Flags().BoolVar(&addSupport, "add-support", false, "Prompt for AI agent support on abnormal node states")
}
