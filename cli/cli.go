package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	a2a "github.com/inference-gateway/a2a-debugger/a2a"
	client "github.com/inference-gateway/adk/client"
	adk "github.com/inference-gateway/adk/types"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	zap "go.uber.org/zap"
	yaml "gopkg.in/yaml.v3"
)

var (
	cfgFile   string
	logger    *zap.Logger
	a2aClient client.A2AClient

	appVersion  string
	buildCommit string
	buildDate   string
)

var rootCmd = &cobra.Command{
	Use:   "a2a",
	Short: "A debugging tool for A2A (Agent-to-Agent) servers",
	Long: `A2A Debugger is a command-line tool for debugging and monitoring A2A servers.
It allows you to connect to A2A servers, list tasks, view conversation histories,
and inspect task statuses.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
	},
}

func Execute(version, commit, date string) {
	appVersion = version
	buildCommit = commit
	buildDate = date

	rootCmd.Version = version

	rootCmd.SetVersionTemplate(`A2A Debugger
Version:    {{.Version}}
Commit:     ` + commit + `
Built:      ` + date + `
`)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.a2a.yaml)")
	rootCmd.PersistentFlags().String("server-url", "http://localhost:8080", "A2A server URL")
	rootCmd.PersistentFlags().Duration("timeout", 30*time.Second, "Request timeout")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().Bool("insecure", false, "Skip TLS verification")
	rootCmd.PersistentFlags().StringP("output", "o", "yaml", "Output format (yaml|json)")

	err := viper.BindPFlag("server-url", rootCmd.PersistentFlags().Lookup("server-url"))
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	err = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	err = viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	err = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	if err != nil {
		log.Fatalf("bind error: %v", err)
	}

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	tasksCmd.AddCommand(listTasksCmd)
	tasksCmd.AddCommand(getTaskCmd)
	tasksCmd.AddCommand(historyCmd)
	tasksCmd.AddCommand(submitTaskCmd)
	tasksCmd.AddCommand(submitStreamingTaskCmd)

	artifactsCmd.AddCommand(downloadArtifactsCmd)

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(artifactsCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(agentCardCmd)
	rootCmd.AddCommand(versionCmd)

	listTasksCmd.Flags().String("state", "", "Filter by task state (submitted, working, completed, failed)")
	listTasksCmd.Flags().String("context-id", "", "Filter by context ID")
	listTasksCmd.Flags().Int("limit", 50, "Maximum number of tasks to return")
	listTasksCmd.Flags().Int("offset", 0, "Number of tasks to skip")
	getTaskCmd.Flags().Int("history-length", 0, "Number of history messages to include")
	submitTaskCmd.Flags().String("context-id", "", "Context ID for the task (optional, will generate new context if not provided)")
	submitTaskCmd.Flags().String("task-id", "", "Task ID to resume (optional)")
	submitStreamingTaskCmd.Flags().String("context-id", "", "Context ID for the task (optional, will generate new context if not provided)")
	submitStreamingTaskCmd.Flags().String("task-id", "", "Task ID to resume (optional)")
	submitStreamingTaskCmd.Flags().Bool("raw", false, "Show raw streaming event data instead of formatted output")
	downloadArtifactsCmd.Flags().String("task-id", "", "Task ID to download artifacts from (required)")
	downloadArtifactsCmd.Flags().String("artifact-id", "", "Specific artifact ID to download (optional, downloads all if not specified)")
	downloadArtifactsCmd.Flags().StringP("output", "o", "./downloads", "Output directory for downloaded artifacts")
	err = downloadArtifactsCmd.MarkFlagRequired("task-id")
	if err != nil {
		log.Fatalf("mark flag required error: %v", err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".a2a")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initLogger() {
	var err error
	if viper.GetBool("debug") {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

func initA2AClient() {
	serverURL := viper.GetString("server-url")
	timeout := viper.GetDuration("timeout")

	config := client.DefaultConfig(serverURL)
	config.Timeout = timeout
	config.Logger = logger

	a2aClient = client.NewClientWithConfig(config)
	logger.Debug("A2A client initialized", zap.String("server_url", serverURL))
}

// ensureA2AClient initializes the A2A client if it hasn't been initialized yet
func ensureA2AClient() {
	if a2aClient == nil {
		initA2AClient()
	}
}

// handleA2AError checks if the error is a MethodNotFoundError and returns a user-friendly message
func handleA2AError(err error, method string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	if strings.Contains(errStr, "MethodNotFoundError") || strings.Contains(errStr, "-32601") {
		displayMethod := method
		if displayMethod == "" {
			displayMethod = "method"
		}

		return fmt.Errorf("❌ Method '%s' not implemented by the agent", displayMethod)
	}

	var jsonErr struct {
		Error *a2a.MethodNotFoundError `json:"error,omitempty"`
	}
	if jsonParseErr := json.Unmarshal([]byte(errStr), &jsonErr); jsonParseErr == nil && jsonErr.Error != nil {
		displayMethod := method
		if displayMethod == "" {
			displayMethod = "method"
		}
		return fmt.Errorf("❌ Method '%s' not implemented by the agent", displayMethod)
	}

	return err
}

// OutputFormat represents supported output formats
type OutputFormat string

const (
	OutputFormatYAML OutputFormat = "yaml"
	OutputFormatJSON OutputFormat = "json"
)

// getOutputFormat returns the configured output format
func getOutputFormat() OutputFormat {
	format := viper.GetString("output")
	switch strings.ToLower(format) {
	case "json":
		return OutputFormatJSON
	case "yaml":
		return OutputFormatYAML
	default:
		return OutputFormatYAML // Default to YAML
	}
}

// formatOutput formats the given data according to the specified format
func formatOutput(data any) ([]byte, error) {
	format := getOutputFormat()
	switch format {
	case OutputFormatJSON:
		return json.MarshalIndent(data, "", "  ")
	case OutputFormatYAML:
		return yaml.Marshal(data)
	default:
		return yaml.Marshal(data)
	}
}

// printFormatted outputs the data in the configured format
func printFormatted(data any) error {
	output, err := formatOutput(data)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	fmt.Print(string(output))
	return nil
}

// Config namespace command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  "Commands for managing A2A debugger configuration settings.",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  "Set a configuration value in the A2A debugger config file.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
			if err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
		}

		fmt.Printf("✅ Configuration updated: %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  "Get a configuration value from the A2A debugger config file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := viper.Get(key)

		if value == nil {
			fmt.Printf("Configuration key '%s' not found\n", key)
			return nil
		}

		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  "List all configuration values from the A2A debugger config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()
		return printFormatted(settings)
	},
}

// Tasks namespace command
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Task management commands",
	Long:  "Commands for managing and inspecting A2A tasks.",
}

// Artifacts namespace command
var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Artifact management commands",
	Long:  "Commands for downloading and managing A2A task artifacts.",
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Test connection to A2A server",
	Long:  "Tests the connection to the A2A server and displays agent information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		logger.Debug("Testing connection to A2A server...")
		ensureA2AClient()

		agentCard, err := a2aClient.GetAgentCard(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to A2A server: %w", err)
		}

		output := map[string]any{
			"connected": true,
			"agent":     agentCard,
		}

		return printFormatted(output)
	},
}

var listTasksCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tasks and their statuses",
	Long:  "Retrieves and displays a list of tasks from the A2A server with their current statuses.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		state, _ := cmd.Flags().GetString("state")
		contextID, _ := cmd.Flags().GetString("context-id")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		params := adk.TaskListParams{
			Limit:  limit,
			Offset: offset,
		}

		if state != "" {
			taskState := adk.TaskState(state)
			params.State = &taskState
		}

		if contextID != "" {
			params.ContextID = &contextID
		}

		logger.Debug("Listing tasks", zap.Any("params", params))

		resp, err := a2aClient.ListTasks(ctx, params)
		if err != nil {
			return handleA2AError(err, "tasks/list")
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		var taskList adk.TaskList
		if err := json.Unmarshal(resultBytes, &taskList); err != nil {
			return fmt.Errorf("failed to unmarshal task list: %w", err)
		}

		output := map[string]any{
			"tasks":   taskList.Tasks,
			"total":   taskList.Total,
			"showing": len(taskList.Tasks),
		}

		return printFormatted(output)
	},
}

var getTaskCmd = &cobra.Command{
	Use:   "get [task-id]",
	Short: "Get detailed information about a specific task",
	Long:  "Retrieves detailed information about a specific task including its history and current status.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		taskID := args[0]

		historyLength, _ := cmd.Flags().GetInt("history-length")

		params := adk.TaskQueryParams{
			ID: taskID,
		}

		if historyLength > 0 {
			params.HistoryLength = &historyLength
		}

		logger.Debug("Getting task", zap.String("task_id", taskID))

		resp, err := a2aClient.GetTask(ctx, params)
		if err != nil {
			return handleA2AError(err, "tasks/get")
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		var task adk.Task
		if err := json.Unmarshal(resultBytes, &task); err != nil {
			return fmt.Errorf("failed to unmarshal task: %w", err)
		}

		return printFormatted(task)
	},
}

var historyCmd = &cobra.Command{
	Use:   "history [context-id]",
	Short: "Get conversation history for a specific context",
	Long:  "Retrieves the conversation history for a specific context ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextID := args[0]
		ensureA2AClient()

		ctx := context.Background()
		params := adk.TaskListParams{
			ContextID: &contextID,
			Limit:     100,
		}

		logger.Debug("Getting conversation history", zap.String("context_id", contextID))

		resp, err := a2aClient.ListTasks(ctx, params)
		if err != nil {
			return handleA2AError(err, "tasks/list")
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		var taskList adk.TaskList
		if err := json.Unmarshal(resultBytes, &taskList); err != nil {
			return fmt.Errorf("failed to unmarshal task list: %w", err)
		}

		output := map[string]any{
			"context_id": contextID,
			"tasks":      taskList.Tasks,
		}

		return printFormatted(output)
	},
}

var agentCardCmd = &cobra.Command{
	Use:   "agent-card",
	Short: "Get agent card information",
	Long:  "Retrieves the agent card information from the A2A server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		logger.Debug("Getting agent card")

		agentCard, err := a2aClient.GetAgentCard(ctx)
		if err != nil {
			return handleA2AError(err, "agent-card")
		}

		return printFormatted(agentCard)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Display version information including version number, commit hash, and build date.",
	RunE: func(cmd *cobra.Command, args []string) error {
		version := map[string]any{
			"name":    "A2A Debugger",
			"version": appVersion,
			"commit":  buildCommit,
			"built":   buildDate,
		}
		return printFormatted(version)
	},
}

var submitTaskCmd = &cobra.Command{
	Use:   "submit [message]",
	Short: "Submit a new task to the A2A server",
	Long:  "Submits a new task to the A2A server with the specified message.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		message := args[0]
		contextID, _ := cmd.Flags().GetString("context-id")
		taskID, _ := cmd.Flags().GetString("task-id")

		messageID := fmt.Sprintf("msg-%d", time.Now().Unix())

		params := adk.MessageSendParams{
			Message: adk.Message{
				Kind:      "message",
				MessageID: messageID,
				Role:      "user",
				Parts: []adk.Part{
					map[string]any{
						"kind": "text",
						"text": message,
					},
				},
			},
		}

		if contextID != "" {
			params.Message.ContextID = &contextID
		}

		if taskID != "" {
			params.Message.TaskID = &taskID
		}

		logger.Debug("submitting new task", zap.String("message", message), zap.String("context_id", contextID), zap.String("task_id", taskID))

		resp, err := a2aClient.SendTask(ctx, params)
		if err != nil {
			return handleA2AError(err, "message/send")
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		var task adk.Task
		if err := json.Unmarshal(resultBytes, &task); err != nil {
			return fmt.Errorf("failed to unmarshal task: %w", err)
		}

		output := map[string]any{
			"submitted":  true,
			"message_id": messageID,
			"task":       task,
		}

		return printFormatted(output)
	},
}

var submitStreamingTaskCmd = &cobra.Command{
	Use:   "submit-streaming [message]",
	Short: "Submit a new streaming task to the A2A server",
	Long:  "Submits a new streaming task to the A2A server with the specified message and displays streaming responses.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		message := args[0]
		contextID, _ := cmd.Flags().GetString("context-id")
		taskID, _ := cmd.Flags().GetString("task-id")
		showRaw, _ := cmd.Flags().GetBool("raw")

		messageID := fmt.Sprintf("msg-%d", time.Now().Unix())
		startTime := time.Now()

		params := adk.MessageSendParams{
			Message: adk.Message{
				Kind:      "message",
				MessageID: messageID,
				Role:      "user",
				Parts: []adk.Part{
					map[string]any{
						"kind": "text",
						"text": message,
					},
				},
			},
		}

		if contextID != "" {
			params.Message.ContextID = &contextID
		}

		if taskID != "" {
			params.Message.TaskID = &taskID
		}

		logger.Debug("submitting new streaming task", zap.String("message", message), zap.String("context_id", contextID), zap.String("task_id", taskID))

		eventChan := make(chan any, 100)

		go func() {
			defer close(eventChan)
			err := a2aClient.SendTaskStreaming(ctx, params, eventChan)
			if err != nil {
				logger.Error("Streaming error", zap.Error(err))
			}
		}()

		fmt.Printf("✅ Streaming task submitted successfully!\n\n")
		if contextID != "" {
			fmt.Printf("Context ID: %s\n", contextID)
		}
		fmt.Printf("Message ID: %s\n", messageID)
		fmt.Printf("\n🔄 Streaming responses:\n\n")

		var streamingSummary struct {
			TaskID          string
			ContextID       string
			FinalStatus     string
			StatusUpdates   int
			ArtifactUpdates int
			TotalEvents     int
			FinalMessage    *adk.Message
		}

		for event := range eventChan {
			streamingSummary.TotalEvents++

			eventJSON, err := json.Marshal(event)
			if err != nil {
				logger.Error("Failed to marshal event", zap.Error(err))
				continue
			}

			var genericEvent map[string]any
			if err := json.Unmarshal(eventJSON, &genericEvent); err != nil {
				logger.Error("Failed to unmarshal generic event", zap.Error(err))
				continue
			}

			kind, ok := genericEvent["kind"].(string)
			if ok {
				switch kind {
				case "status-update":
					streamingSummary.StatusUpdates++
					var statusEvent a2a.TaskStatusUpdateEvent
					if err := json.Unmarshal(eventJSON, &statusEvent); err == nil {
						if streamingSummary.TaskID == "" {
							streamingSummary.TaskID = statusEvent.TaskID
						}
						if streamingSummary.ContextID == "" {
							streamingSummary.ContextID = statusEvent.ContextID
						}
						streamingSummary.FinalStatus = string(statusEvent.Status.State)
						if statusEvent.Status.Message != nil {
							adkParts := make([]adk.Part, len(statusEvent.Status.Message.Parts))
							for i, part := range statusEvent.Status.Message.Parts {
								adkParts[i] = adk.Part(part)
							}

							adkMessage := &adk.Message{
								Kind:      statusEvent.Status.Message.Kind,
								MessageID: statusEvent.Status.Message.MessageID,
								Role:      statusEvent.Status.Message.Role,
								Parts:     adkParts,
								ContextID: statusEvent.Status.Message.ContextID,
							}
							streamingSummary.FinalMessage = adkMessage
						}
					}
				case "artifact-update":
					streamingSummary.ArtifactUpdates++
					var artifactEvent a2a.TaskArtifactUpdateEvent
					if err := json.Unmarshal(eventJSON, &artifactEvent); err == nil {
						if streamingSummary.TaskID == "" {
							streamingSummary.TaskID = artifactEvent.TaskID
						}
						if streamingSummary.ContextID == "" {
							streamingSummary.ContextID = artifactEvent.ContextID
						}
					}
				}
			}

			if showRaw {
				eventJSONFormatted, err := json.MarshalIndent(event, "", "  ")
				if err != nil {
					logger.Error("Failed to marshal event", zap.Error(err))
					continue
				}
				fmt.Printf("📡 Raw Event:\n%s\n\n", eventJSONFormatted)
			} else {
				if !ok {
					fmt.Printf("🔔 Unknown Event (no kind field)\n")
					continue
				}

				switch kind {
				case "status-update":
					var statusEvent a2a.TaskStatusUpdateEvent
					if err := json.Unmarshal(eventJSON, &statusEvent); err != nil {
						logger.Error("Failed to unmarshal status event", zap.Error(err))
						continue
					}

					fmt.Printf("📊 Status Update: %s", statusEvent.Status.State)
					if statusEvent.Status.Message != nil {
						fmt.Printf(" (Message: %s)", statusEvent.Status.Message.MessageID)
					}
					if statusEvent.Final {
						fmt.Printf(" [FINAL]")
					}
					fmt.Printf("\n")

					if statusEvent.Status.Message != nil && len(statusEvent.Status.Message.Parts) > 0 {
						fmt.Printf("\n💬 Agent Response:\n")
						for _, part := range statusEvent.Status.Message.Parts {
							if partMap, ok := part.(map[string]any); ok {
								if kind, ok := partMap["kind"].(string); ok && kind == "text" {
									if text, ok := partMap["text"].(string); ok {
										fmt.Printf("%s\n", text)
									}
								}
							}
						}
						fmt.Printf("\n")
					}

				case "artifact-update":
					var artifactEvent a2a.TaskArtifactUpdateEvent
					if err := json.Unmarshal(eventJSON, &artifactEvent); err != nil {
						logger.Error("Failed to unmarshal artifact event", zap.Error(err))
						continue
					}

					fmt.Printf("📄 Artifact Update:\n")
					fmt.Printf("  Artifact ID: %s\n", artifactEvent.Artifact.ArtifactID)
					if artifactEvent.Artifact.Name != nil {
						fmt.Printf("  Name: %s\n", *artifactEvent.Artifact.Name)
					}
					if artifactEvent.Artifact.Description != nil {
						fmt.Printf("  Description: %s\n", *artifactEvent.Artifact.Description)
					}
					if len(artifactEvent.Artifact.Parts) > 0 {
						fmt.Printf("  Parts:\n")
						for i, part := range artifactEvent.Artifact.Parts {
							if partMap, ok := part.(map[string]any); ok {
								if kind, exists := partMap["kind"]; exists {
									fmt.Printf("    Part %d: [%v]", i+1, kind)
									if text, exists := partMap["text"]; exists {
										fmt.Printf(" %v", text)
									}
									fmt.Printf("\n")
								}
							}
						}
					}
					if artifactEvent.LastChunk != nil && *artifactEvent.LastChunk {
						fmt.Printf("  [LAST CHUNK]\n")
					}

				default:
					fmt.Printf("🔔 Unknown Event Type: %s\n", kind)
				}
				fmt.Printf("\n")
			}
		}

		duration := time.Since(startTime)

		fmt.Printf("✅ Streaming completed!\n\n")
		fmt.Printf("📋 Streaming Summary:\n")
		fmt.Printf("  Task ID: %s\n", streamingSummary.TaskID)
		fmt.Printf("  Context ID: %s\n", streamingSummary.ContextID)
		fmt.Printf("  Final Status: %s\n", streamingSummary.FinalStatus)
		fmt.Printf("  Duration: %s\n", duration.Round(time.Millisecond))
		fmt.Printf("  Total Events: %d\n", streamingSummary.TotalEvents)
		fmt.Printf("    Status Updates: %d\n", streamingSummary.StatusUpdates)
		fmt.Printf("    Artifact Updates: %d\n", streamingSummary.ArtifactUpdates)

		if streamingSummary.FinalMessage != nil {
			fmt.Printf("  Final Message Parts: %d\n", len(streamingSummary.FinalMessage.Parts))
		}

		fmt.Printf("\n")
		return nil
	},
}

var downloadArtifactsCmd = &cobra.Command{
	Use:   "download",
	Short: "Download artifacts from a specific task",
	Long:  "Downloads artifacts from a specific task using the task ID. Optionally download only a specific artifact by ID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ensureA2AClient()

		taskID, _ := cmd.Flags().GetString("task-id")
		artifactID, _ := cmd.Flags().GetString("artifact-id")
		outputDir, _ := cmd.Flags().GetString("output")

		logger.Debug("Downloading artifacts", zap.String("task_id", taskID), zap.String("artifact_id", artifactID), zap.String("output_dir", outputDir))

		// Get the task to retrieve its artifacts
		params := adk.TaskQueryParams{
			ID: taskID,
		}

		resp, err := a2aClient.GetTask(ctx, params)
		if err != nil {
			return handleA2AError(err, "tasks/get")
		}

		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		var task adk.Task
		if err := json.Unmarshal(resultBytes, &task); err != nil {
			return fmt.Errorf("failed to unmarshal task: %w", err)
		}

		// Check if task has artifacts
		if len(task.Artifacts) == 0 {
			fmt.Printf("⚠️  No artifacts found for task ID: %s\n", taskID)
			return nil
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Filter artifacts if specific artifact ID is provided
		var artifactsToDownload []adk.Artifact
		if artifactID != "" {
			for _, artifact := range task.Artifacts {
				if artifact.ArtifactID == artifactID {
					artifactsToDownload = append(artifactsToDownload, artifact)
					break
				}
			}
			if len(artifactsToDownload) == 0 {
				return fmt.Errorf("artifact with ID '%s' not found in task '%s'", artifactID, taskID)
			}
		} else {
			artifactsToDownload = task.Artifacts
		}

		fmt.Printf("📦 Found %d artifact(s) to download from task %s\n\n", len(artifactsToDownload), taskID)

		downloadedCount := 0
		for _, artifact := range artifactsToDownload {
			fmt.Printf("🔽 Downloading artifact: %s\n", artifact.ArtifactID)
			if artifact.Name != nil {
				fmt.Printf("   Name: %s\n", *artifact.Name)
			}
			if artifact.Description != nil {
				fmt.Printf("   Description: %s\n", *artifact.Description)
			}

			// Create artifact directory
			artifactDir := filepath.Join(outputDir, artifact.ArtifactID)
			if err := os.MkdirAll(artifactDir, 0755); err != nil {
				fmt.Printf("❌ Failed to create artifact directory %s: %v\n", artifactDir, err)
				continue
			}

			// Process each part of the artifact
			partCount := 0
			for i, part := range artifact.Parts {
				partMap, ok := part.(map[string]any)
				if !ok {
					fmt.Printf("⚠️  Skipping part %d: invalid format\n", i+1)
					continue
				}

				kind, exists := partMap["kind"].(string)
				if !exists {
					fmt.Printf("⚠️  Skipping part %d: no kind specified\n", i+1)
					continue
				}

				switch kind {
				case "text":
					if text, ok := partMap["text"].(string); ok {
						filename := fmt.Sprintf("part_%d.txt", i+1)
						if name, exists := partMap["name"].(string); exists && name != "" {
							filename = name
							if !strings.HasSuffix(strings.ToLower(filename), ".txt") {
								filename += ".txt"
							}
						}
						
						filePath := filepath.Join(artifactDir, filename)
						if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
							fmt.Printf("❌ Failed to write text part to %s: %v\n", filePath, err)
							continue
						}
						fmt.Printf("   ✅ Saved text part: %s\n", filename)
						partCount++
					}

				case "file":
					filename := fmt.Sprintf("part_%d_file", i+1)
					if name, exists := partMap["name"].(string); exists && name != "" {
						filename = name
					}

					// Handle file content (could be base64 encoded or direct content)
					var content []byte
					if data, exists := partMap["data"].(string); exists {
						// Assume base64 encoded content
						decoded, err := base64.StdEncoding.DecodeString(data)
						if err != nil {
							// If base64 decode fails, treat as plain text
							content = []byte(data)
						} else {
							content = decoded
						}
					} else if text, exists := partMap["text"].(string); exists {
						content = []byte(text)
					} else {
						fmt.Printf("⚠️  Skipping file part %d: no content found\n", i+1)
						continue
					}

					filePath := filepath.Join(artifactDir, filename)
					if err := os.WriteFile(filePath, content, 0644); err != nil {
						fmt.Printf("❌ Failed to write file part to %s: %v\n", filePath, err)
						continue
					}
					fmt.Printf("   ✅ Saved file part: %s\n", filename)
					partCount++

				case "data":
					// Handle structured data (JSON)
					filename := fmt.Sprintf("part_%d.json", i+1)
					if name, exists := partMap["name"].(string); exists && name != "" {
						filename = name
						if !strings.HasSuffix(strings.ToLower(filename), ".json") {
							filename += ".json"
						}
					}

					var dataContent []byte
					if data, exists := partMap["data"]; exists {
						dataContent, err = json.MarshalIndent(data, "", "  ")
						if err != nil {
							fmt.Printf("❌ Failed to marshal data part %d: %v\n", i+1, err)
							continue
						}
					} else {
						fmt.Printf("⚠️  Skipping data part %d: no data found\n", i+1)
						continue
					}

					filePath := filepath.Join(artifactDir, filename)
					if err := os.WriteFile(filePath, dataContent, 0644); err != nil {
						fmt.Printf("❌ Failed to write data part to %s: %v\n", filePath, err)
						continue
					}
					fmt.Printf("   ✅ Saved data part: %s\n", filename)
					partCount++

				default:
					fmt.Printf("⚠️  Skipping part %d: unsupported kind '%s'\n", i+1, kind)
				}
			}

			if partCount > 0 {
				fmt.Printf("   📂 Artifact saved to: %s (%d parts)\n", artifactDir, partCount)
				downloadedCount++
			} else {
				fmt.Printf("   ⚠️  No parts could be downloaded for this artifact\n")
			}
			fmt.Println()
		}

		if downloadedCount > 0 {
			fmt.Printf("✅ Successfully downloaded %d artifact(s) to %s\n", downloadedCount, outputDir)
		} else {
			fmt.Printf("❌ No artifacts were successfully downloaded\n")
		}

		return nil
	},
}
