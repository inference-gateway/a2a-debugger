package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	a2a "github.com/inference-gateway/a2a-debugger/a2a"
	client "github.com/inference-gateway/adk/client"
	adk "github.com/inference-gateway/adk/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// mockA2AClient implements the A2AClient interface for testing
type mockA2AClient struct {
	sendTaskStreamingFunc func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- interface{}) error
}

func (m *mockA2AClient) GetAgentCard(ctx context.Context) (*adk.AgentCard, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) GetHealth(ctx context.Context) (*client.HealthResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) ListTasks(ctx context.Context, params adk.TaskListParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) GetTask(ctx context.Context, params adk.TaskQueryParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SendTask(ctx context.Context, params adk.MessageSendParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SendTaskStreaming(ctx context.Context, params adk.MessageSendParams, eventChan chan<- interface{}) error {
	if m.sendTaskStreamingFunc != nil {
		return m.sendTaskStreamingFunc(ctx, params, eventChan)
	}
	return fmt.Errorf("not implemented")
}

func (m *mockA2AClient) CancelTask(ctx context.Context, params adk.TaskIdParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SetTimeout(timeout time.Duration) {}

func (m *mockA2AClient) SetHTTPClient(client *http.Client) {}

func (m *mockA2AClient) GetBaseURL() string {
	return "http://localhost:8080"
}

func (m *mockA2AClient) SetLogger(logger *zap.Logger) {}

func (m *mockA2AClient) GetLogger() *zap.Logger {
	return zap.NewNop()
}

func TestSubmitStreamingTaskCmd_StreamingSummary(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	mockClient := &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- interface{}) error {
			statusEvent := a2a.TaskStatusUpdateEvent{
				Kind:      "status-update",
				TaskID:    "test-task-123",
				ContextID: "test-context-456",
				Status: a2a.TaskStatus{
					State: a2a.TaskStateWorking,
					Message: &a2a.Message{
						Kind:      "message",
						MessageID: "msg-123",
						Role:      "assistant",
						Parts: []a2a.Part{
							map[string]interface{}{
								"kind": "text",
								"text": "Test response",
							},
						},
					},
				},
				Final: false,
			}
			eventChan <- statusEvent

			artifactEvent := a2a.TaskArtifactUpdateEvent{
				Kind:      "artifact-update",
				TaskID:    "test-task-123",
				ContextID: "test-context-456",
				Artifact: a2a.Artifact{
					ArtifactID: "artifact-123",
					Parts: []a2a.Part{
						map[string]interface{}{
							"kind": "text",
							"text": "Test artifact content",
						},
					},
				},
			}
			eventChan <- artifactEvent

			finalStatusEvent := a2a.TaskStatusUpdateEvent{
				Kind:      "status-update",
				TaskID:    "test-task-123",
				ContextID: "test-context-456",
				Status: a2a.TaskStatus{
					State: a2a.TaskStateCompleted,
					Message: &a2a.Message{
						Kind:      "message",
						MessageID: "msg-124",
						Role:      "assistant",
						Parts: []a2a.Part{
							map[string]interface{}{
								"kind": "text",
								"text": "Task completed",
							},
						},
					},
				},
				Final: true,
			}
			eventChan <- finalStatusEvent
			
			return nil
		},
	}
	a2aClient = mockClient

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().String("context-id", "", "Context ID for the task")
	cmd.Flags().Bool("raw", false, "Show raw streaming event data")
	
	err := submitStreamingTaskCmd.RunE(cmd, []string{"test message"})

	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	a2aClient = originalClient
	logger = originalLogger

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedParts := []string{
		"Streaming Summary:",
		"Task ID: test-task-123",
		"Context ID: test-context-456",
		"Final Status: completed",
		"Duration:",
		"Total Events: 3",
		"Status Updates: 2",
		"Artifact Updates: 1",
		"Final Message Parts: 1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", part, output)
		}
	}
}

func TestSubmitStreamingTaskCmd_RawMode(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	mockClient := &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- interface{}) error {
			statusEvent := a2a.TaskStatusUpdateEvent{
				Kind:      "status-update",
				TaskID:    "test-task-456",
				ContextID: "test-context-789",
				Status: a2a.TaskStatus{
					State: a2a.TaskStateCompleted,
				},
				Final: true,
			}
			eventChan <- statusEvent
			
			return nil
		},
	}
	a2aClient = mockClient

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().String("context-id", "", "Context ID for the task")
	cmd.Flags().Bool("raw", true, "Show raw streaming event data")
	_ = cmd.Flag("raw").Value.Set("true")
	
	err := submitStreamingTaskCmd.RunE(cmd, []string{"test message"})

	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	a2aClient = originalClient
	logger = originalLogger

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedParts := []string{
		"Raw Event:",
		"Streaming Summary:",
		"Task ID: test-task-456",
		"Context ID: test-context-789",
		"Final Status: completed",
		"Total Events: 1",
		"Status Updates: 1",
		"Artifact Updates: 0",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", part, output)
		}
	}
}
