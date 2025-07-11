# A2A Debugger Example

This example demonstrates how to use the a2a-debugger CLI tool to debug and monitor A2A (Agent-to-Agent) servers in a containerized environment.

## 🚀 Quick Start

1. **Start the services:**

```bash
docker compose up -d
```

2. **Test the connection:**

```bash
docker compose run --rm a2a-debugger connect
```

3. **List available tasks:**

```bash
docker compose run --rm a2a-debugger tasks list
```

4. **Test streaming capabilities:**

```bash
# Check if streaming is supported
docker compose run --rm a2a-debugger connect

# Try a streaming task
docker compose run --rm a2a-debugger tasks submit-streaming "Hello, can you demonstrate streaming?"
```

## 🔧 Available Commands

### Connection Testing

```bash
# Test connection to the A2A server and display agent information
docker compose run --rm a2a-debugger connect

# Get detailed agent card information in JSON format
docker compose run --rm a2a-debugger agent-card
```

### Task Management

```bash
# List all tasks
docker compose run --rm a2a-debugger tasks list

# List tasks with filtering
docker compose run --rm a2a-debugger tasks list --state working --limit 10

# Submit a new task
docker compose run --rm a2a-debugger tasks submit "Hello, can you help me?"

# Submit a task with a specific context ID
docker compose run --rm a2a-debugger tasks submit "Continue our conversation" --context-id ctx-123

# Get detailed information about a specific task
docker compose run --rm a2a-debugger tasks get <task-id>

# View conversation history for a context
docker compose run --rm a2a-debugger tasks history <context-id>
```

### Streaming Task Management

```bash
# Submit a streaming task and watch real-time responses
docker compose run --rm a2a-debugger tasks submit-streaming "Hello, can you help me with a coding task?"

# Submit streaming task with context continuity
docker compose run --rm a2a-debugger tasks submit-streaming "Start a conversation" --context-id streaming-ctx-123
docker compose run --rm a2a-debugger tasks submit-streaming "Continue the conversation" --context-id streaming-ctx-123

# Show raw streaming event data for debugging
docker compose run --rm a2a-debugger tasks submit-streaming "Debug streaming response" --raw

# Test streaming capabilities with complex requests
docker compose run --rm a2a-debugger tasks submit-streaming "Write a Python function to calculate fibonacci numbers"

# Monitor streaming artifacts and status updates
docker compose run --rm a2a-debugger tasks submit-streaming "Explain how A2A streaming works" --context-id docs-ctx
```

### Configuration Management

```bash
# Set configuration values
docker compose run --rm a2a-debugger config set server-url http://a2a-server:8080

# Get configuration values
docker compose run --rm a2a-debugger config get server-url

# List all configuration
docker compose run --rm a2a-debugger config list
```

### Utility Commands

```bash
# Show version information
docker compose run --rm a2a-debugger version

# Get help for any command
docker compose run --rm a2a-debugger --help
docker compose run --rm a2a-debugger tasks --help
```

## 📝 Notes

- The a2a-server runs in demo mode (`APP_DEMO_MODE=true`)
- All services communicate over the `a2a-network` bridge network
- The debugger is configured to use `http://a2a-server:8080` as the default server URL
- Use `connect` command first to check if the A2A server supports streaming capabilities
- Streaming commands provide real-time monitoring of task execution and responses
- The `--raw` flag is useful for debugging streaming protocol compliance and event structure

## 🔗 Related Documentation

- [A2A Debugger Main Repository](https://github.com/inference-gateway/a2a-debugger)
- [A2A ADK Documentation](https://github.com/inference-gateway/a2a)
- [Inference Gateway](https://github.com/inference-gateway)
