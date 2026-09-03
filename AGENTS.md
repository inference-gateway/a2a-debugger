# AGENTS.md

`a2a` is a Cobra-based CLI (binary name: `a2a`) for inspecting and exercising A2A (Agent-to-Agent) servers. It is a thin client around `github.com/inference-gateway/adk/client` — most commands are JSON-RPC calls (`tasks/list`, `tasks/get`, `message/send`, `message/stream`, `agent/card`) whose responses are rendered as YAML (default) or JSON.

## Build, test, lint

Build automation goes through [Task](https://taskfile.dev/), not `make`:

| Task | Purpose |
| --- | --- |
| `task build` | Builds `dist/a2a` with version/commit/date via `-ldflags -X main.{version,commit,date}` |
| `task build:dev` | Plain `go build` — faster iteration |
| `task test` | `go test ./...` |
| `task test:coverage` | `go test -cover ./...` |
| `task lint` | `golangci-lint run` (CI pins v2.12.2) |
| `task tidy` | `go mod tidy` for every module — run before pushing; CI fails on a dirty `go.mod` |

Run one test: `go test ./cli -run TestSubmitStreamingTaskCmd_RawMode -v`. Try the binary end-to-end against a real server via `example/docker-compose.yml` (spins up `mock-agent` + the debugger image on a bridge network — no API keys needed).

## Architecture

**Single-package CLI.** All commands, flag wiring, viper bindings, output helpers, and the JSON-RPC error normalizer live in `cli/cli.go`; `main.go` is just `cli.Execute(version, commit, date)`. Add a command by declaring `var fooCmd = &cobra.Command{...}` in `cli.go`, registering it in `init()` under `tasksCmd` or `rootCmd`, and wiring flags there. Namespaces: `config` (set/get/list, viper-backed) and `tasks` (list/get/history/submit/submit-streaming); `connect`, `agent-card`, and `version` sit on root.

**Lazy A2A client.** The `a2aClient` package-global stays `nil` until `ensureA2AClient()` is called inside a command's `RunE`. Never call `initA2AClient()` at package init — it needs viper config and the logger.

**Output is centralized.** Always render through `printFormatted(data)` so `--output yaml|json` is honored; never `fmt.Println(yaml.Marshal(...))`. In `submit-streaming`, freeform progress prints directly and only the final summary goes through formatted output.

**JSON-RPC errors.** Wrap every `a2aClient.*` error with `handleA2AError(err, methodName)` so code `-32601` becomes "Method not implemented by the agent". Streaming events are heuristically typed in `submit-streaming` by probing for `artifact`/`final`/`id` keys — update that switch when ADK adds event kinds.

## Testing

Tests in `cli/cli_test.go` swap the `a2aClient` package-global with a `mockA2AClient` satisfying `client.A2AClient`. Always save and restore the original (`originalClient := a2aClient` … `a2aClient = originalClient`), and capture stdout via the `os.Pipe()` swap of `os.Stdout`.

## Conventions

- **Conventional commits are load-bearing.** `.releaserc.yaml` drives semantic-release: `feat` → minor; `fix|impr|refactor|perf|ci|docs|style|test|build|security|chore` → patch; breaking → major. Capitalized descriptions (`feat(client): Add retry mechanism`). `chore(release):` is reserved for the release bot.
- **Go version is pinned in `go.mod` (1.26.7)**; CI reads it via `go-version-file`.
- **Config is `~/.a2a.yaml`**, loaded by viper; `viper.AutomaticEnv()` means `SERVER_URL=...` overrides `server-url`.
- **Security:** never commit local config, credentials, or private server URLs. Use `--insecure` only for local or test servers.

## Release

Manual: trigger `.github/workflows/release.yml` (workflow_dispatch). It runs semantic-release against `main` (or `rc/*`) to tag and create the GitHub release; that event fires `artifacts.yml`, where goreleaser builds linux/darwin × amd64/arm/arm64 binaries, pushes multi-arch Docker images to `ghcr.io/inference-gateway/a2a-debugger`, and cosign-signs. `release: disable: true` in `.goreleaser.yaml` is intentional — semantic-release owns the release.
