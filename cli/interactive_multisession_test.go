package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	adk "github.com/inference-gateway/adk/types"
)

// streamEvent builds a final status-update event like the mock-agent emits.
func streamEvent(taskID, contextID, text string) adk.JSONRPCSuccessResponse {
	return adk.JSONRPCSuccessResponse{
		Result: map[string]any{
			"taskId":    taskID,
			"contextId": contextID,
			"final":     true,
			"status": map[string]any{
				"state": string(adk.TaskStateCompleted),
				"message": map[string]any{
					"messageId": "m-" + taskID,
					"role":      string(adk.RoleAgent),
					"parts":     []map[string]any{{"text": text}},
				},
			},
		},
	}
}

// otherSessionID returns the one session id in m that is not exclude.
func otherSessionID(m interactiveModel, exclude string) string {
	for id := range m.sessions {
		if id != exclude {
			return id
		}
	}
	return ""
}

func TestMultiSession_NoStateBleed(t *testing.T) {
	m := newInteractiveModel(modeStreaming, "http://mock:8080", "MockAgent", "ctx-A")
	// Make it ready so refreshViewport is a no-op-safe path via WindowSize.
	mi, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mi.(interactiveModel)

	// Session A: user turn + agent reply from "mock-agent".
	m.addLine(senderUser, "hello from A")
	m.applyStreamEvent(streamEvent("task-A", "ctx-A", "reply in A"))
	m.finishAgentReply()

	if m.lastTaskID != "task-A" {
		t.Fatalf("session A lastTaskID = %q, want task-A", m.lastTaskID)
	}
	linesA := len(m.lines)
	if linesA == 0 {
		t.Fatal("session A should have transcript lines")
	}

	// /new -> creates + switches to a fresh session, saving A.
	mi, _ = m.handleSlashCommand("/new")
	m = mi.(interactiveModel)
	newID := m.activeSession
	if newID == "ctx-A" {
		t.Fatal("/new did not switch away from ctx-A")
	}
	if len(m.sessions) != 2 {
		t.Fatalf("want 2 sessions, got %d", len(m.sessions))
	}

	// Fresh session must not inherit A's state.
	if m.lastTaskID != "" {
		t.Errorf("new session leaked lastTaskID %q from A", m.lastTaskID)
	}
	if m.contextID == "ctx-A" {
		t.Error("new session leaked contextID from A")
	}
	// Only the system "new session created" line should exist, no A transcript.
	for _, l := range m.lines {
		if l.sender != senderSystem {
			t.Errorf("new session leaked non-system line from A: %q", l.text)
		}
	}

	// Session B: its own turn with a distinct task id.
	m.addLine(senderUser, "hello from B")
	m.applyStreamEvent(streamEvent("task-B", newID, "reply in B"))
	m.finishAgentReply()
	if m.lastTaskID != "task-B" {
		t.Fatalf("session B lastTaskID = %q, want task-B", m.lastTaskID)
	}

	// Switch back to A: its state must be restored intact.
	mi, _ = m.handleSlashCommand("/session ctx-A")
	m = mi.(interactiveModel)
	if m.activeSession != "ctx-A" {
		t.Fatalf("did not switch back to ctx-A, active=%s", m.activeSession)
	}
	if m.lastTaskID != "task-A" {
		t.Errorf("A restored lastTaskID = %q, want task-A (B bled through?)", m.lastTaskID)
	}
	if m.contextID != "ctx-A" {
		t.Errorf("A restored contextID = %q, want ctx-A", m.contextID)
	}
	// Switching back adds one system line ("switched to..."); A's own lines remain.
	foundReplyA := false
	for _, l := range m.lines {
		if l.sender == senderAgent && l.text == "reply in A" {
			foundReplyA = true
		}
		if l.text == "reply in B" {
			t.Error("session B transcript bled into A")
		}
	}
	if !foundReplyA {
		t.Error("session A transcript was lost after round trip")
	}

	// Switch to B again: B's state intact, A not bleeding.
	mi, _ = m.handleSlashCommand("/session " + otherSessionID(m, "ctx-A"))
	m = mi.(interactiveModel)
	if m.lastTaskID != "task-B" {
		t.Errorf("B restored lastTaskID = %q, want task-B", m.lastTaskID)
	}
	for _, l := range m.lines {
		if l.text == "reply in A" {
			t.Error("session A transcript bled into B")
		}
	}
}
