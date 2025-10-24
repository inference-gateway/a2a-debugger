<div align="center">

# A2A Debugger

[![CI](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml/badge.svg)](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/inference-gateway/a2a-debugger?status.svg)](https://godoc.org/github.com/inference-gateway/a2a-debugger)
[![Release](https://img.shields.io/github/release/inference-gateway/a2a-debugger.svg)](https://github.com/inference-gateway/a2a-debugger/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**The ultimate A2A (Agent-to-Agent) troubleshooting and debugging tool**

A powerful command-line utility for debugging, monitoring, and inspecting A2A servers. Connect to A2A servers, list tasks, view conversation histories, and inspect task statuses with ease.

</div>

## ⚠️ Warning

> **This project is in its early stages of development.**
>
> Breaking changes are expected as we actively develop and refine the tool. Use with caution in production environments and be prepared for API changes, configuration format updates, and command-line interface modifications between versions.
>
> We recommend pinning to specific versions in your scripts and monitoring the [CHANGELOG.md](CHANGELOG.md) for breaking changes.

## 🚀 Features

- **Server Connectivity**: Test connections to A2A servers and retrieve agent information
- **Interactive Mode**: Chat interface with streaming and background modes using Bubble Tea UI
- **Task Management**: List, filter, and inspect tasks with detailed status information
- **Real-time Streaming**: Submit streaming tasks and monitor real-time agent responses
- **Streaming Summaries**: Summaries with Task IDs, durations, and event counts
- **Conversation History**: View detailed conversation histories and message flows
- **Agent Information**: Retrieve and display agent cards with capabilities
- **Configuration Management**: Set, get, and list configuration values with namespace commands
- **Flexible Configuration**: Support for configuration files and environment variables
- **Debug Logging**: Comprehensive logging with configurable verbosity levels
- **Namespace Commands**: Organized command structure with `config` and `tasks` namespaces
- **Multiple Output Formats**: Support for YAML (default) and JSON output formats for structured data

## 📦 Installation

### Quick Install (Recommended)

Use our install script to automatically download and install the latest binary:

```bash
curl -fsSL https://raw.githubusercontent.com/inference-gateway/a2a-debugger/main/install.sh | bash
```

Or download and run the script manually:

```bash
wget https://raw.githubusercontent.com/inference-gateway/a2a-debugger/main/install.sh
chmod +x install.sh
./install.sh
```

**Install Options:**

- Install specific version: `./install.sh --version v1.0.0`
- Custom install directory: `INSTALL_DIR=~/bin ./install.sh`
- Show help: `./install.sh --help`

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
a2a connect --server-url http://localhost:8080
```

Set server URL in configuration:

```bash
a2a config set server-url http://localhost:8080
```

List all tasks:

```bash
a2a tasks list
```

Start interactive chat mode:

```bash
a2a interactive
```

Get specific task details:

```bash
a2a tasks get <task-id>
```

View conversation history:

```bash
a2a tasks history <context-id>
```

### Command Structure

The A2A Debugger uses a namespace-based command structure for better organization:

#### Config Commands

```bash
a2a config set <key> <value>    # Set a configuration value
a2a config get <key>            # Get a configuration value
a2a config list                 # List all configuration values
```

#### Task Commands

```bash
a2a tasks list                     # List available tasks
a2a tasks get <task-id>            # Get detailed task information
a2a tasks history <context-id>     # Get conversation history for a context
a2a tasks submit <message>         # Submit a task and get response
a2a tasks submit-streaming <msg>   # Submit streaming task with real-time responses and summary
```

#### Interactive Mode

```bash
a2a interactive                 # Start interactive chat mode with A2A server
```

#### Server Commands

```bash
a2a connect                     # Test connection to A2A server
a2a agent-card                  # Get agent card information
```

### Configuration

Create a configuration file at `~/.a2a.yaml`:

```yaml
server-url: http://localhost:8080
timeout: 30s
debug: false
insecure: false
output: yaml  # or json
```

### Command Options

#### Global Options

- `--server-url`: A2A server URL (default: http://localhost:8080)
- `--timeout`: Request timeout (default: 30s)
- `--debug`: Enable debug logging
- `--insecure`: Skip TLS verification
- `--config`: Config file path
- `--output, -o`: Output format (yaml|json) (default: yaml)

#### Task List Options

- `--state`: Filter by task state (submitted, working, completed, failed)
- `--context-id`: Filter by context ID
- `--limit`: Maximum number of tasks to return (default: 50)
- `--offset`: Number of tasks to skip (default: 0)
- `--include-history`: Include conversation history in the output (default: false)

#### Task Get Options

- `--history-length`: Number of history messages to include

### Interactive Mode

The interactive mode provides a chat-like interface for communicating with A2A agents in real-time:

```bash
a2a interactive
```

**Features:**
- **Dual Modes**: Switch between Streaming (realtime) and Background (long running tasks) modes
- **Live Chat**: Type messages and get real-time responses
- **Bubble Tea UI**: Clean terminal interface with message history and status updates
- **Mode Switching**: Press Tab to switch between streaming and background modes
- **Auto-scrolling**: Message history automatically scrolls to show latest messages
- **Error Handling**: Clear error messages and connection status

**Controls:**
- Type your message and press **Enter** to send
- Press **Tab** to switch between streaming/background modes
- Press **Ctrl+C** to quit

**Example Session:**
```
🤖 A2A Interactive Chat - Streaming Mode

Context ID: ctx-1725670123

ℹ️  🚀 Interactive A2A Chat Session Started
ℹ️  📡 Mode: Streaming (realtime) 
ℹ️  💬 Type your message and press Enter to send
ℹ️  ⌨️  Press Tab to switch modes, Ctrl+C to quit
👤 You: Hello, can you help me with my project?
🤖 Agent: Hello! I'd be happy to help you with your project. What kind of project are you working on?

Interactive Mode - Press Tab to switch modes, Ctrl+C to quit

💬 Hello, can you help me debug this code?
```

### Examples

#### Configuration Management

```bash
# Set server URL in config
$ a2a config set server-url http://localhost:8080

✅ Configuration updated: server-url = http://localhost:8080

# Get current server URL
$ a2a config get server-url

server-url = http://localhost:8080

# List all configuration
$ a2a config list

📋 Configuration:

  server-url = http://localhost:8080
  timeout = 30s
  debug = false
```

#### Connect and view agent information

```bash
$ a2a connect --server-url http://localhost:8080

✅ Successfully connected to A2A server!

Agent Information:
  Name: My A2A Agent
  Description: A helpful assistant agent
  Version: 1.0.0
  URL: http://localhost:8080

Capabilities:
  Streaming: true
  Push Notifications: false
  State Transition History: true
```

#### List tasks with filtering

```bash
$ a2a tasks list --state working --limit 5

📋 Tasks (Total: 23, Showing: 5)

1. Task ID: task-abc123
   Context ID: ctx-xyz789
   Status: working
   Message ID: msg-456
   Role: assistant

2. Task ID: task-def456
   Context ID: ctx-uvw123
   Status: working
   Message ID: msg-789
   Role: user
```

#### Include conversation history in task list

By default, `tasks list` excludes conversation history to keep output manageable. Use `--include-history` to show complete task data:

```bash
# Without history (default - cleaner output)
$ a2a tasks list --limit 1

# With history (complete task data including conversation)
$ a2a tasks list --limit 1 --include-history
```

#### View detailed task information

```bash
$ a2a tasks get task-abc123

� Task Details

ID: task-abc123
Context ID: ctx-xyz789
Status: completed

Current Message:
  Message ID: msg-456
  Role: assistant
  Parts: 1
    1. Kind: text
       Text: Hello! How can I help you today?
```

#### View conversation history

```bash
$ a2a tasks history ctx-xyz789

💬 Conversation History for Context: ctx-xyz789

Task: task-abc123 (Status: completed)
  1. [user] msg-123
     1: I need help with my project
  2. [assistant] msg-456
     1: Hello! How can I help you today?
```

#### Output Formats

By default, all commands output structured data in YAML format. You can switch to JSON using the `-o` flag:

```bash
# YAML output (default)
$ a2a tasks list --limit 2
tasks:
  - id: task-abc123
    context_id: ctx-xyz789
    kind: task
    status:
      state: completed
      message:
        message_id: msg-456
        role: assistant
    artifacts: []
    metadata: {}
  - id: task-def456
    context_id: ctx-uvw123
    kind: task
    status:
      state: working
      message:
        message_id: msg-789
        role: user
    artifacts: []
    metadata: {}
total: 23
showing: 2

# JSON output
$ a2a tasks list --limit 2 -o json
{
  "tasks": [
    {
      "id": "task-abc123",
      "context_id": "ctx-xyz789",
      "kind": "task",
      "status": {
        "state": "completed",
        "message": {
          "message_id": "msg-456",
          "role": "assistant"
        }
      },
      "artifacts": [],
      "metadata": {}
    }
  ],
  "total": 23,
  "showing": 2
}
```

## 🛠️ Development

### Prerequisites

- Go 1.25 or later
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
