package cli

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	require "github.com/stretchr/testify/require"

	sdk "github.com/cloudevents/sdk-go/v2"
	uuid "github.com/google/uuid"
	zap "go.uber.org/zap"

	server "github.com/inference-gateway/adk/server"
	config "github.com/inference-gateway/adk/server/config"
	types "github.com/inference-gateway/adk/types"
)

// echoAgent streams three deltas and an iteration-completed event, enough to
// exercise every branch of the StreamResponse dispatch without any LLM.
type echoAgent struct{}

func (echoAgent) RunWithStream(ctx context.Context, messages []types.Message) (<-chan sdk.Event, error) {
	taskID, contextID := "", ""
	if task, ok := ctx.Value(server.TaskContextKey).(*types.Task); ok && task != nil {
		taskID, contextID = task.ID, task.GetContextID()
	}

	events := make(chan sdk.Event, 8)
	go func() {
		defer close(events)

		status := sdk.NewEvent()
		status.SetType(types.EventTaskStatusChanged)
		_ = status.SetData(sdk.ApplicationJSON, types.TaskStatus{State: types.TaskStateWorking})
		events <- status

		var full string
		for _, word := range []string{"streaming", " ", "pong"} {
			full += word
			delta := sdk.NewEvent()
			delta.SetType(types.EventDelta)
			_ = delta.SetData(sdk.ApplicationJSON, types.Message{
				Role:  types.RoleAgent,
				Parts: []types.Part{types.CreateTextPart(word)},
			})
			events <- delta
		}

		final := types.Message{
			MessageID: uuid.NewString(),
			Role:      types.RoleAgent,
			TaskID:    &taskID,
			ContextID: &contextID,
			Parts:     []types.Part{types.CreateTextPart(full)},
		}
		events <- types.NewIterationCompletedEvent(1, taskID, &final)
	}()

	return events, nil
}

// echoHandler answers background tasks by echoing the prompt back.
type echoHandler struct{ agent server.OpenAICompatibleAgent }

func (h *echoHandler) GetAgent() server.OpenAICompatibleAgent  { return h.agent }
func (h *echoHandler) SetAgent(a server.OpenAICompatibleAgent) { h.agent = a }
func (h *echoHandler) HandleTask(ctx context.Context, task *types.Task, message *types.Message) (*types.Task, error) {
	prompt := ""
	if message != nil {
		for _, part := range message.Parts {
			if part.Text != nil {
				prompt = *part.Text
			}
		}
	}

	reply := types.Message{
		MessageID: uuid.NewString(),
		ContextID: task.ContextID,
		TaskID:    &task.ID,
		Role:      types.RoleAgent,
		Parts:     []types.Part{types.CreateTextPart("Echo: " + prompt)},
	}
	task.History = append(task.History, reply)
	task.Status.State = types.TaskStateCompleted
	task.Status.Message = &reply

	return task, nil
}

// startAgent boots a real ADK A2A server on a free port and returns its URL.
func startAgent(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	require.NoError(t, listener.Close())

	streaming, push := true, false
	cfg := config.Config{
		AgentName:          "e2e-agent",
		AgentDescription:   "echo agent for the debugger end-to-end test",
		AgentVersion:       "0.0.0",
		CapabilitiesConfig: config.CapabilitiesConfig{Streaming: streaming},
		QueueConfig:        config.QueueConfig{CleanupInterval: time.Minute},
		ServerConfig:       config.ServerConfig{Port: port},
	}

	agent := echoAgent{}
	a2aServer, err := server.NewA2AServerBuilder(cfg, zap.NewNop()).
		WithAgent(agent).
		WithBackgroundTaskHandler(&echoHandler{agent: agent}).
		WithDefaultStreamingTaskHandler().
		WithAgentCard(types.AgentCard{
			Name:                cfg.AgentName,
			Description:         cfg.AgentDescription,
			Version:             cfg.AgentVersion,
			SupportedInterfaces: []types.AgentInterface{{URL: "http://127.0.0.1:" + port, ProtocolBinding: "JSONRPC", ProtocolVersion: "1.0"}},
			Capabilities:        types.AgentCapabilities{Streaming: &streaming, PushNotifications: &push},
			DefaultInputModes:   []string{"text/plain"},
			DefaultOutputModes:  []string{"text/plain"},
			Skills:              []types.AgentSkill{},
		}).
		Build()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = a2aServer.Start(ctx) }()
	t.Cleanup(func() {
		cancel()
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopCancel()
		_ = a2aServer.Stop(stopCtx)
	})

	url := "http://127.0.0.1:" + port
	require.Eventually(t, func() bool {
		resp, err := http.Get(url + "/.well-known/agent-card.json")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 50*time.Millisecond, "agent never became ready")

	return url
}

