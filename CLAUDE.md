# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`a2a-debugger` is a Cobra-based CLI (binary name: `a2a`) for inspecting and exercising A2A (Agent-to-Agent) servers. It is a thin client around `github.com/inference-gateway/adk/client` — most "commands" are JSON-RPC calls (`tasks/list`, `tasks/get`, `message/send`, `message/stream`, `agent/card`) whose responses are unmarshaled into `adk.*` types and rendered as YAML (default) or JSON.

## Commands

Build automation goes through [Task](https://taskfile.dev/), not `make`:

| Task              | What it does                                                                                  |
| ----------------- | --------------------------------------------------------------------------------------------- |
| `task build`      | Builds `dist/a2a` with version/commit/date injected via `-ldflags -X main.{version,commit,date}` |
| `task build:dev`  | Plain `go build` (no ldflags) — faster iteration                                              |
| `task test`       | `go test ./...`                                                                               |
| `task lint`       | `golangci-lint run` (CI pins v2.12.2)                                                         |
| `task tidy`       | `find . -name go.mod -execdir go mod tidy \;` — run before pushing (see CI note below)        |
| `task install`    | `go install .` to `$GOPATH/bin`                                                               |
| `task build:docker` | Builds the multi-arch tagged image locally                                                  |

Run a single test: `go test ./cli -run TestSubmitStreamingTaskCmd_RawMode -v`

Try the binary against a real server end-to-end via `example/docker-compose.yml`, which spins up `ghcr.io/inference-gateway/mock-agent` plus the debugger image on a shared bridge network — no API keys needed.

## Architecture

**Single-package CLI.** All commands, flag wiring, viper bindings, output helpers, and the JSON-RPC error normalizer live in `cli/cli.go`. `main.go` is just `cli.Execute(version, commit, date)`. Adding a command means: declare a `var fooCmd = &cobra.Command{...}` in `cli.go`, register it in `init()` under the right namespace (`tasksCmd.AddCommand(...)` or `rootCmd.AddCommand(...)`), and wire its flags there too.

**Two namespaces under root**: `config` (set/get/list backed by viper) and `tasks` (list/get/history/submit/submit-streaming). `connect`, `agent-card`, and `version` sit directly on root.

**Lazy A2A client.** The `a2aClient` package-global is `nil` until `ensureA2AClient()` is called inside a command's `RunE`. Do not call `initA2AClient()` at package init — it depends on viper having loaded config and on the logger existing.

**Output is centralized.** Never call `fmt.Println(yaml.Marshal(...))` directly. Use `printFormatted(data)` so the user's `--output yaml|json` (default `yaml`) is respected. For mixed-content commands (e.g. `submit-streaming`) where humans want narrated progress AND structured summary, the streaming command prints freeform text and only the final summary goes through formatted output — keep that split intentional.

**JSON-RPC error handling.** Any error from `a2aClient.*` calls must be wrapped via `handleA2AError(err, methodName)` so that JSON-RPC code `-32601` (MethodNotFound) becomes the friendly `❌ Method 'X' not implemented by the agent` instead of leaking raw error strings.

**Streaming events are heuristically typed.** `SendTaskStreaming` returns generic `JSONRPCSuccessResponse`s; `submit-streaming` decides between `status-update`, `artifact-update`, and `task` snapshot by probing for the presence of `artifact`/`final`/`id` keys in the unmarshaled map. If you add new event kinds in ADK, update that switch.

**Test pattern.** Tests in `cli/cli_test.go` swap the `a2aClient` package-global with a `mockA2AClient` that satisfies `client.A2AClient`. Always save and restore the original (`originalClient := a2aClient` ... `defer`/end: `a2aClient = originalClient`). Output assertions capture stdout via an `os.Pipe()` swap of `os.Stdout` — follow the same pattern for new command tests so they remain hermetic.

## Conventions to honor

- **Conventional commits are load-bearing.** `.releaserc.yaml` drives semantic-release: `feat` → minor, `fix|impr|refactor|perf|ci|docs|style|test|build|security|chore` → patch, breaking changes → major. Use capitalized descriptions (`feat(client): Add retry mechanism`). `chore(release): ...` is reserved for the release bot.
- **CI fails on dirty `go.mod` after `task tidy`.** The CI job runs `task tidy` then `git diff --exit-code`. Always run `task tidy` locally before committing dependency changes.
- **Go version is pinned in `go.mod` (1.26.2).** CI uses `go-version-file: 'go.mod'` — bump `go.mod` if you need a newer toolchain.
- **The `a2a/generated_types.go` path is marked `linguist-generated`** in `.gitattributes` (the path is anticipatory — no such file exists today, but treat any future generated file as not-for-hand-editing).
- **Default config file is `~/.a2a.yaml`**, loaded by viper in `initConfig()`. Env vars are picked up via `viper.AutomaticEnv()` (so `SERVER_URL=...` overrides `server-url`).

## Release flow

Manual: trigger `.github/workflows/release.yml` (workflow_dispatch). It runs `semantic-release` against `main` (or `rc/*` for prereleases) to compute the next version, tag, and create a GitHub release. The release event then fires `artifacts.yml` which runs `goreleaser` to build cross-platform binaries (linux/darwin × amd64/arm/arm64), build & push multi-arch Docker images to `ghcr.io/inference-gateway/a2a-debugger`, and sign them with cosign. `release: disable: true` in `.goreleaser.yaml` is intentional — semantic-release owns the GitHub release; goreleaser only produces assets and images.
