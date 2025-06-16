<div align="center">

# A2A Debugger

[![CI](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml/badge.svg)](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/inference-gateway/a2a-debugger)](https://goreportcard.com/report/github.com/inference-gateway/a2a-debugger)
[![GoDoc](https://godoc.org/github.com/inference-gateway/a2a-debugger?status.svg)](https://godoc.org/github.com/inference-gateway/a2a-debugger)
[![Release](https://img.shields.io/github/release/inference-gateway/a2a-debugger.svg)](https://github.com/inference-gateway/a2a-debugger/releases/latest)

**The ultimate A2A (Agent-to-Agent) troubleshooting and debugging tool**

A powerful command-line utility for debugging, monitoring, and inspecting A2A servers. Connect to A2A servers, list tasks, view conversation histories, and inspect task statuses with ease.

</div>

## 🚀 Features

- **Server Connectivity**: Test connections to A2A servers and retrieve agent information
- **Task Management**: List, filter, and inspect tasks with detailed status information
- **Conversation History**: View detailed conversation histories and message flows
- **Agent Information**: Retrieve and display agent cards with capabilities
- **Flexible Configuration**: Support for configuration files and environment variables
- **Debug Logging**: Comprehensive logging with configurable verbosity levels

## 📦 Installation

### Using Go Install

```bash
go install github.com/inference-gateway/a2a-debugger@latest
```

### From Release

Download the latest binary from the [releases page](https://github.com/inference-gateway/a2a-debugger/releases).

### Build from Source

```bash
git clone https://github.com/inference-gateway/a2a-debugger.git
cd a2a-debugger
task build
```

## 🔧 Usage

### Quick Start

Test connection to an A2A server:

```bash
a2a-debugger connect --server-url http://localhost:8080
```

List all tasks:

```bash
a2a-debugger list-tasks --server-url http://localhost:8080
```

Get specific task details:

```bash
a2a-debugger get-task --server-url http://localhost:8080 --task-id <task-id>
```

View conversation history:

```bash
a2a-debugger history --server-url http://localhost:8080 --context-id <context-id>
```

### Configuration

Create a configuration file at `~/.a2a-debugger.yaml`:

```yaml
server-url: http://localhost:8080
timeout: 30s
debug: false
insecure: false
```

### Command Options

#### Global Options

- `--server-url`: A2A server URL (default: http://localhost:8080)
- `--timeout`: Request timeout (default: 30s)
- `--debug`: Enable debug logging
- `--insecure`: Skip TLS verification
- `--config`: Config file path

#### List Tasks Options

- `--state`: Filter by task state (pending, running, completed, failed)
- `--context-id`: Filter by context ID
- `--limit`: Maximum number of tasks to return
- `--offset`: Number of tasks to skip

### Examples

#### Connect and view agent information

```bash
$ a2a-debugger connect --server-url https://my-agent.example.com

✅ Successfully connected to A2A server!

Agent Information:
  Name: My A2A Agent
  Description: A helpful assistant agent
  Version: 1.0.0
  URL: https://my-agent.example.com

Capabilities:
  Streaming: true
  Push Notifications: false
  State Transition History: true
```

#### List tasks with filtering

```bash
$ a2a-debugger list-tasks --state running --limit 5

📋 Tasks (Total: 23, Showing: 5)

1. Task ID: task-abc123
   Context ID: ctx-xyz789
   Status: running
   Message ID: msg-456
   Role: assistant

2. Task ID: task-def456
   Context ID: ctx-uvw123
   Status: running
   Message ID: msg-789
   Role: user
```

#### View detailed task information

```bash
$ a2a-debugger get-task --task-id task-abc123

📝 Task Details

Task ID: task-abc123
Context ID: ctx-xyz789
Status: completed
Created: 2025-06-17T10:30:00Z
Updated: 2025-06-17T10:35:00Z

Message:
  ID: msg-456
  Role: assistant
  Content: Hello! How can I help you today?
```

## 🛠️ Development

### Prerequisites

- Go 1.24 or later
- [Task](https://taskfile.dev/) for build automation

### Available Tasks

```bash
task generate    # Generate code from schemas
task lint       # Run linting
task build      # Build the application
task test       # Run tests
task clean      # Clean build artifacts
```

### Development Workflow

1. Make your changes
2. Run `task generate` to update generated files
3. Run `task lint` to check code quality
4. Run `task build` to verify compilation
5. Run `task test` to ensure all tests pass

## 📚 Related Projects

- [Inference Gateway](https://github.com/inference-gateway) - Main project
- [A2A ADK](https://github.com/inference-gateway/a2a) - Agent Development Kit
- [Go SDK](https://github.com/inference-gateway/go-sdk) - Go SDK for Inference Gateway
- [TypeScript SDK](https://github.com/inference-gateway/typescript-sdk) - TypeScript SDK
- [Python SDK](https://github.com/inference-gateway/python-sdk) - Python SDK
- [Documentation](https://github.com/inference-gateway/docs) - Project documentation

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
