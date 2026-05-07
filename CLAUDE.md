# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A2A Debugger is a CLI tool for debugging and monitoring A2A (Agent-to-Agent) servers. It's built with Go and uses the Cobra framework for command-line interface, with Viper for configuration management.

**Core Dependencies:**
- `github.com/inference-gateway/adk` - The A2A Agent Development Kit client library
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `go.uber.org/zap` - Structured logging

## Architecture

### A2A Types

A2A protocol types come from `github.com/inference-gateway/adk/types` (generated upstream from the A2A JSON-RPC schema). To pick up schema changes, bump the `adk` dependency rather than running a local generator — there is no local generation in this repo.

### CLI Command Structure

All CLI commands are implemented in a **single file**: `cli/cli.go` (~806 lines). The command structure is namespace-based:

- **Config namespace** (`a2a config`): Configuration management (set, get, list)
- **Tasks namespace** (`a2a tasks`): Task operations (list, get, history, submit, submit-streaming)
- **Server commands**: `connect`, `agent-card`, `version`

Key architectural patterns:
- `ensureA2AClient()` - Lazy initialization of A2A client
- `handleA2AError()` - Centralized error handling for method-not-found errors
- `formatOutput()` / `printFormatted()` - Support for YAML (default) and JSON output formats

### Entry Point

`main.go` is minimal - it just passes version info to `cli.Execute()`.

## Common Commands

### Development Workflow

```bash
# Run linting
task lint

# Build with version info
task build

# Quick dev build (no version info)
task build:dev

# Run tests
task test

# Run tests with coverage
task test:coverage

# Clean build artifacts
task clean

# Tidy dependencies
task tidy
```

### Testing the CLI

```bash
# Run the built binary
./dist/a2a --help

# Test with local A2A server (example environment)
cd example
docker compose up -d
docker compose run --rm a2a-debugger connect
```

### Installation

```bash
# Install locally
task install

# Uninstall
task uninstall
```

## Configuration

- Default config location: `~/.a2a.yaml`
- Environment variables are auto-loaded via Viper
- All flags can be set in config file or via CLI flags

Config file format:
```yaml
server-url: http://localhost:8080
timeout: 30s
debug: false
insecure: false
output: yaml  # or json
```

## CI/CD Pipeline

The CI workflow (`task tidy` → dirty check → `task lint` → `task build` → `task test`) ensures:
1. Dependencies are tidy
2. No uncommitted changes
3. Code passes linting
4. Build succeeds
5. Tests pass

## Adding New Commands

When adding commands to `cli/cli.go`:
1. Define the command with Cobra
2. Use `ensureA2AClient()` before making A2A calls
3. Use `handleA2AError()` for error handling
4. Use `printFormatted()` for structured output
5. Register the command in `init()` function
6. Add flags if needed (global flags are on `rootCmd`, command-specific on the command itself)

## Running Tests

Single test: `go test ./cli -run TestSpecificFunction`
All tests: `task test`
Coverage: `task test:coverage`

## Docker Build

```bash
# Build Docker image with version info
task build:docker
```
