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

### Code Generation

The `a2a/generated_types.go` file is **auto-generated** from the A2A JSON-RPC schema and should **NEVER** be edited manually. To regenerate:

1. Download latest schema: `task a2a:download:schema`
2. Generate types: `task generate`

The generator reads `a2a/schema.yaml` and produces Go types for the A2A protocol.

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
# Generate code from schema (required after schema updates)
task generate

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

The CI workflow (`task generate` → `task tidy` → dirty check → `task lint` → `task build` → `task test`) ensures:
1. Generated code is up-to-date
2. Dependencies are tidy
3. No uncommitted changes
4. Code passes linting
5. Build succeeds
6. Tests pass

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
