package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeons/adapters"
)

var (
	agentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1)
	userStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("253"))
	systemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	waitStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

// Internal message types for the streaming loop
type streamTokenMsg struct {
	token string
	ch    <-chan string
}
type streamDoneMsg struct{}
type streamErrMsg struct{ err error }

type AgentModel struct {
	llm             adapters.LLMAdapter
	clusterContext  string
	chatHistory     []adapters.Message
	uiHistory       []string
	userInput       string
	currentResponse string // Holds the active string being streamed
	waiting         bool
	quitting        bool
}

func NewAgentModel(ctx string, llm adapters.LLMAdapter) AgentModel {
	return AgentModel{
		llm:            llm,
		clusterContext: ctx,
		uiHistory: []string{
			agentStyle.Render("🤖 PigeonCare+: I have analyzed the cluster topography. How can I assist with triage?"),
		},
	}
}

func (m AgentModel) Init() tea.Cmd {
	return nil
}

func (m AgentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.waiting && msg.String() != "ctrl+c" && msg.String() != "esc" {
			return m, nil // Lock input while streaming
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if strings.TrimSpace(m.userInput) == "" {
				return m, nil
			}

			userText := m.userInput
			m.chatHistory = append(m.chatHistory, adapters.Message{Role: "user", Content: userText})
			m.uiHistory = append(m.uiHistory, userStyle.Render("Drew: "+userText))

			// Append an empty string for the agent that we will update live
			m.uiHistory = append(m.uiHistory, agentStyle.Render("🤖 PigeonCare+: "))

			m.userInput = ""
			m.currentResponse = ""
			m.waiting = true

			return m, m.startStream()

		case "backspace":
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		default:
			m.userInput += msg.String()
		}

	case streamErrMsg:
		m.waiting = false
		m.uiHistory[len(m.uiHistory)-1] = agentStyle.Render("🤖 PigeonCare+ Error: " + msg.err.Error())
		return m, nil

	case streamTokenMsg:
		m.currentResponse += msg.token
		// Overwrite the last message in the UI slice with the growing string
		m.uiHistory[len(m.uiHistory)-1] = agentStyle.Render("🤖 PigeonCare+: " + m.currentResponse)
		// Instantly request the next token
		return m, waitForToken(msg.ch)

	case streamDoneMsg:
		m.waiting = false
		m.chatHistory = append(m.chatHistory, adapters.Message{Role: "assistant", Content: m.currentResponse})
		return m, nil
	}

	return m, nil
}

// startStream initiates the HTTP request and grabs the channel
func (m AgentModel) startStream() tea.Cmd {
	return func() tea.Msg {
		sysPrompt := "You are PigeonCare+, an expert site reliability engineer. Live cluster context:\n" + m.clusterContext
		ch, err := m.llm.StreamChat(context.Background(), sysPrompt, m.chatHistory)
		if err != nil {
			return streamErrMsg{err}
		}
		// Kick off the channel reader
		return waitForToken(ch)()
	}
}

// waitForToken is a recursive loop that pulls from the channel without blocking the main UI thread
func waitForToken(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		token, ok := <-ch
		if !ok {
			return streamDoneMsg{} // Channel closed, stream is finished
		}
		return streamTokenMsg{token: token, ch: ch}
	}
}

func (m AgentModel) View() string {
	if m.quitting {
		return "PigeonCare+ session terminated.\n"
	}

	var sb strings.Builder
	sb.WriteString("Topological Fault Context Injected:\n")
	sb.WriteString(systemStyle.Render(m.clusterContext) + "\n\n")

	for _, msg := range m.uiHistory {
		sb.WriteString(msg + "\n")
	}

	if !m.waiting {
		sb.WriteString("\n" + promptStyle.Render("Ask PigeonCare+ > ") + m.userInput)
	}

	return sb.String()
}