// runCLI executes the root command with args and returns everything it printed.
// ponytail: the pipe is drained only after the command returns, so this works
// for command output below the pipe buffer; stream it if an assertion ever needs more.
func runCLI(t *testing.T, args ...string) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	rootCmd.SetArgs(args)
	runErr := rootCmd.Execute()

	require.NoError(t, w.Close())
	os.Stdout = original
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	require.NoError(t, runErr, "command %v failed, output: %s", args, out)
	return string(out)
}

// TestEndToEndAgainstLiveAgent walks the command set the example README documents
// against a real ADK server, so a breaking change in the ADK surface fails here
// rather than only when someone runs example/docker-compose.yml.
func TestEndToEndAgainstLiveAgent(t *testing.T) {
	url := startAgent(t)

	original := a2aClient
	t.Cleanup(func() { a2aClient = original })
	a2aClient = nil

	connect := runCLI(t, "connect", "--server-url", url, "--output", "json")
	require.Contains(t, connect, `"connected": true`)
	require.Contains(t, connect, "e2e-agent")

	card := runCLI(t, "agent-card", "--server-url", url, "--output", "json")
	require.Contains(t, card, "e2e-agent")

	submit := runCLI(t, "tasks", "submit", "ping", "--context-id", "e2e-ctx", "--output", "json")
	require.Contains(t, submit, "e2e-ctx")

	taskID := extractJSONString(t, submit, "id")

	var get string
	for deadline := time.Now().Add(10 * time.Second); ; {
		get = runCLI(t, "tasks", "get", taskID, "--output", "json")
		if strings.Contains(get, string(types.TaskStateCompleted)) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("task never completed, last output: %s", get)
		}
		time.Sleep(100 * time.Millisecond)
	}

	require.Contains(t, get, "Echo: ping")

	list := runCLI(t, "tasks", "list", "--output", "json", "--limit", "10")
	require.Contains(t, list, taskID)

	filtered := runCLI(t, "tasks", "list", "--output", "json", "--state", "completed", "--context-id", "e2e-ctx")
	require.Contains(t, filtered, taskID)

	history := runCLI(t, "tasks", "history", "e2e-ctx", "--output", "json")
	require.Contains(t, history, "Echo: ping")

	stream := runCLI(t, "tasks", "submit-streaming", "pong?", "--context-id", "e2e-stream", "--output", "json")
	require.Contains(t, stream, "Context ID: e2e-stream")
	require.Contains(t, stream, "Status Update: working", "states must be printed in their human form")
	require.Contains(t, stream, "User Message", "the task snapshot carries the prompt, not an agent response")
	require.Contains(t, stream, "Streaming Summary")
	require.NotContains(t, stream, "TASK_STATE_", "raw enum values must not leak into the progress output")

	raw := runCLI(t, "tasks", "submit-streaming", "pong?", "--raw", "--output", "json")
	require.Contains(t, raw, "Raw Event")
}

// extractJSONString pulls the value of a "key": "value" pair out of CLI JSON output.
func extractJSONString(t *testing.T, output, key string) string {
	t.Helper()

	_, after, found := strings.Cut(output, `"`+key+`": "`)
	require.True(t, found, "key %q missing from output: %s", key, output)
	value, _, found := strings.Cut(after, `"`)
	require.True(t, found, "unterminated value for %q in output: %s", key, output)

	return value
}
