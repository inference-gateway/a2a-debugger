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
	sendTaskStreamingFunc func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- any) error
	getTaskFunc           func(ctx context.Context, params adk.TaskQueryParams) (*adk.JSONRPCSuccessResponse, error)
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
	if m.getTaskFunc != nil {
		return m.getTaskFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SendTask(ctx context.Context, params adk.MessageSendParams) (*adk.JSONRPCSuccessResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockA2AClient) SendTaskStreaming(ctx context.Context, params adk.MessageSendParams, eventChan chan<- any) error {
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
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- any) error {
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
							map[string]any{
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
						map[string]any{
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
							map[string]any{
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
		sendTaskStreamingFunc: func(ctx context.Context, params adk.MessageSendParams, eventChan chan<- any) error {
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

func TestDownloadArtifactsCmd_Success(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	// Create temporary directory for testing
	tempDir := t.TempDir()

	mockClient := &mockA2AClient{
		getTaskFunc: func(ctx context.Context, params adk.TaskQueryParams) (*adk.JSONRPCSuccessResponse, error) {
			task := adk.Task{
				ID:        "test-task-123",
				ContextID: "test-context-456",
				Kind:      "task",
				Artifacts: []adk.Artifact{
					{
						ArtifactID:  "artifact-1",
						Name:        stringPtr("test-artifact"),
						Description: stringPtr("A test artifact"),
						Parts: []adk.Part{
							map[string]any{
								"kind": "text",
								"text": "Hello, World!",
								"name": "hello.txt",
							},
						},
					},
					{
						ArtifactID: "artifact-2",
						Parts: []adk.Part{
							map[string]any{
								"kind": "data",
								"data": map[string]any{"key": "value"},
								"name": "data.json",
							},
						},
					},
				},
			}

			return &adk.JSONRPCSuccessResponse{
				ID:     "1",
				Result: task,
			}, nil
		},
	}
	a2aClient = mockClient

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().String("task-id", "test-task-123", "Task ID to download artifacts from")
	cmd.Flags().String("artifact-id", "", "Specific artifact ID to download")
	cmd.Flags().StringP("output", "o", tempDir, "Output directory for downloaded artifacts")
	_ = cmd.Flag("task-id").Value.Set("test-task-123")
	_ = cmd.Flag("output").Value.Set(tempDir)

	err := downloadArtifactsCmd.RunE(cmd, []string{})

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
		"Found 2 artifact(s) to download",
		"Downloading artifact: artifact-1",
		"Downloading artifact: artifact-2",
		"Successfully downloaded 2 artifact(s)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", part, output)
		}
	}

	// Check if files were created
	artifact1Dir := fmt.Sprintf("%s/artifact-1", tempDir)
	artifact2Dir := fmt.Sprintf("%s/artifact-2", tempDir)

	if _, err := os.Stat(artifact1Dir); os.IsNotExist(err) {
		t.Errorf("Expected artifact-1 directory to exist at %s", artifact1Dir)
	}

	if _, err := os.Stat(artifact2Dir); os.IsNotExist(err) {
		t.Errorf("Expected artifact-2 directory to exist at %s", artifact2Dir)
	}

	// Check if text file was created
	textFile := fmt.Sprintf("%s/hello.txt", artifact1Dir)
	if content, err := os.ReadFile(textFile); err != nil {
		t.Errorf("Expected to read text file %s, got error: %v", textFile, err)
	} else if string(content) != "Hello, World!" {
		t.Errorf("Expected text file content to be 'Hello, World!', got: %s", string(content))
	}
}

func TestDownloadArtifactsCmd_NoArtifacts(t *testing.T) {
	originalClient := a2aClient
	originalLogger := logger

	testLogger, _ := zap.NewDevelopment()
	logger = testLogger

	mockClient := &mockA2AClient{
		getTaskFunc: func(ctx context.Context, params adk.TaskQueryParams) (*adk.JSONRPCSuccessResponse, error) {
			task := adk.Task{
				ID:        "test-task-123",
				ContextID: "test-context-456",
				Kind:      "task",
				Artifacts: []adk.Artifact{}, // No artifacts
			}

			return &adk.JSONRPCSuccessResponse{
				ID:     "1",
				Result: task,
			}, nil
		},
	}
	a2aClient = mockClient

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().String("task-id", "test-task-123", "Task ID to download artifacts from")
	cmd.Flags().String("artifact-id", "", "Specific artifact ID to download")
	cmd.Flags().StringP("output", "o", "./downloads", "Output directory for downloaded artifacts")
	_ = cmd.Flag("task-id").Value.Set("test-task-123")

	err := downloadArtifactsCmd.RunE(cmd, []string{})

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

	expectedMessage := "No artifacts found for task ID: test-task-123"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", expectedMessage, output)
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
