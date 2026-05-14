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

	client "github.com/inference-gateway/adk/client"
	adk "github.com/inference-gateway/adk/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// mockA2AClient implements the A2AClient interface for testing
type mockA2AClient struct {
	sendTaskStreamingFunc func(ctx context.Context, params adk.MessageSendParams) (<-chan adk.JSONRPCSuccessResponse, error)
}

func (m *mockA2AClient) GetAgentCard(ctx context.Context) (*adk.AgentCard, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) GetAuthenticatedExtendedCard(ctx context.Context, params adk.GetAuthenticatedExtendedCardParams) (*adk.JSONRPCSuccessResponse, error) {
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

func (m *mockA2AClient) SendTaskStreaming(ctx context.Context, params adk.MessageSendParams) (<-chan adk.JSONRPCSuccessResponse, error) {
	if m.sendTaskStreamingFunc != nil {
		return m.sendTaskStreamingFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) CancelTask(ctx context.Context, params adk.TaskIdParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) ResubscribeTask(ctx context.Context, params adk.TaskResubscriptionParams) (<-chan adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SetTaskPushNotificationConfig(ctx context.Context, params adk.TaskPushNotificationConfig) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) GetTaskPushNotificationConfig(ctx context.Context, params adk.GetTaskPushNotificationConfigParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) ListTaskPushNotificationConfig(ctx context.Context, params adk.ListTaskPushNotificationConfigParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) DeleteTaskPushNotificationConfig(ctx context.Context, params adk.DeleteTaskPushNotificationConfigParams) (*adk.JSONRPCSuccessResponse, error) {
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

func (m *mockA2AClient) GetArtifactHelper() *client.ArtifactHelper {
	return client.NewArtifactHelper()
}

func TestSubmitStreamingTaskCmd_StreamingSummary(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	mockClient := &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams) (<-chan adk.JSONRPCSuccessResponse, error) {
			ch := make(chan adk.JSONRPCSuccessResponse, 3)
			textResponse := "Test response"
			textCompleted := "Task completed"
			textArtifact := "Test artifact content"

			ch <- adk.JSONRPCSuccessResponse{
				Result: map[string]any{
					"taskId":    "test-task-123",
					"contextId": "test-context-456",
					"final":     false,
					"status": map[string]any{
						"state": string(adk.TaskStateWorking),
						"message": map[string]any{
							"messageId": "msg-123",
							"role":      string(adk.RoleAgent),
							"parts": []map[string]any{
								{"text": textResponse},
							},
						},
					},
				},
			}

			ch <- adk.JSONRPCSuccessResponse{
				Result: map[string]any{
					"taskId":    "test-task-123",
					"contextId": "test-context-456",
					"artifact": map[string]any{
						"artifactId": "artifact-123",
						"parts": []map[string]any{
							{"text": textArtifact},
						},
					},
				},
			}

			ch <- adk.JSONRPCSuccessResponse{
				Result: map[string]any{
					"taskId":    "test-task-123",
					"contextId": "test-context-456",
					"final":     true,
					"status": map[string]any{
						"state": string(adk.TaskStateCompleted),
						"message": map[string]any{
							"messageId": "msg-124",
							"role":      string(adk.RoleAgent),
							"parts": []map[string]any{
								{"text": textCompleted},
							},
						},
					},
				},
			}

			close(ch)
			return ch, nil
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
		"Final Status: " + string(adk.TaskStateCompleted),
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
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams) (<-chan adk.JSONRPCSuccessResponse, error) {
			ch := make(chan adk.JSONRPCSuccessResponse, 1)
			ch <- adk.JSONRPCSuccessResponse{
				Result: map[string]any{
					"taskId":    "test-task-456",
					"contextId": "test-context-789",
					"final":     true,
					"status": map[string]any{
						"state": string(adk.TaskStateCompleted),
					},
				},
			}
			close(ch)
			return ch, nil
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
		"Final Status: " + string(adk.TaskStateCompleted),
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

func TestSubmitStreamingTaskCmd_TaskSnapshot(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	mockClient := &mockA2AClient{
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams) (<-chan adk.JSONRPCSuccessResponse, error) {
			ch := make(chan adk.JSONRPCSuccessResponse, 1)
			finalText := "All done"
			ch <- adk.JSONRPCSuccessResponse{
				Result: map[string]any{
					"id":        "task-snapshot-1",
					"contextId": "ctx-snapshot-1",
					"history":   []map[string]any{},
					"status": map[string]any{
						"state": string(adk.TaskStateCompleted),
						"message": map[string]any{
							"messageId": "msg-final",
							"role":      string(adk.RoleAgent),
							"parts": []map[string]any{
								{"text": finalText},
							},
						},
					},
				},
			}
			close(ch)
			return ch, nil
		},
	}
	a2aClient = mockClient

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().String("context-id", "", "Context ID for the task")
	cmd.Flags().Bool("raw", false, "Show raw streaming event data")

	err := submitStreamingTaskCmd.RunE(cmd, []string{"snapshot"})

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
		"Task Snapshot: task-snapshot-1",
		"Streaming Summary:",
		"Task ID: task-snapshot-1",
		"Context ID: ctx-snapshot-1",
		"Final Status: " + string(adk.TaskStateCompleted),
		"Total Events: 1",
		"Status Updates: 0",
		"Artifact Updates: 0",
		"Final Message Parts: 1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", part, output)
		}
	}
}
