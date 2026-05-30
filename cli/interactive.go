package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	textinput "github.com/charmbracelet/bubbles/textinput"
	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	uuid "github.com/google/uuid"
	viper "github.com/spf13/viper"

	adk "github.com/inference-gateway/adk/types"
)

// backgroundPollInterval is how often a background task is polled for completion.
const backgroundPollInterval = 1 * time.Second

// chatMode selects how messages are exchanged with the A2A server.
type chatMode int

const (
	modeStreaming chatMode = iota
	modeBackground
)

func (m chatMode) String() string {
	if m == modeBackground {
		return "background"
	}
	return "streaming"
}

// msgSender identifies who authored a line in the transcript.
type msgSender int

const (
	senderUser msgSender = iota
	senderAgent
	senderSystem
)

type chatLine struct {
	sender msgSender
	text   string
}

// --- Bubble Tea messages ---

// streamStartedMsg carries the channel returned by SendTaskStreaming.
type streamStartedMsg struct {
	ch <-chan adk.JSONRPCSuccessResponse
}

// streamEventMsg carries a single event read from the streaming channel.
type streamEventMsg struct {
	resp adk.JSONRPCSuccessResponse
	ok   bool
	ch   <-chan adk.JSONRPCSuccessResponse
}

// taskSubmittedMsg is emitted in background mode once a task is accepted.
type taskSubmittedMsg struct {
	taskID    string
	contextID string
}

// taskPolledMsg carries the latest task snapshot while polling in background mode.
type taskPolledMsg struct {
	task adk.Task
	done bool
}

// agentErrorMsg surfaces an error from the A2A client as a system line.
type agentErrorMsg struct {
	err error
}

// --- styles ---

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	metaStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	userLabelStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	agentLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	bodyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	systemStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240"))
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	spinnerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)

// interactiveModel is the Bubble Tea model backing the chat interface.
type interactiveModel struct {
	input    textinput.Model
	viewport viewport.Model
	spinner  spinner.Model

	lines     []chatLine
	mode      chatMode
	serverURL string
	agentName string

	contextID  string
	lastTaskID string
	lastState  adk.TaskState

	waiting  bool
	agentBuf string
	replyIdx int

	width  int
	height int
	ready  bool
}

func newInteractiveModel(mode chatMode, serverURL, agentName, contextID string) interactiveModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message and press Enter"
	ti.Prompt = "> "
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	if agentName == "" {
		agentName = "Agent"
	}
	if contextID == "" {
		contextID = uuid.NewString()
	}

	m := interactiveModel{
		input:     ti,
		spinner:   sp,
		mode:      mode,
		serverURL: serverURL,
		agentName: agentName,
		contextID: contextID,
		replyIdx:  -1,
	}
	m.addLine(senderSystem, fmt.Sprintf("connected to %s · %s mode · context %s", serverURL, mode.String(), shortID(contextID)))
	m.addLine(senderSystem, "type a message and press Enter to begin")
	return m
}

