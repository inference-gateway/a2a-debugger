<div align="center">

# A2A Debugger

[![CI](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml/badge.svg)](https://github.com/inference-gateway/a2a-debugger/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/inference-gateway/a2a-debugger?status.svg)](https://godoc.org/github.com/inference-gateway/a2a-debugger)
[![Release](https://img.shields.io/github/release/inference-gateway/a2a-debugger.svg)](https://github.com/inference-gateway/a2a-debugger/releases/latest)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

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
- **Task Management**: List, filter, and inspect tasks with detailed status information
- **Real-time Streaming**: Submit streaming tasks and monitor real-time agent responses
- **Streaming Summaries**: Summaries with Task IDs, durations, and event counts
- **Interactive Chat Mode**: A terminal chat UI (built with Bubble Tea) to converse with an agent in streaming or background mode
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

#### Server Commands

```bash
a2a connect                     # Test connection to A2A server
a2a agent-card                  # Get agent card information
a2a auth <token>                # Verify a credential and fetch the authenticated extended card
```

#### Interactive Mode

```bash
a2a interactive                 # Start a chat session (streaming mode)
a2a interactive --background    # Start a chat session in background (long-running task) mode
a2a chat                        # Alias for "a2a interactive"
```

### Configuration

Create a configuration file at `~/.a2a.yaml`:

```yaml
server-url: http://localhost:8080
timeout: 30s
debug: false
insecure: false
output: yaml  # or json
token: ""             # bearer token, stored in plaintext - prefer --token or the TOKEN env var
auth-header: Authorization
```

### Command Options

#### Global Options

- `--server-url`: A2A server URL (default: http://localhost:8080)
- `--timeout`: Request timeout (default: 30s)
- `--debug`: Enable debug logging
- `--insecure`: Skip TLS verification
- `--config`: Config file path
- `--output, -o`: Output format (yaml|json) (default: yaml)
- `--token`: Bearer token (or credential value) sent to the A2A server
- `--auth-header`: Header carrying the credential (default: Authorization)

#### Task List Options

- `--state`: Filter by task state (submitted, working, completed, failed)
- `--context-id`: Filter by context ID
- `--limit`: Maximum number of tasks to return (default: 50)
- `--offset`: Number of tasks to skip (default: 0)
- `--include-history`: Include conversation history in the output (default: false)

#### Task Get Options

- `--history-length`: Number of history messages to include

#### Interactive Mode Options

- `--background, -b`: Use background (long-running task) mode instead of streaming (default: false)
- `--context-id`: Resume an existing context ID (optional; a new one is generated otherwise)

### Authentication

Per the A2A spec the credential is obtained out of band; the debugger only transmits it.
The public agent card is always unauthenticated, every other request carries the token in
the header named by `--auth-header` (for `Authorization` the value is prefixed with `Bearer `).

`a2a auth <token>` verifies a credential by sending one authenticated request (`ListTasks`):
a `401` means the server rejected it, anything else means it was accepted. When the public card
advertises `supportsExtendedAgentCard`, the authenticated extended card is fetched and printed too.

```bash
# Get a token from your identity provider (Keycloak, client credentials grant)
$ TOKEN=$(curl -s http://localhost:8081/realms/inference-gateway-realm/protocol/openid-connect/token \
    -d grant_type=client_credentials -d client_id=inference-gateway-client -d client_secret=very-secret \
    | jq -r .access_token)

# Verify it
$ a2a auth "$TOKEN" --server-url http://localhost:8080

# Use it with any other command
$ a2a tasks list --token "$TOKEN"
$ TOKEN="$TOKEN" a2a interactive

# API-key style servers
$ a2a agent-card --token "$KEY" --auth-header X-Api-Key
```

`a2a config set token <jwt>` persists the token, but it is written in plaintext to
`~/.a2a.yaml` - prefer the `--token` flag or the `TOKEN` environment variable.
A runnable Keycloak setup lives in [`example/`](example/README.md#authentication).

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

auth-header: Authorization
debug: false
insecure: false
output: yaml
server-url: http://localhost:8080
timeout: 30s
token: ""
```

#### Connect and view agent information

```bash
$ a2a connect --server-url http://localhost:8080

agent:
    capabilities:
        extendedagentcard: null
        extensions: []
        pushnotifications: false
        streaming: true
    defaultinputmodes:
        - text/plain
    defaultoutputmodes:
        - text/plain
    description: echo agent for the debugger end-to-end test
    documentationurl: null
    iconurl: null
    name: e2e-agent
    provider: null
    securityrequirements: []
    securityschemes: {}
    signatures: []
    skills: []
    supportedinterfaces:
        - protocolbinding: JSONRPC
          protocolversion: "1.0"
          tenant: null
          url: http://localhost:8080
    version: 0.0.0
connected: true
```

#### List tasks with filtering

```bash
$ a2a tasks list --state completed --limit 5

showing: 1
tasks:
    - artifacts: []
      contextid: demo-ctx
      history: []
      id: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
      metadata: null
      status:
        message:
            contextid: demo-ctx
            extensions: []
            messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
            metadata: null
            parts:
                - data: null
                  filename: null
                  mediatype: null
                  metadata: null
                  raw: null
                  text: 'Echo: I need help with my project'
                  url: null
            referencetaskids: []
            role: ROLE_AGENT
            taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
        state: TASK_STATE_COMPLETED
        timestamp: 2026-10-04T15:01:41.256127955Z
total: 1
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
$ a2a tasks get a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a

artifacts: []
contextid: demo-ctx
history:
    - contextid: demo-ctx
      extensions: []
      messageid: msg-1791126101
      metadata: null
      parts:
        - data: null
          filename: null
          mediatype: null
          metadata: null
          raw: null
          text: I need help with my project
          url: null
      referencetaskids: []
      role: ROLE_USER
      taskid: null
    - contextid: demo-ctx
      extensions: []
      messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
      metadata: null
      parts:
        - data: null
          filename: null
          mediatype: null
          metadata: null
          raw: null
          text: 'Echo: I need help with my project'
          url: null
      referencetaskids: []
      role: ROLE_AGENT
      taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
id: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
metadata: null
status:
    message:
        contextid: demo-ctx
        extensions: []
        messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
        metadata: null
        parts:
            - data: null
              filename: null
              mediatype: null
              metadata: null
              raw: null
              text: 'Echo: I need help with my project'
              url: null
        referencetaskids: []
        role: ROLE_AGENT
        taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
    state: TASK_STATE_COMPLETED
    timestamp: 2026-10-04T15:01:41.256127955Z
```

#### View conversation history

Unlike `tasks list`, `tasks history` always returns the full task objects including their
conversation history:

```bash
$ a2a tasks history demo-ctx

context_id: demo-ctx
tasks:
    - artifacts: []
      contextid: demo-ctx
      history:
        - contextid: demo-ctx
          extensions: []
          messageid: msg-1791126101
          metadata: null
          parts:
            - data: null
              filename: null
              mediatype: null
              metadata: null
              raw: null
              text: I need help with my project
              url: null
          referencetaskids: []
          role: ROLE_USER
          taskid: null
        - contextid: demo-ctx
          extensions: []
          messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
          metadata: null
          parts:
            - data: null
              filename: null
              mediatype: null
              metadata: null
              raw: null
              text: 'Echo: I need help with my project'
              url: null
          referencetaskids: []
          role: ROLE_AGENT
          taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
      id: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
      metadata: null
      status:
        message:
            contextid: demo-ctx
            extensions: []
            messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
            metadata: null
            parts:
                - data: null
                  filename: null
                  mediatype: null
                  metadata: null
                  raw: null
                  text: 'Echo: I need help with my project'
                  url: null
            referencetaskids: []
            role: ROLE_AGENT
            taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
        state: TASK_STATE_COMPLETED
        timestamp: 2026-10-04T15:01:41.256127955Z
```

#### Interactive chat mode

Start a chat session to converse with the agent directly from your terminal. By default messages
are exchanged in **streaming** mode, where the agent's reply is rendered in real time:

```bash
$ a2a interactive --server-url http://localhost:8080
```

```text
 A2A Chat  http://localhost:8080 · streaming · context 1a2b3c4d

You
  What's the weather like today?

My A2A Agent
  It's sunny with a high of 24°C.

ready
> ▏
enter: send · ctrl+t: toggle mode · ctrl+c: quit
```

Use **background** mode for long-running tasks. Each message is submitted as a task and polled
until it reaches a terminal state:

```bash
$ a2a interactive --background
```

Key bindings:

- `Enter` — send the current message
- `Ctrl+T` — toggle between streaming and background mode mid-session
- `Ctrl+C` / `Esc` — quit

The session keeps a single context ID so the whole conversation is threaded. When the agent
responds with `input-required`, your next message automatically continues the same task. You can
also resume a previous conversation with `--context-id <id>`.

#### Output Formats

By default, all commands output structured data in YAML format. You can switch to JSON using the `-o` flag.

JSON carries the A2A v1.0.1 field names (`contextId`, `messageId`) and omits unset fields. YAML is
produced from the same Go structs, which carry no YAML tags, so keys are lowercased Go field names
(`contextid`, `messageid`) and unset fields appear as `null`:

```bash
# YAML output (default)
$ a2a tasks list --limit 2
showing: 1
tasks:
    - artifacts: []
      contextid: demo-ctx
      history: []
      id: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
      metadata: null
      status:
        message:
            contextid: demo-ctx
            extensions: []
            messageid: 06e1a21c-0af7-4edd-9de8-b6f369aa820c
            metadata: null
            parts:
                - data: null
                  filename: null
                  mediatype: null
                  metadata: null
                  raw: null
                  text: 'Echo: I need help with my project'
                  url: null
            referencetaskids: []
            role: ROLE_AGENT
            taskid: a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a
        state: TASK_STATE_COMPLETED
        timestamp: 2026-10-04T15:01:41.256127955Z
total: 1

# JSON output
$ a2a tasks list --limit 2 -o json
{
  "showing": 1,
  "tasks": [
    {
      "contextId": "demo-ctx",
      "id": "a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a",
      "status": {
        "message": {
          "contextId": "demo-ctx",
          "messageId": "06e1a21c-0af7-4edd-9de8-b6f369aa820c",
          "parts": [
            {
              "text": "Echo: I need help with my project"
            }
          ],
          "role": "ROLE_AGENT",
          "taskId": "a7fb8748-bd60-4f13-bb28-9dd43e2ecb3a"
        },
        "state": "TASK_STATE_COMPLETED",
        "timestamp": "2026-10-04T15:01:41.256127955Z"
      }
    }
  ],
  "total": 1
}
```

Note that `tasks list` strips artifacts and history unless `--include-artifacts` /
`--include-history` is passed.

## 🛠️ Development

### Prerequisites

- Go 1.26 or later
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

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
