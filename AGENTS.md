# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go CLI for debugging A2A servers. `main.go` only starts the CLI through `cli.Execute`; most command definitions, flag wiring, output helpers, and A2A client handling live in `cli/cli.go`. Tests are colocated in `cli/*_test.go`. Build and release assets are at the root, including `Taskfile.yaml`, `Dockerfile`, `Dockerfile.goreleaser`, `install.sh`, and `.releaserc.yaml`. Example runtime setup is in `example/`, including a Docker Compose flow for exercising the debugger against a mock agent.

## Build, Test, and Development Commands

Use Task for standard workflows:

- `task build`: build `dist/a2a` with version, commit, and date injected.
- `task build:dev`: build `dist/a2a` without release metadata for fast local iteration.
- `task test`: run `go test ./...`.
- `task test:coverage`: run `go test -cover ./...`.
- `task lint`: run `golangci-lint run`.
- `task tidy`: run `go mod tidy` for every module.

For a focused test, use `go test ./cli -run TestName -v`.

## Coding Style & Naming Conventions

Follow normal Go style: `gofmt` formatting, tabs for indentation, short package names, and exported identifiers only when needed outside their package. Keep CLI behavior centralized in `cli/cli.go` unless a larger split is clearly justified. Add Cobra commands as `*cobra.Command` values, register them in `init()`, and bind flags near command registration. Respect `--output yaml|json` by using `printFormatted(data)` for structured output.

## Testing Guidelines

Use Go's standard `testing` package. Name tests `TestXxx` and keep them hermetic. Existing tests replace the package-global `a2aClient` with a mock and restore it afterward; follow that pattern for command behavior. Capture stdout with the existing pipe-based approach when asserting CLI output. Run `task test` before opening a PR, and add coverage for new commands, flags, and output modes.

## Commit & Pull Request Guidelines

Use Conventional Commits; semantic-release reads these for versioning. Examples: `feat(tasks): Add cancel command`, `fix(cli): Handle missing task id`, `docs: Update examples`. Recent history uses lowercase types such as `chore`, `ci`, and `fix`; reserve `chore(release)` for automation. PRs should describe the change, list tests run, link related issues, and include terminal output or screenshots when CLI behavior changes.

## Security & Configuration Tips

The default config file is `~/.a2a.yaml`; environment variables are read by Viper. Avoid committing local config, credentials, or server URLs that are not meant for public examples. Use `--insecure` only for local or test servers.
