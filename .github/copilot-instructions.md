# Custom Instructions for Copilot

Today is June 17, 2025.

- Never modify files with generated\_ prefix: These files are automatically generated and should not be edited manually.
- Always use context7 to check for the latest updates, features, or best practices of a library relevant to the task at hand.
- Always prefer Table-Driven Testing: When writing tests.
- Always use Early Returns: Favor early returns to simplify logic and avoid deep nesting with if-else structures.
- Always prefer switch statements over if-else chains: Use switch statements for cleaner and more readable code when checking multiple conditions.
- Always run `task generate` before committing code to ensure all generated files are up-to-date.
- Always run `task lint` before committing code to ensure it adheres to the project's linting rules.
- Always run `task build` to verify compilation after making changes.
- Always run `task test` before committing code to ensure all tests pass.
- Always search for the simplest solution first before considering more complex alternatives.
- Always prefer type safety over dynamic typing: Use strong typing and interfaces to ensure type safety and reduce runtime errors.
- When possible code to an interface so it's easier to mock in tests.
- When writing tests, each test case should have it's own isolated mock server mock dependecies so it's easier to understand and maintain.

## Development Workflow

### Configuration Changes

When adding new configuration fields:

2. Run `task generate` to regenerate all configuration-related files
3. Run `task lint` to ensure code quality
4. Run `task build` to verify successful compilation
5. Run `task test` to ensure all tests pass

## Available Tools and MCPs

- context7 - Helps by finding the latest updates, features, or best practices of a library relevant to the task at hand.

## Related Repositories

- [Inference Gateway](https://github.com/inference-gateway)
  - [Inference Gateway UI](https://github.com/inference-gateway/ui)
  - [Go SDK](https://github.com/inference-gateway/go-sdk)
  - [Rust SDK](https://github.com/inference-gateway/rust-sdk)
  - [TypeScript SDK](https://github.com/inference-gateway/typescript-sdk)
  - [Python SDK](https://github.com/inference-gateway/python-sdk)
  - [A2A ADK](https://github.com/inference-gateway/a2a)
  - [Documentation](https://github.com/inference-gateway/docs)
