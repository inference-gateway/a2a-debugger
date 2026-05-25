# AGENTS.md

Context and conventions for coding agents working on the A2A Debugger.

## Project Overview

A CLI tool for debugging A2A (Agent-to-Agent) servers. Built with Go 1.26, Cobra (CLI), Viper (config), and `go.uber.org/zap` (logging). A2A protocol types come from `github.com/inference-gateway/adk/types` -- never generate them locally, just bump the dependency.

## Build & Test

```bash
task build         # Full build with ldflags (version/commit/date) -> dist/a2a
task build:dev     # Quick dev build, no ldflags
task test          # go test ./...
task test:coverage # go test -cover ./...
task lint          # golangci-lint run
task tidy          # find . -name go.mod -execdir go mod tidy \;
task clean         # rm -rf dist/
```

Single test: `go test ./cli -run TestSpecificFunction`

## Code Architecture

- **All CLI logic lives in `cli/cli.go`** -- one file, ~840 lines. Tests in `cli/cli_test.go` and `cli/cli_output_test.go`.
- **`main.go` is a shim**: passes ldflags-injected `version`/`commit`/`date` to `cli.Execute()`.
- **Namespace commands**: `a2a config {set,get,list}`, `a2a tasks {list,get,history,submit,submit-streaming}`, plus `connect`, `agent-card`, `version`.

## Conventions & Patterns

- **`ensureA2AClient()`** -- lazy-init the A2A client before making API calls. Always call this.
- **`handleA2AError(err, method)`** -- wraps A2A errors into user-friendly messages. Always use for A2A call errors.
- **`printFormatted(data)` / `formatOutput(data)`** -- output in YAML (default) or JSON (`--output json` / `-o json`). Use for all structured output.
- **`init()` function** registers all subcommands and flags. Add new commands/flags here.
- **Global flags** on `rootCmd.PersistentFlags()`: `--server-url`, `--timeout`, `--debug`, `--insecure`, `--output`.
- **Command-specific flags** set on the command itself (e.g., `--state`, `--limit`, `--context-id`).

## Config

Default: `~/.a2a.yaml`. Viper loads env vars automatically. Config keys: `server-url`, `timeout`, `debug`, `insecure`, `output`.

## Testing

- Tests use `mockA2AClient` (implements the `client.A2AClient` interface). Add mock methods as needed.
- **End-to-end harness**: `example/docker-compose.yml` boots `ghcr.io/inference-gateway/mock-agent` alongside the debugger. Use for validating real A2A behavior.

## CI Pipeline

Order: `task tidy` -> dirty check -> `task lint` -> `task build` -> `task test`. Keep this order when mimicking CI.

## Docker

- `Dockerfile` -- multi-stage build (golang:1.26.2-alpine -> alpine), binary at `/a2a`.
- `Dockerfile.goreleaser` -- UPX-compressed, distroless base, nonroot user.

## Security

- `CGO_ENABLED=0` in production builds.
- `--insecure` flag skips TLS verification -- warn against production use.
- No secrets in config; credentials come from env vars (Viper `AutomaticEnv`).
