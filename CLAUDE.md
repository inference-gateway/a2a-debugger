# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A2A Debugger is a command-line debugging tool for A2A (Agent-to-Agent) servers. It's built in Go using the Cobra CLI framework and integrates with the Inference Gateway A2A ecosystem for testing connections, managing tasks, viewing conversation histories, and monitoring streaming responses.

## Development Commands

### Core Development Workflow

```bash
# Generate Go code from A2A schema (always run before commit)
task generate

# Run linting (required before commit)
task lint

# Build the project
task build

# Build for development (no version info)
task build:dev

# Run tests
task test

# Run tests with coverage
task test:coverage

# Clean build artifacts
task clean
```

### Schema and Dependencies

```bash
# Download latest A2A schema
task a2a:download:schema

# Clean up Go module dependencies
task tidy
```

### Installation

```bash
# Install locally
task install

# Uninstall
task uninstall
```

### Docker

```bash
# Build Docker image with version info
task build:docker
```

## Code Architecture

### Project Structure

- `main.go`: Entry point, passes version info to CLI
- `cli/cli.go`: Main CLI implementation using Cobra framework
- `a2a/`: A2A protocol types and schema
  - `generated_types.go`: Auto-generated from A2A JSON-RPC schema
  - `schema.yaml`: Official A2A schema (downloaded from inference-gateway/schemas)

### Key Components

#### CLI Architecture (cli/cli.go)

- **Root Command**: Main `a2a` command with global flags
- **Namespace Commands**:
  - `config`: Configuration management (set/get/list)
  - `tasks`: Task operations (list/get/history/submit/submit-streaming)
  - Standalone: `connect`, `agent-card`, `version`

#### Configuration System

- Uses Viper for configuration management
- Config file: `~/.a2a.yaml`
- Environment variable support with automatic binding
- Global flags: `--server-url`, `--timeout`, `--debug`, `--insecure`, `--config`

#### A2A Client Integration

- Uses `github.com/inference-gateway/a2a/adk/client` for server communication
- Lazy initialization via `ensureA2AClient()`
- Error handling with user-friendly messages for MethodNotFoundError (-32601)
- Structured logging with Zap

#### Task Management

- **List Tasks**: Filter by state, context-id, with pagination
- **Get Task**: Detailed task info with optional history
- **Submit Task**: Send messages to A2A server
- **Submit Streaming**: Real-time streaming with event processing
- **History**: Conversation history by context ID

#### Streaming Support

- Event channel-based streaming architecture
- Event types: `status-update`, `artifact-update`
- Raw mode for debugging protocol compliance
- Graceful handling of streaming responses

### Dependencies

- **CLI Framework**: `github.com/spf13/cobra` + `github.com/spf13/viper`
- **A2A SDK**: `github.com/inference-gateway/a2a`
- **Logging**: `go.uber.org/zap`
- **Go Version**: 1.24+

## Development Guidelines

### Code Generation

- Always run `task generate` before committing to update generated types
- Generated files have `generated_` prefix - never modify manually
- Schema updates require downloading latest from inference-gateway/schemas

### Testing Strategy

- Use table-driven testing patterns
- Each test case should have isolated mock servers
- Early returns over deep nesting
- Prefer switch statements over if-else chains

### Error Handling

- Custom error handling for A2A protocol errors
- User-friendly messages for common errors (e.g., MethodNotFoundError)
- Structured logging for debugging

### Type Safety

- Code to interfaces for easier mocking
- Strong typing over dynamic typing
- Use generated types from A2A schema

### Pre-commit Workflow

1. `task generate` - Update generated files
2. `task lint` - Code quality checks
3. `task build` - Verify compilation
4. `task test` - Ensure tests pass

## Configuration

### Default Configuration

```yaml
server-url: http://localhost:8080
timeout: 30s
debug: false
insecure: false
```

### Environment Variables

All configuration keys can be set via environment variables (automatic Viper binding).

## Related Repositories

- [Inference Gateway](https://github.com/inference-gateway)
- [A2A ADK](https://github.com/inference-gateway/a2a) - Agent Development Kit
- [Go SDK](https://github.com/inference-gateway/go-sdk)
- [Schemas](https://github.com/inference-gateway/schemas) - A2A protocol schemas
