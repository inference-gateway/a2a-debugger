package cli

import (
	"context"
	"strings"
	"testing"

	bubbletea "github.com/charmbracelet/bubbletea"

	types "github.com/inference-gateway/adk/types"
)

func lastAgentLine(m interactiveModel) (string, bool) {
	for i := len(m.lines) - 1; i >= 0; i-- {
		if m.lines[i].sender == senderAgent {
			return m.lines[i].text, true
		}
	}
	return "", false
}

func hasSystemLine(m interactiveModel, substr string) bool {
	for _, l := range m.lines {
		if l.sender == senderSystem && strings.Contains(l.text, substr) {
			return true
		}
	}
	return false
}

func statusEventResp(text string, state types.TaskState, final bool) types.JSONRPCSuccessResponse {
	status := map[string]any{"state": string(state)}
	if text != "" {
		status["message"] = map[string]any{
			"messageId": "m-agent",
			"role":      string(types.RoleAgent),
			"parts":     []map[string]any{{"text": text}},
		}
	}
	return types.JSONRPCSuccessResponse{
		Result: map[string]any{
			"taskId":    "task-1",
			"contextId": "ctx-1",
			"final":     final,
			"status":    status,
		},
	}
}

func TestNewInteractiveModel(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "http://localhost:8080", "TestAgent", "")
	if m.mode != modeStreaming {
		t.Errorf("expected streaming mode, got %v", m.mode)
	}
	if m.contextID == "" {
		t.Error("expected a generated context ID")
	}
	if m.replyIdx != -1 {
		t.Errorf("expected replyIdx -1, got %d", m.replyIdx)
	}
	if m.agentName != "TestAgent" {
		t.Errorf("expected agent name TestAgent, got %q", m.agentName)
	}

	resumed := newInteractiveModel(modeBackground, "url", "", "ctx-resume")
	if resumed.contextID != "ctx-resume" {
		t.Errorf("expected resumed context ctx-resume, got %q", resumed.contextID)
	}
	if resumed.agentName != "Agent" {
		t.Errorf("expected default agent name Agent, got %q", resumed.agentName)
	}
}

func TestInteractiveStreamingAccumulatesText(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.waiting = true

	updated, _ := m.Update(streamEventMsg{ok: true, resp: statusEventResp("Hello ", types.TaskStateWorking, false)})
	m = updated.(interactiveModel)
	updated, _ = m.Update(streamEventMsg{ok: true, resp: statusEventResp("world", types.TaskStateCompleted, true)})
	m = updated.(interactiveModel)
	updated, _ = m.Update(streamEventMsg{ok: false})
	m = updated.(interactiveModel)

	if m.waiting {
		t.Error("expected waiting=false after stream completion")
	}
	got, ok := lastAgentLine(m)
	if !ok {
		t.Fatal("expected an agent line")
	}
	if got != "Hello world" {
		t.Errorf("expected accumulated reply 'Hello world', got %q", got)
	}
}

func TestInteractiveStreamingNoResponse(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.waiting = true

	updated, _ := m.Update(streamEventMsg{ok: false})
	m = updated.(interactiveModel)

	if !hasSystemLine(m, "no response received") {
		t.Error("expected a 'no response received' system line")
	}
}

func TestInteractiveStreamingInputRequiredPrompt(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.waiting = true

	updated, _ := m.Update(streamEventMsg{ok: true, resp: statusEventResp("need more info", types.TaskStateInputRequired, true)})
	m = updated.(interactiveModel)
	updated, _ = m.Update(streamEventMsg{ok: false})
	m = updated.(interactiveModel)

	if m.lastState != types.TaskStateInputRequired {
		t.Errorf("expected lastState input-required, got %v", m.lastState)
	}
	if !hasSystemLine(m, "needs more input") {
		t.Error("expected a system prompt requesting more input")
	}
}

