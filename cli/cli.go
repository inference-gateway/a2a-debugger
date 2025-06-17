package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/inference-gateway/a2a-debugger/a2a"
	"github.com/inference-gateway/a2a/adk"
	"github.com/inference-gateway/a2a/adk/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	tasksCmd.AddCommand(listTasksCmd)
	tasksCmd.AddCommand(getTaskCmd)
	tasksCmd.AddCommand(historyCmd)
	tasksCmd.AddCommand(submitTaskCmd)

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(agentCardCmd)
	rootCmd.AddCommand(versionCmd)

	listTasksCmd.Flags().String("state", "", "Filter by task state (submitted, working, completed, failed)")
	listTasksCmd.Flags().String("context-id", "", "Filter by context ID")
	listTasksCmd.Flags().Int("limit", 50, "Maximum number of tasks to return")
	listTasksCmd.Flags().Int("offset", 0, "Number of tasks to skip")
	getTaskCmd.Flags().Int("history-length", 0, "Number of history messages to include")
	submitTaskCmd.Flags().String("context-id", "", "Context ID for the task (optional, will generate new context if not provided)")
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

		if len(settings) == 0 {
			fmt.Printf("No configuration found\n")
			return nil
		}

		fmt.Printf("📋 Configuration:\n\n")
		for key, value := range settings {
			fmt.Printf("  %s = %v\n", key, value)
		}

		return nil
	},
}

// Tasks namespace command
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Task management commands",
	Long:  "Commands for managing and inspecting A2A tasks.",
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

		fmt.Printf("✅ Successfully connected to A2A server!\n\n")
		fmt.Printf("Agent Information:\n")
		fmt.Printf("  Name: %s\n", agentCard.Name)
		fmt.Printf("  Description: %s\n", agentCard.Description)
		fmt.Printf("  Version: %s\n", agentCard.Version)
		fmt.Printf("  URL: %s\n", agentCard.URL)

		fmt.Printf("\nCapabilities:\n")
		if agentCard.Capabilities.Streaming != nil {
			fmt.Printf("  Streaming: %t\n", *agentCard.Capabilities.Streaming)
		}
		if agentCard.Capabilities.PushNotifications != nil {
			fmt.Printf("  Push Notifications: %t\n", *agentCard.Capabilities.PushNotifications)
		}
		if agentCard.Capabilities.StateTransitionHistory != nil {
			fmt.Printf("  State Transition History: %t\n", *agentCard.Capabilities.StateTransitionHistory)
		}

		return nil
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

		fmt.Printf("📋 Tasks (Total: %d, Showing: %d)\n\n", taskList.Total, len(taskList.Tasks))

		if len(taskList.Tasks) == 0 {
			fmt.Printf("No tasks found.\n")
			return nil
		}

		for i, task := range taskList.Tasks {
			fmt.Printf("%d. Task ID: %s\n", i+1, task.ID)
			fmt.Printf("   Context ID: %s\n", task.ContextID)
			fmt.Printf("   Status: %s\n", task.Status.State)
			if task.Status.Message != nil {
				fmt.Printf("   Message ID: %s\n", task.Status.Message.MessageID)
				fmt.Printf("   Role: %s\n", task.Status.Message.Role)
			}
			if task.Status.Timestamp != nil {
				fmt.Printf("   Timestamp: %s\n", *task.Status.Timestamp)
			}
			fmt.Printf("\n")
		}

		return nil
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

		fmt.Printf("🔍 Task Details\n\n")
		fmt.Printf("ID: %s\n", task.ID)
		fmt.Printf("Context ID: %s\n", task.ContextID)
		fmt.Printf("Status: %s\n", task.Status.State)

		if task.Status.Message != nil {
			fmt.Printf("\nCurrent Message:\n")
			fmt.Printf("  Message ID: %s\n", task.Status.Message.MessageID)
			fmt.Printf("  Role: %s\n", task.Status.Message.Role)
			fmt.Printf("  Parts: %d\n", len(task.Status.Message.Parts))

			for i, part := range task.Status.Message.Parts {
				if partMap, ok := part.(map[string]interface{}); ok {
					if kind, exists := partMap["kind"]; exists {
						fmt.Printf("    %d. Kind: %v\n", i+1, kind)
						if text, exists := partMap["text"]; exists {
							fmt.Printf("       Text: %v\n", text)
						}
					}
				}
			}
		}

		if len(task.History) > 0 {
			fmt.Printf("\nConversation History (%d messages):\n", len(task.History))
			for i, msg := range task.History {
				fmt.Printf("  %d. [%s] %s\n", i+1, msg.Role, msg.MessageID)
				for j, part := range msg.Parts {
					if partMap, ok := part.(map[string]interface{}); ok {
						if text, exists := partMap["text"]; exists {
							fmt.Printf("     Part %d: %v\n", j+1, text)
						}
					}
				}
			}
		}

		return nil
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

		if len(taskList.Tasks) == 0 {
			fmt.Printf("No conversation history found for context: %s\n", contextID)
			return nil
		}

		fmt.Printf("💬 Conversation History for Context: %s\n\n", contextID)

		for _, task := range taskList.Tasks {
			fmt.Printf("Task: %s (Status: %s)\n", task.ID, task.Status.State)

			if len(task.History) > 0 {
				for i, msg := range task.History {
					fmt.Printf("  %d. [%s] %s\n", i+1, msg.Role, msg.MessageID)
					for j, part := range msg.Parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							if text, exists := partMap["text"]; exists {
								fmt.Printf("     %d: %v\n", j+1, text)
							}
						}
					}
				}
			}

			if task.Status.Message != nil {
				fmt.Printf("  Current: [%s] %s\n", task.Status.Message.Role, task.Status.Message.MessageID)
				for j, part := range task.Status.Message.Parts {
					if partMap, ok := part.(map[string]interface{}); ok {
						if text, exists := partMap["text"]; exists {
							fmt.Printf("     %d: %v\n", j+1, text)
						}
					}
				}
			}
			fmt.Printf("\n")
		}

		return nil
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

		agentCardJSON, err := json.MarshalIndent(agentCard, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal agent card: %w", err)
		}

		fmt.Printf("🤖 Agent Card\n\n")
		fmt.Printf("%s\n", agentCardJSON)

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Display version information including version number, commit hash, and build date.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("A2A Debugger\n")
		fmt.Printf("Version:    %s\n", appVersion)
		fmt.Printf("Commit:     %s\n", buildCommit)
		fmt.Printf("Built:      %s\n", buildDate)
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

		if contextID != "" {
			params.Message.ContextID = &contextID
		}

		logger.Debug("Submitting new task", zap.String("message", message), zap.String("context_id", contextID))

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

		fmt.Printf("✅ Task submitted successfully!\n\n")
		fmt.Printf("Task Details:\n")
		fmt.Printf("  Task ID: %s\n", task.ID)
		fmt.Printf("  Context ID: %s\n", task.ContextID)
		fmt.Printf("  Status: %s\n", task.Status.State)
		fmt.Printf("  Message ID: %s\n", messageID)

		return nil
	},
}
