package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	a2a "github.com/inference-gateway/a2a-debugger/a2a"
	client "github.com/inference-gateway/adk/client"
	adk "github.com/inference-gateway/adk/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cobra "github.com/spf13/cobra"
)

// Interactive mode types
type sessionMode int

const (
	StreamingMode sessionMode = iota
	BackgroundMode
)

type chatModel struct {
	// UI State
	input        string
	messages     []chatMessage
	viewport     viewport
	mode         sessionMode
	ready        bool
	quitting     bool

	// A2A State
	contextID    string
	taskID       string
	client       client.A2AClient

	// Status
	statusLine   string
	isWaiting    bool
	lastResponse time.Time

	// Config
	width  int
	height int
}

type chatMessage struct {
	content   string
	role      string
	timestamp time.Time
	isError   bool
}

type viewport struct {
	content []string
	offset  int
}

// Bubble Tea messages
type tickMsg time.Time
type responseMsg struct {
	message string
	error   error
}
type streamEventMsg struct {
	event interface{}
	error error
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive chat mode with A2A server",
	Long: `Start an interactive chat session with the A2A server.
Supports both streaming (realtime) and background (long running tasks) modes.

Use Ctrl+C to quit, Tab to switch between streaming/background modes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ensureA2AClient()

		// Test connection first
		ctx := context.Background()
		_, err := a2aClient.GetAgentCard(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to A2A server: %w", err)
		}

		// Initialize model
		m := initialModel()
		m.client = a2aClient

		// Start Bubble Tea program
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func initialModel() chatModel {
	return chatModel{
		input:     "",
		messages:  []chatMessage{},
		viewport:  viewport{content: []string{}, offset: 0},
		mode:      StreamingMode,
		ready:     false,
		quitting:  false,
		contextID: fmt.Sprintf("ctx-%d", time.Now().Unix()),
		taskID:    "",
		statusLine: "Interactive Mode - Press Tab to switch modes, Ctrl+C to quit",
		isWaiting: false,
		width:     80,
		height:    24,
	}
}

func (m chatModel) Init() tea.Cmd {
	m.addSystemMessage("🚀 Interactive A2A Chat Session Started")
	m.addSystemMessage("📡 Mode: Streaming (realtime)")
	m.addSystemMessage("💬 Type your message and press Enter to send")
	m.addSystemMessage("⌨️  Press Tab to switch modes, Ctrl+C to quit")
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			if m.mode == StreamingMode {
				m.mode = BackgroundMode
				m.addSystemMessage("🔄 Switched to Background Mode (long running tasks)")
			} else {
				m.mode = StreamingMode
				m.addSystemMessage("🔄 Switched to Streaming Mode (realtime)")
			}
			return m, nil

		case "enter":
			if m.input != "" && !m.isWaiting {
				return m.sendMessage()
			}
			return m, nil

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		default:
			if !m.isWaiting {
				m.input += msg.String()
			}
			return m, nil
		}

	case responseMsg:
		m.isWaiting = false
		m.lastResponse = time.Now()
		if msg.error != nil {
			m.addErrorMessage(fmt.Sprintf("Error: %v", msg.error))
		} else {
			m.addAssistantMessage(msg.message)
		}
		return m, nil

	case streamEventMsg:
		if msg.error != nil {
			m.isWaiting = false
			m.addErrorMessage(fmt.Sprintf("Stream error: %v", msg.error))
		} else {
			m.handleStreamEvent(msg.event)
		}
		return m, nil

	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m chatModel) View() string {
	if !m.ready {
		return "Initializing interactive chat..."
	}

	var b strings.Builder

	// Header
	modeStr := "Streaming"
	if m.mode == BackgroundMode {
		modeStr = "Background"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	header := headerStyle.Render(fmt.Sprintf("🤖 A2A Interactive Chat - %s Mode", modeStr))
	b.WriteString(header + "\n")

	// Context info
	contextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Margin(1, 0)

	contextInfo := contextStyle.Render(fmt.Sprintf("Context ID: %s", m.contextID))
	if m.taskID != "" {
		contextInfo += contextStyle.Render(fmt.Sprintf(" | Task ID: %s", m.taskID))
	}
	b.WriteString(contextInfo + "\n")

	// Messages area
	messagesHeight := m.height - 8 // Reserve space for header, input, etc.
	messages := m.renderMessages(messagesHeight)
	b.WriteString(messages)

	// Status line
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Italic(true).
		Margin(1, 0)

	status := m.statusLine
	if m.isWaiting {
		status = "⏳ Waiting for response..."
	}
	b.WriteString(statusStyle.Render(status) + "\n")

	// Input area
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1)

	prompt := "💬 "
	if m.isWaiting {
		prompt = "⏳ "
	}

	input := inputStyle.Render(fmt.Sprintf("%s%s", prompt, m.input))
	b.WriteString(input)

	return b.String()
}

func (m *chatModel) sendMessage() (tea.Model, tea.Cmd) {
	message := strings.TrimSpace(m.input)
	if message == "" {
		return *m, nil
	}

	m.addUserMessage(message)
	m.input = ""
	m.isWaiting = true

	messageID := fmt.Sprintf("msg-%d", time.Now().Unix())

	params := adk.MessageSendParams{
		Message: adk.Message{
			Kind:      "message",
			MessageID: messageID,
			Role:      "user",
			Parts: []adk.Part{
				map[string]interface{}{
					"kind": "text",
					"text": message,
				},
			},
		},
	}

	params.Message.ContextID = &m.contextID
	if m.taskID != "" {
		params.Message.TaskID = &m.taskID
	}

	if m.mode == StreamingMode {
		return *m, m.sendStreamingMessage(params)
	} else {
		return *m, m.sendBackgroundMessage(params)
	}
}

func (m *chatModel) sendBackgroundMessage(params adk.MessageSendParams) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := m.client.SendTask(ctx, params)
		if err != nil {
			return responseMsg{message: "", error: handleA2AError(err, "message/send")}
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return responseMsg{message: "", error: fmt.Errorf("failed to marshal response: %w", err)}
		}

		var task adk.Task
		if err := json.Unmarshal(resultBytes, &task); err != nil {
			return responseMsg{message: "", error: fmt.Errorf("failed to unmarshal task: %w", err)}
		}

		// Update task ID for future messages
		m.taskID = task.ID

		// Extract response message
		if task.Status.Message != nil && len(task.Status.Message.Parts) > 0 {
			var responseText strings.Builder
			for _, part := range task.Status.Message.Parts {
				if partMap, ok := part.(map[string]interface{}); ok {
					if kind, ok := partMap["kind"].(string); ok && kind == "text" {
						if text, ok := partMap["text"].(string); ok {
							responseText.WriteString(text)
						}
					}
				}
			}
			return responseMsg{message: responseText.String(), error: nil}
		}

		return responseMsg{message: fmt.Sprintf("Task submitted (Status: %s)", task.Status.State), error: nil}
	}
}

func (m *chatModel) sendStreamingMessage(params adk.MessageSendParams) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		eventChan := make(chan interface{}, 100)

		go func() {
			defer close(eventChan)
			err := m.client.SendTaskStreaming(ctx, params, eventChan)
			if err != nil {
				eventChan <- streamEventMsg{event: nil, error: handleA2AError(err, "message/send")}
			}
		}()

		// Process first event
		select {
		case event := <-eventChan:
			if event == nil {
				return streamEventMsg{event: nil, error: fmt.Errorf("no response received")}
			}
			return streamEventMsg{event: event, error: nil}
		case <-time.After(30 * time.Second):
			return streamEventMsg{event: nil, error: fmt.Errorf("timeout waiting for response")}
		}
	}
}

func (m *chatModel) handleStreamEvent(event interface{}) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		m.addErrorMessage(fmt.Sprintf("Failed to parse stream event: %v", err))
		return
	}

	var genericEvent map[string]interface{}
	if err := json.Unmarshal(eventJSON, &genericEvent); err != nil {
		m.addErrorMessage(fmt.Sprintf("Failed to unmarshal stream event: %v", err))
		return
	}

	kind, ok := genericEvent["kind"].(string)
	if !ok {
		m.addErrorMessage("Stream event missing kind field")
		return
	}

	switch kind {
	case "status-update":
		var statusEvent a2a.TaskStatusUpdateEvent
		if err := json.Unmarshal(eventJSON, &statusEvent); err != nil {
			m.addErrorMessage(fmt.Sprintf("Failed to parse status event: %v", err))
			return
		}

		// Update task ID
		if m.taskID == "" {
			m.taskID = statusEvent.TaskID
		}

		// Process message parts
		if statusEvent.Status.Message != nil && len(statusEvent.Status.Message.Parts) > 0 {
			var responseText strings.Builder
			for _, part := range statusEvent.Status.Message.Parts {
				if partMap, ok := part.(map[string]interface{}); ok {
					if kind, ok := partMap["kind"].(string); ok && kind == "text" {
						if text, ok := partMap["text"].(string); ok {
							responseText.WriteString(text)
						}
					}
				}
			}
			if responseText.Len() > 0 {
				m.addAssistantMessage(responseText.String())
			}
		}

		if statusEvent.Final {
			m.isWaiting = false
		}

	case "artifact-update":
		var artifactEvent a2a.TaskArtifactUpdateEvent
		if err := json.Unmarshal(eventJSON, &artifactEvent); err != nil {
			m.addErrorMessage(fmt.Sprintf("Failed to parse artifact event: %v", err))
			return
		}

		artifactInfo := fmt.Sprintf("📄 Artifact: %s", artifactEvent.Artifact.ArtifactID)
		if artifactEvent.Artifact.Name != nil {
			artifactInfo += fmt.Sprintf(" (%s)", *artifactEvent.Artifact.Name)
		}
		m.addSystemMessage(artifactInfo)

		// Update task ID
		if m.taskID == "" {
			m.taskID = artifactEvent.TaskID
		}
	}
}

func (m *chatModel) addUserMessage(content string) {
	m.addMessage(chatMessage{
		content:   content,
		role:      "user",
		timestamp: time.Now(),
		isError:   false,
	})
}

func (m *chatModel) addAssistantMessage(content string) {
	m.addMessage(chatMessage{
		content:   content,
		role:      "assistant",
		timestamp: time.Now(),
		isError:   false,
	})
}

func (m *chatModel) addSystemMessage(content string) {
	m.addMessage(chatMessage{
		content:   content,
		role:      "system",
		timestamp: time.Now(),
		isError:   false,
	})
}

func (m *chatModel) addErrorMessage(content string) {
	m.addMessage(chatMessage{
		content:   content,
		role:      "error",
		timestamp: time.Now(),
		isError:   true,
	})
}

func (m *chatModel) addMessage(msg chatMessage) {
	m.messages = append(m.messages, msg)
}

func (m *chatModel) renderMessages(height int) string {
	if len(m.messages) == 0 {
		return ""
	}

	var lines []string
	for _, msg := range m.messages {
		lines = append(lines, m.renderMessage(msg)...)
	}

	// Handle viewport scrolling
	start := 0
	if len(lines) > height {
		start = len(lines) - height
	}

	var result strings.Builder
	for i := start; i < len(lines) && i < start+height; i++ {
		result.WriteString(lines[i] + "\n")
	}

	return result.String()
}

func (m *chatModel) renderMessage(msg chatMessage) []string {
	var style lipgloss.Style
	var prefix string

	switch msg.role {
	case "user":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		prefix = "👤 You: "
	case "assistant":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		prefix = "🤖 Agent: "
	case "system":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		prefix = "ℹ️  "
	case "error":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		prefix = "❌ "
	}

	if msg.isError {
		style = style.Foreground(lipgloss.Color("9"))
	}

	// Wrap long messages
	maxWidth := m.width - 10
	if maxWidth < 50 {
		maxWidth = 50
	}

	lines := wrapText(msg.content, maxWidth)
	var styledLines []string

	for i, line := range lines {
		if i == 0 {
			styledLines = append(styledLines, style.Render(prefix+line))
		} else {
			styledLines = append(styledLines, style.Render("    "+line))
		}
	}

	return styledLines
}

// Simple text wrapping function
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}