func TestInteractiveBackgroundFlow(t *testing.T) {
	m := newInteractiveModel(modeBackground, "url", "Agent", "ctx-1")
	m.waiting = true

	updated, cmd := m.Update(taskSubmittedMsg{taskID: "task-9", contextID: "ctx-9"})
	m = updated.(interactiveModel)
	if cmd == nil {
		t.Error("expected a poll command after task submission")
	}
	if m.contextID != "ctx-9" {
		t.Errorf("expected context ID to be updated to ctx-9, got %q", m.contextID)
	}
	if m.lastTaskID != "task-9" {
		t.Errorf("expected lastTaskID task-9, got %q", m.lastTaskID)
	}

	workingTask := types.Task{ID: "task-9", Status: types.TaskStatus{State: types.TaskStateWorking}}
	updated, cmd = m.Update(taskPolledMsg{task: workingTask, done: false})
	m = updated.(interactiveModel)
	if cmd == nil {
		t.Error("expected another poll command while task is working")
	}
	if !m.waiting {
		t.Error("expected to still be waiting while task is working")
	}

	finalText := "final answer"
	doneTask := types.Task{
		ID: "task-9",
		Status: types.TaskStatus{
			State: types.TaskStateCompleted,
			Message: &types.Message{
				Role:  types.RoleAgent,
				Parts: []types.Part{{Text: &finalText}},
			},
		},
	}
	updated, _ = m.Update(taskPolledMsg{task: doneTask, done: true})
	m = updated.(interactiveModel)

	if m.waiting {
		t.Error("expected waiting=false after task completion")
	}
	got, ok := lastAgentLine(m)
	if !ok {
		t.Fatal("expected an agent line for the completed task")
	}
	if got != finalText {
		t.Errorf("expected agent line %q, got %q", finalText, got)
	}
}

func TestInteractiveBackgroundUsesHistoryFallback(t *testing.T) {
	m := newInteractiveModel(modeBackground, "url", "Agent", "ctx-1")
	m.waiting = true

	historyText := "answer from history"
	doneTask := types.Task{
		ID:     "task-h",
		Status: types.TaskStatus{State: types.TaskStateCompleted},
		History: []types.Message{
			{Role: types.RoleUser, Parts: []types.Part{}},
			{Role: types.RoleAgent, Parts: []types.Part{{Text: &historyText}}},
		},
	}
	updated, _ := m.Update(taskPolledMsg{task: doneTask, done: true})
	m = updated.(interactiveModel)

	got, ok := lastAgentLine(m)
	if !ok || got != historyText {
		t.Errorf("expected fallback to history text %q, got %q (ok=%v)", historyText, got, ok)
	}
}

func TestInteractiveToggleMode(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	updated, _ := m.Update(bubbletea.KeyMsg{Type: bubbletea.KeyCtrlT})
	m = updated.(interactiveModel)
	if m.mode != modeBackground {
		t.Errorf("expected background mode after toggle, got %v", m.mode)
	}
	if !hasSystemLine(m, "switched to background mode") {
		t.Error("expected a system line announcing the mode switch")
	}
}

func TestInteractiveEnterSubmitsMessage(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.input.SetValue("hi there")

	updated, cmd := m.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
	m = updated.(interactiveModel)

	if !m.waiting {
		t.Error("expected waiting=true after submitting a message")
	}
	if cmd == nil {
		t.Error("expected a command batch after submitting a message")
	}
	if m.input.Value() != "" {
		t.Error("expected input to be reset after submission")
	}

	var foundUser bool
	for _, l := range m.lines {
		if l.sender == senderUser && l.text == "hi there" {
			foundUser = true
		}
	}
	if !foundUser {
		t.Error("expected the submitted message to appear as a user line")
	}
}

func TestInteractiveEnterIgnoredWhileWaiting(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.waiting = true
	m.input.SetValue("should be ignored")

	updated, _ := m.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
	m = updated.(interactiveModel)

	for _, l := range m.lines {
		if l.sender == senderUser {
			t.Error("expected no user line to be added while waiting")
		}
	}
}

func TestBuildParamsContinuesInputRequiredTask(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "url", "Agent", "ctx-1")
	m.lastTaskID = "task-7"
	m.lastState = types.TaskStateInputRequired

	params := m.buildParams("continue please")
	if params.Message.TaskID == nil || *params.Message.TaskID != "task-7" {
		t.Error("expected task ID to be set when continuing an input-required task")
	}
	if params.Message.ContextID == nil || *params.Message.ContextID != "ctx-1" {
		t.Error("expected context ID to always be set")
	}

	m.lastState = types.TaskStateCompleted
	params = m.buildParams("new message")
	if params.Message.TaskID != nil {
		t.Error("expected task ID to be unset for a completed prior task")
	}
}