func (m interactiveModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		m.ready = true
		m.refreshViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.waiting {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			return m.submit(text)
		case tea.KeyCtrlT:
			if !m.waiting {
				m.toggleMode()
				m.refreshViewport()
			}
			return m, nil
		}

	case streamStartedMsg:
		return m, readStreamCmd(msg.ch)

	case streamEventMsg:
		if !msg.ok {
			m.finishAgentReply()
			m.refreshViewport()
			return m, nil
		}
		m.applyStreamEvent(msg.resp)
		m.refreshViewport()
		return m, readStreamCmd(msg.ch)

	case taskSubmittedMsg:
		if msg.taskID != "" {
			m.lastTaskID = msg.taskID
		}
		if msg.contextID != "" {
			m.contextID = msg.contextID
		}
		m.addLine(senderSystem, fmt.Sprintf("task %s submitted, waiting for completion...", shortID(msg.taskID)))
		m.refreshViewport()
		return m, pollTaskCmd(m.lastTaskID)

	case taskPolledMsg:
		m.lastState = msg.task.Status.State
		if !msg.done {
			return m, pollTaskCmd(msg.task.ID)
		}
		m.applyFinalTask(msg.task)
		m.waiting = false
		m.refreshViewport()
		return m, nil

	case agentErrorMsg:
		m.addLine(senderSystem, "⚠ "+msg.err.Error())
		m.waiting = false
		m.replyIdx = -1
		m.refreshViewport()
		return m, nil

	case spinner.TickMsg:
		if !m.waiting {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m interactiveModel) submit(text string) (tea.Model, tea.Cmd) {
	m.addLine(senderUser, text)
	m.input.Reset()
	m.waiting = true
	m.refreshViewport()

	params := m.buildParams(text)

	if m.mode == modeBackground {
		return m, tea.Batch(m.spinner.Tick, submitBackgroundCmd(params))
	}

	m.agentBuf = ""
	m.replyIdx = -1
	return m, tea.Batch(m.spinner.Tick, startStreamCmd(params))
}

func (m interactiveModel) buildParams(text string) adk.MessageSendParams {
	contextID := m.contextID
	params := adk.MessageSendParams{
		Message: adk.Message{
			MessageID: uuid.NewString(),
			Role:      adk.RoleUser,
			ContextID: &contextID,
			Parts:     []adk.Part{{Text: &text}},
		},
	}
	// Continue an in-progress task when the agent previously asked for more input.
	if m.lastTaskID != "" && m.lastState == adk.TaskStateInputRequired {
		taskID := m.lastTaskID
		params.Message.TaskID = &taskID
	}
	return params
}

func (m *interactiveModel) applyStreamEvent(resp adk.JSONRPCSuccessResponse) {
	eventJSON, err := json.Marshal(resp.Result)
	if err != nil {
		return
	}

	var generic map[string]any
	if err := json.Unmarshal(eventJSON, &generic); err != nil {
		return
	}

	_, hasArtifact := generic["artifact"]
	_, hasFinal := generic["final"]
	_, hasID := generic["id"]

	switch {
	case hasArtifact:
		var ev adk.TaskArtifactUpdateEvent
		if err := json.Unmarshal(eventJSON, &ev); err != nil {
			return
		}
		if ev.TaskID != "" {
			m.lastTaskID = ev.TaskID
		}
		if text := partsToText(ev.Artifact.Parts); text != "" {
			m.appendAgentText(text)
		}
	case hasFinal:
		var ev adk.TaskStatusUpdateEvent
		if err := json.Unmarshal(eventJSON, &ev); err != nil {
			return
		}
		if ev.TaskID != "" {
			m.lastTaskID = ev.TaskID
		}
		m.lastState = ev.Status.State
		if ev.Status.Message != nil {
			if text := partsToText(ev.Status.Message.Parts); text != "" {
				m.appendAgentText(text)
			}
		}
	case hasID:
		var task adk.Task
		if err := json.Unmarshal(eventJSON, &task); err != nil {
			return
		}
		if task.ID != "" {
			m.lastTaskID = task.ID
		}
		m.lastState = task.Status.State
		if task.Status.Message != nil {
			if text := partsToText(task.Status.Message.Parts); text != "" {
				m.appendAgentText(text)
			}
		}
	}
}

func (m *interactiveModel) appendAgentText(text string) {
	m.agentBuf += text
	if m.replyIdx >= 0 && m.replyIdx < len(m.lines) {
		m.lines[m.replyIdx].text = m.agentBuf
		return
	}
	m.lines = append(m.lines, chatLine{sender: senderAgent, text: m.agentBuf})
	m.replyIdx = len(m.lines) - 1
}

func (m *interactiveModel) finishAgentReply() {
	if m.replyIdx == -1 {
		m.addLine(senderSystem, "(no response received)")
	}
	if m.lastState == adk.TaskStateInputRequired {
		m.addLine(senderSystem, "agent needs more input — type your reply")
	}
	m.waiting = false
	m.replyIdx = -1
	m.agentBuf = ""
}

func (m *interactiveModel) applyFinalTask(task adk.Task) {
	if task.ID != "" {
		m.lastTaskID = task.ID
	}
	m.lastState = task.Status.State

	text := ""
	if task.Status.Message != nil {
		text = partsToText(task.Status.Message.Parts)
	}
	if text == "" {
		text = latestAgentText(task.History)
	}

	switch {
	case text != "":
		m.addLine(senderAgent, text)
	case task.Status.State == adk.TaskStateCompleted:
		m.addLine(senderSystem, "task completed with no message")
	default:
		m.addLine(senderSystem, "task ended: "+humanState(task.Status.State))
	}

	if task.Status.State == adk.TaskStateInputRequired {
		m.addLine(senderSystem, "agent needs more input — type your reply")
	}
}

func (m *interactiveModel) toggleMode() {
	if m.mode == modeStreaming {
		m.mode = modeBackground
	} else {
		m.mode = modeStreaming
	}
	m.addLine(senderSystem, "switched to "+m.mode.String()+" mode")
}

func (m *interactiveModel) addLine(sender msgSender, text string) {
	m.lines = append(m.lines, chatLine{sender: sender, text: text})
}

func (m *interactiveModel) layout() {
	vpHeight := m.height - lipgloss.Height(m.headerView()) - lipgloss.Height(m.footerView())
	if vpHeight < 3 {
		vpHeight = 3
	}
	if m.ready {
		m.viewport.Width = m.width
		m.viewport.Height = vpHeight
	} else {
		m.viewport = viewport.New(m.width, vpHeight)
	}
	inputWidth := m.width - 4
	if inputWidth < 10 {
		inputWidth = 10
	}
	m.input.Width = inputWidth
}

func (m *interactiveModel) refreshViewport() {
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.renderConversation())
	m.viewport.GotoBottom()
}

