package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

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
		dbURL := os.Getenv("PIGEONS_DB_URL")
		if dbURL == "" {
			dbURL = "postgres://drewv:ctd_password@localhost:5432/edge_ledger"
		}

		store, err := ledger.NewStore(context.Background(), dbURL)
		if err != nil {
			fmt.Printf("Fatal: Could not connect to ledger: %v\n", err)
			os.Exit(1)
		}
		// Deferring close here keeps the connection pool alive for PigeonCare+ if it boots
		defer store.Close()

		// DISTINCT ON efficiently grabs exactly one row (the most recent) per node
		query := `
			SELECT DISTINCT ON (node_id)
				node_id, component, status, event_timestamp
			FROM telemetric_ledger
			ORDER BY node_id, event_timestamp DESC;
		`

		rows, err := store.Query(context.Background(), query)
		if err != nil {
			fmt.Printf("Fatal: Failed to fetch cluster status: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		if verydetailed {
			fmt.Fprintln(w, "node\t| last component\t| state\t| detail")
		} else {
			fmt.Fprintln(w, "node\t| last component\t| state")
		}
		fmt.Fprintln(w, "__________________________________________________________________\t")

		var degradedNodes []string
		needsSupport := false

		for rows.Next() {
			var nodeID, component, dbStatus string
			var lastSeen time.Time

			if err := rows.Scan(&nodeID, &component, &dbStatus, &lastSeen); err == nil {
				state := "running"
				detail := "clean"

				// Dynamic health check based on ledger staleness
				timeSince := time.Since(lastSeen)
				if timeSince > 3*time.Minute {
					state = "oof!!!"
					detail = fmt.Sprintf("unreachable (%.0fm ago)", timeSince.Minutes())
					needsSupport = true
					degradedNodes = append(degradedNodes, fmt.Sprintf("%s (%s)", nodeID, detail))
				} else if strings.Contains(strings.ToLower(dbStatus), "error") || strings.Contains(strings.ToLower(dbStatus), "fail") {
					state = "degraded"
					detail = dbStatus
					needsSupport = true
					degradedNodes = append(degradedNodes, fmt.Sprintf("%s (%s)", nodeID, detail))
				}

				if verydetailed {
					fmt.Fprintf(w, "%s\t| %s\t| %s\t| %s\n", nodeID, component, state, detail)
				} else {
					fmt.Fprintf(w, "%s\t| %s\t| %s\n", nodeID, component, state)
				}
			}
		}
		rows.Close()
		w.Flush()

		if addSupport && needsSupport {
			fmt.Println("\n⚠ Abnormal node states detected in the cluster.")
			fmt.Print("Contact PigeonCare+ Agent support? [Y/n] ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				fmt.Println("\n[+] Handoff approved. Booting PigeonCare+...")

				// Inject exactly which nodes failed directly into Qwen's context window
				var contextBuilder strings.Builder
				contextBuilder.WriteString("The following nodes are currently degraded or unreachable:\n")
				for _, issue := range degradedNodes {
					contextBuilder.WriteString(fmt.Sprintf("- %s\n", issue))
				}

				llmCfg := adapters.LLMConfig{
					Provider: "ollama",
					Endpoint: "http://localhost:11434",
					Model:    "qwen2.5-coder",
					Store:    store,
				}
				agentLLM := adapters.NewLLMAdapter(llmCfg)

				p := tea.NewProgram(tui.NewAgentModel(contextBuilder.String(), agentLLM))
				if _, err := p.Run(); err != nil {
					fmt.Printf("Fatal error launching PigeonCare+: %v\n", err)
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
