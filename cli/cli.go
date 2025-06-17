package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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
)

var rootCmd = &cobra.Command{
	Use:   "a2a-debugger",
	Short: "A debugging tool for A2A (Agent-to-Agent) servers",
	Long: `A2A Debugger is a command-line tool for debugging and monitoring A2A servers.
It allows you to connect to A2A servers, list tasks, view conversation histories,
and inspect task statuses.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
		initA2AClient()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.a2a-debugger.yaml)")
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

	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(listTasksCmd)
	rootCmd.AddCommand(getTaskCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(agentCardCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".a2a-debugger")
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
	logger.Info("A2A client initialized", zap.String("server_url", serverURL))
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Test connection to A2A server",
	Long:  "Tests the connection to the A2A server and displays agent information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		logger.Info("Testing connection to A2A server...")

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
	Use:   "list-tasks",
	Short: "List available tasks and their statuses",
	Long:  "Retrieves and displays a list of tasks from the A2A server with their current statuses.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

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

		logger.Info("Listing tasks", zap.Any("params", params))

		resp, err := a2aClient.ListTasks(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}

		// Parse the response
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
	Use:   "get-task [task-id]",
	Short: "Get detailed information about a specific task",
	Long:  "Retrieves detailed information about a specific task including its history and current status.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		taskID := args[0]

		historyLength, _ := cmd.Flags().GetInt("history-length")

		params := adk.TaskQueryParams{
			ID: taskID,
		}

		if historyLength > 0 {
			params.HistoryLength = &historyLength
		}

		logger.Info("Getting task", zap.String("task_id", taskID))

		resp, err := a2aClient.GetTask(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Parse the response
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

		ctx := context.Background()
		params := adk.TaskListParams{
			ContextID: &contextID,
			Limit:     100,
		}

		logger.Info("Getting conversation history", zap.String("context_id", contextID))

		resp, err := a2aClient.ListTasks(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to list tasks for context: %w", err)
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

		logger.Info("Getting agent card")

		agentCard, err := a2aClient.GetAgentCard(ctx)
		if err != nil {
			return fmt.Errorf("failed to get agent card: %w", err)
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

func init() {
	listTasksCmd.Flags().String("state", "", "Filter by task state (submitted, working, completed, failed)")
	listTasksCmd.Flags().String("context-id", "", "Filter by context ID")
	listTasksCmd.Flags().Int("limit", 50, "Maximum number of tasks to return")
	listTasksCmd.Flags().Int("offset", 0, "Number of tasks to skip")
	getTaskCmd.Flags().Int("history-length", 0, "Number of history messages to include")
}
