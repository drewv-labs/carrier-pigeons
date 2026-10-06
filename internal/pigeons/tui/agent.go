package tui

import (
	"context"
	"fmt"
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

// Custom message type to receive the LLM payload back into the UI thread
type llmResponseMsg struct {
	content string
	err     error
}

type AgentModel struct {
	llm            adapters.LLMAdapter
	clusterContext string
	chatHistory    []adapters.Message // Keeps strict JSON format for the API
	uiHistory      []string           // Formatted strings for the terminal screen
	userInput      string
	waiting        bool
	quitting       bool
}

func NewAgentModel(ctx string, llm adapters.LLMAdapter) AgentModel {
	return AgentModel{
		llm:            llm,
		clusterContext: ctx,
		chatHistory:    []adapters.Message{},
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

	// Handle Keystrokes
	case tea.KeyMsg:
		if m.waiting && msg.String() != "ctrl+c" && msg.String() != "esc" {
			return m, nil // Block typing while the LLM is thinking
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if strings.TrimSpace(m.userInput) == "" {
				return m, nil
			}

			// 1. Record the user's prompt
			userText := m.userInput
			m.chatHistory = append(m.chatHistory, adapters.Message{Role: "user", Content: userText})
			m.uiHistory = append(m.uiHistory, userStyle.Render("Drew: "+userText))

			// 2. Lock the input and fire the background LLM command
			m.userInput = ""
			m.waiting = true

			return m, m.fetchLLMResponse()

		case "backspace":
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}

		default:
			m.userInput += msg.String()
		}

	// Handle Background LLM Responses
	case llmResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.uiHistory = append(m.uiHistory, agentStyle.Render(fmt.Sprintf("🤖 PigeonCare+ Error: %v", msg.err)))
		} else {
			m.chatHistory = append(m.chatHistory, adapters.Message{Role: "assistant", Content: msg.content})
			m.uiHistory = append(m.uiHistory, agentStyle.Render("🤖 PigeonCare+: "+msg.content))
		}
		return m, nil
	}

	return m, nil
}

// fetchLLMResponse runs the inference call in a background goroutine
func (m AgentModel) fetchLLMResponse() tea.Cmd {
	return func() tea.Msg {
		systemPrompt := "You are PigeonCare+, an expert site reliability engineer. Here is the live topological cluster failure state:\n" + m.clusterContext

		reply, err := m.llm.Chat(context.Background(), systemPrompt, m.chatHistory)
		return llmResponseMsg{content: reply, err: err}
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

	if m.waiting {
		sb.WriteString("\n" + waitStyle.Render("Agent is analyzing metrics..."))
	} else {
		sb.WriteString("\n" + promptStyle.Render("Ask PigeonCare+ > ") + m.userInput)
	}

	return sb.String()
}