func TestStartStreamCmd(t *testing.T) {
	originalClient := a2aClient
	defer func() { a2aClient = originalClient }()

	ch := make(chan types.JSONRPCSuccessResponse)
	close(ch)
	a2aClient = &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params types.MessageSendParams) (<-chan types.JSONRPCSuccessResponse, error) {
			return ch, nil
		},
	}

	msg := startStreamCmd(types.MessageSendParams{})()
	if _, ok := msg.(streamStartedMsg); !ok {
		t.Fatalf("expected streamStartedMsg, got %T", msg)
	}
}

func TestStartStreamCmdMethodNotFound(t *testing.T) {
	originalClient := a2aClient
	defer func() { a2aClient = originalClient }()

	a2aClient = &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params types.MessageSendParams) (<-chan types.JSONRPCSuccessResponse, error) {
			return nil, &mockError{msg: "MethodNotFoundError: -32601"}
		},
	}

	msg := startStreamCmd(types.MessageSendParams{})()
	errMsg, ok := msg.(agentErrorMsg)
	if !ok {
		t.Fatalf("expected agentErrorMsg, got %T", msg)
	}
	if !strings.Contains(errMsg.err.Error(), "not implemented by the agent") {
		t.Errorf("expected friendly method-not-found message, got %q", errMsg.err.Error())
	}
}

func TestSubmitBackgroundCmd(t *testing.T) {
	originalClient := a2aClient
	defer func() { a2aClient = originalClient }()

	a2aClient = &mockA2AClient{
		sendTaskFunc: func(ctx context.Context, params types.MessageSendParams) (*types.JSONRPCSuccessResponse, error) {
			return &types.JSONRPCSuccessResponse{
				Result: map[string]any{
					"id":        "task-bg",
					"contextId": "ctx-bg",
					"status":    map[string]any{"state": string(types.TaskStateSubmitted)},
				},
			}, nil
		},
	}

	msg := submitBackgroundCmd(types.MessageSendParams{})()
	sub, ok := msg.(taskSubmittedMsg)
	if !ok {
		t.Fatalf("expected taskSubmittedMsg, got %T", msg)
	}
	if sub.taskID != "task-bg" {
		t.Errorf("expected task ID task-bg, got %q", sub.taskID)
	}
	if sub.contextID != "ctx-bg" {
		t.Errorf("expected context ID ctx-bg, got %q", sub.contextID)
	}
}

func TestIsTerminalState(t *testing.T) {
	terminal := []types.TaskState{
		types.TaskStateCompleted,
		types.TaskStateFailed,
		types.TaskStateCancelled,
		types.TaskStateRejected,
		types.TaskStateInputRequired,
	}
	for _, s := range terminal {
		if !isTerminalState(s) {
			t.Errorf("expected %s to be terminal", s)
		}
	}

	nonTerminal := []types.TaskState{
		types.TaskStateWorking,
		types.TaskStateSubmitted,
		types.TaskStateUnspecified,
	}
	for _, s := range nonTerminal {
		if isTerminalState(s) {
			t.Errorf("expected %s to be non-terminal", s)
		}
	}
}

func TestPartsToText(t *testing.T) {
	a := "foo"
	b := "bar"
	parts := []types.Part{{Text: &a}, {Data: nil}, {Text: &b}}
	if got := partsToText(parts); got != "foobar" {
		t.Errorf("expected 'foobar', got %q", got)
	}
}

func TestHumanState(t *testing.T) {
	if got := humanState(types.TaskStateInputRequired); got != "input_required" {
		t.Errorf("expected 'input_required', got %q", got)
	}
}

func TestShortID(t *testing.T) {
	if got := shortID("1234567890abcdef"); got != "12345678" {
		t.Errorf("expected first 8 chars, got %q", got)
	}
	if got := shortID("abc"); got != "abc" {
		t.Errorf("expected short id unchanged, got %q", got)
	}
}

func TestChatModeString(t *testing.T) {
	if modeStreaming.String() != "streaming" {
		t.Errorf("expected 'streaming', got %q", modeStreaming.String())
	}
	if modeBackground.String() != "background" {
		t.Errorf("expected 'background', got %q", modeBackground.String())
	}
}

// mockError is a simple error type for exercising handleA2AError paths.
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
