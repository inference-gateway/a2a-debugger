# A2A Debugger Example

This example demonstrates how to use the a2a-debugger CLI tool to debug and monitor A2A (Agent-to-Agent) servers in a containerized environment.

## 🚀 Quick Start

1. **Start the services:**

```bash
docker compose up -d
```

2. **Test the connection:**

```bash
docker compose run --rm a2a-debugger connect --config /config/.a2a.yaml
```

3. **List available tasks:**

```bash
docker compose run --rm a2a-debugger tasks list --config /config/.a2a.yaml
```

## 🔧 Available Commands

### Connection Testing

```bash
# Test connection to the A2A server and display agent information
docker compose run --rm a2a-debugger connect --config /config/.a2a.yaml

# Get detailed agent card information in JSON format
docker compose run --rm a2a-debugger agent-card --config /config/.a2a.yaml
```

### Task Management

```bash
# List all tasks
docker compose run --rm a2a-debugger tasks list --config /config/.a2a.yaml

# List tasks with filtering
docker compose run --rm a2a-debugger tasks list --state working --limit 10 --config /config/.a2a.yaml

# Get detailed information about a specific task
docker compose run --rm a2a-debugger tasks get <task-id> --config /config/.a2a.yaml

# View conversation history for a context
docker compose run --rm a2a-debugger tasks history <context-id> --config /config/.a2a.yaml
```

### Configuration Management

```bash
# Set configuration values
docker compose run --rm a2a-debugger config set server-url http://a2a-server:8080 --config /config/.a2a.yaml

# Get configuration values
docker compose run --rm a2a-debugger config get server-url --config /config/.a2a.yaml

# List all configuration
docker compose run --rm a2a-debugger config list --config /config/.a2a.yaml
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

## 🔗 Related Documentation

- [A2A Debugger Main Repository](https://github.com/inference-gateway/a2a-debugger)
- [A2A ADK Documentation](https://github.com/inference-gateway/a2a)
- [Inference Gateway](https://github.com/inference-gateway)