func (m *interactiveModel) renderConversation() string {
	width := m.viewport.Width - 2
	if width < 20 {
		width = 20
	}

	var b strings.Builder
	for i, line := range m.lines {
		if i > 0 {
			b.WriteString("\n")
		}
		switch line.sender {
		case senderUser:
			b.WriteString(userLabelStyle.Render("You"))
			b.WriteString("\n")
			b.WriteString(bodyStyle.Width(width).Render(line.text))
		case senderAgent:
			b.WriteString(agentLabelStyle.Render(m.agentName))
			b.WriteString("\n")
			b.WriteString(bodyStyle.Width(width).Render(line.text))
		default:
			b.WriteString(systemStyle.Width(width).Render(line.text))
		}
	}
	return b.String()
}

func (m interactiveModel) headerView() string {
	title := titleStyle.Render("A2A Chat")
	meta := metaStyle.Render(fmt.Sprintf("%s · %s · context %s", m.serverURL, m.mode.String(), shortID(m.contextID)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, " ", meta)
}

func (m interactiveModel) footerView() string {
	var status string
	if m.waiting {
		status = m.spinner.View() + " " + dimStyle.Render("waiting for "+m.agentName+"...")
	} else {
		status = dimStyle.Render("ready")
	}
	help := dimStyle.Render("enter: send · ctrl+t: toggle mode · ctrl+c: quit")
	return strings.Join([]string{status, m.input.View(), help}, "\n")
}

func (m interactiveModel) View() string {
	if !m.ready {
		return "initializing..."
	}
	return strings.Join([]string{m.headerView(), m.viewport.View(), m.footerView()}, "\n")
}

// --- commands ---

func startStreamCmd(params adk.MessageSendParams) tea.Cmd {
	return func() tea.Msg {
		ch, err := a2aClient.SendTaskStreaming(context.Background(), params)
		if err != nil {
			return agentErrorMsg{err: handleA2AError(err, "message/stream")}
		}
		return streamStartedMsg{ch: ch}
	}
}

func readStreamCmd(ch <-chan adk.JSONRPCSuccessResponse) tea.Cmd {
	return func() tea.Msg {
		resp, ok := <-ch
		return streamEventMsg{resp: resp, ok: ok, ch: ch}
	}
}

func submitBackgroundCmd(params adk.MessageSendParams) tea.Cmd {
	return func() tea.Msg {
		resp, err := a2aClient.SendTask(context.Background(), params)
		if err != nil {
			return agentErrorMsg{err: handleA2AError(err, "message/send")}
		}
		task, err := taskFromResult(resp.Result)
		if err != nil {
			return agentErrorMsg{err: err}
		}
		return taskSubmittedMsg{taskID: task.ID, contextID: task.ContextID}
	}
}

func pollTaskCmd(taskID string) tea.Cmd {
	return tea.Tick(backgroundPollInterval, func(time.Time) tea.Msg {
		resp, err := a2aClient.GetTask(context.Background(), adk.TaskQueryParams{ID: taskID})
		if err != nil {
			return agentErrorMsg{err: handleA2AError(err, "tasks/get")}
		}
		task, err := taskFromResult(resp.Result)
		if err != nil {
			return agentErrorMsg{err: err}
		}
		return taskPolledMsg{task: task, done: isTerminalState(task.Status.State)}
	})
}

// --- helpers ---

func taskFromResult(result any) (adk.Task, error) {
	var task adk.Task
	b, err := json.Marshal(result)
	if err != nil {
		return task, fmt.Errorf("failed to marshal task result: %w", err)
	}
	if err := json.Unmarshal(b, &task); err != nil {
		return task, fmt.Errorf("failed to unmarshal task: %w", err)
	}
	return task, nil
}

func isTerminalState(state adk.TaskState) bool {
	switch state {
	case adk.TaskStateCompleted,
		adk.TaskStateFailed,
		adk.TaskStateCancelled,
		adk.TaskStateRejected,
		adk.TaskStateInputRequired:
		return true
	default:
		return false
	}
}

func partsToText(parts []adk.Part) string {
	var b strings.Builder
	for _, p := range parts {
		if p.Text != nil {
			b.WriteString(*p.Text)
		}
	}
	return b.String()
}

func latestAgentText(history []adk.Message) string {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == adk.RoleAgent {
			if text := partsToText(history[i].Parts); text != "" {
				return text
			}
		}
	}
	return ""
}

func humanState(s adk.TaskState) string {
	return strings.ToLower(strings.TrimPrefix(string(s), "TASK_STATE_"))
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// runInteractiveChat boots the Bubble Tea program for the chat interface.
func runInteractiveChat(mode chatMode, contextID string) error {
	ensureA2AClient()

	agentName := "Agent"
	streamingSupported := true
	if card, err := a2aClient.GetAgentCard(context.Background()); err == nil && card != nil {
		if card.Name != "" {
			agentName = card.Name
		}
		if card.Capabilities.Streaming != nil {
			streamingSupported = *card.Capabilities.Streaming
		}
	}

	model := newInteractiveModel(mode, viper.GetString("server-url"), agentName, contextID)
	if mode == modeStreaming && !streamingSupported {
		model.addLine(senderSystem, "⚠ agent does not advertise streaming support; responses may not stream")
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
