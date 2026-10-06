package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI Styling via Lipgloss
var (
	agentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(1)

	contextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
)

// AgentModel holds the state of our TUI application.
type AgentModel struct {
	clusterContext string
	chatHistory    []string
	userInput      string
	quitting       bool
}

// NewAgentModel initializes the TUI with the degraded node context.
func NewAgentModel(ctx string) AgentModel {
	return AgentModel{
		clusterContext: ctx,
		chatHistory: []string{
			"🤖 PigeonCare+: I analyzed the cluster topography. Multiple correlated physical faults detected.",
			"   Pulling Postgres ledger metrics to verify power draw right before the drop...",
		},
	}
}

// Init runs any initial I/O (like starting a loading spinner or making an initial MCP call).
func (m AgentModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming events like keystrokes.
func (m AgentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if strings.TrimSpace(m.userInput) != "" {
				// Append user message
				m.chatHistory = append(m.chatHistory, "Drew: "+m.userInput)
				// Placeholder for LLM response via MCP
				m.chatHistory = append(m.chatHistory, "🤖 PigeonCare+: [Querying MCP sidecar for: "+m.userInput+"]")
				m.userInput = ""
			}
			return m, nil
		case "backspace":
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		default:
			// Capture standard typing
			m.userInput += msg.String()
		}
	}
	return m, nil
}

// View renders the UI to the terminal every time the state updates.
func (m AgentModel) View() string {
	if m.quitting {
		return "PigeonCare+ session terminated. Returning to standard shell.\n"
	}

	var sb strings.Builder

	// Render the hidden context the LLM received
	sb.WriteString("Topological Fault Context Injected:\n")
	sb.WriteString(contextStyle.Render(m.clusterContext) + "\n\n")

	// Render the chat history
	for _, msg := range m.chatHistory {
		if strings.HasPrefix(msg, "🤖") {
			sb.WriteString(agentStyle.Render(msg) + "\n")
		} else {
			sb.WriteString(msg + "\n")
		}
	}

	// Render the active prompt
	sb.WriteString("\n" + promptStyle.Render("Ask PigeonCare+ > ") + m.userInput)

	return sb.String()
}
