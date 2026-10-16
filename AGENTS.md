# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go module for resolving Minecraft crafting materials. Core recipe models and parsing/resolution logic live in `internal/`; the command-line entry point is `cmd/cli/main.go`. Input data is stored under `assets/recipes/` and `assets/item-tag-groups/` as JSON. Tests sit beside the package code in `internal/*_test.go`. Keep new parsing and domain logic in `internal`, and keep CLI argument and file handling in `cmd/cli`.

## Build, Test, and Development Commands

- `go test ./...` runs all Go tests with the standard toolchain.
- `just test` runs the suite with `gotestsum --format short ./...`; run it directly without extra cache arguments.
- `just test-watch` reruns tests as files change and also requires `gotestsum`.
- `go run ./cmd/cli --recipes assets/recipes/` runs the CLI against the bundled recipes. Use `--recipes <dir>` to point it at another directory.
- `gofmt -w internal cmd` formats Go source files.

## Agent Tooling

Use the Jujutsu MCP for repository history and changes; do not run `jj` through the shell or change its configuration to work around a command error. When discovering MCP capabilities, inspect only the specific tool needed rather than dumping the full tool catalog.

## Coding Style & Naming Conventions

Use standard Go formatting (`gofmt`), tabs for indentation, and idiomatic exported `PascalCase` and unexported `camelCase` names. Keep package names short and lowercase. Wrap errors with context when returning them, and preserve sentinel errors such as `ErrUnsupportedRecipeType` where callers inspect error types. Recipe data files use lowercase Minecraft IDs with underscores and a `.json` suffix.

## Testing Guidelines

Tests use Go's built-in `testing` package, with `github.com/google/go-cmp/cmp` for comparisons. Add or update tests in the relevant `internal/*_test.go` file; name test functions `Test<Behavior>`. Cover supported recipe formats, malformed or unsupported JSON, and material resolution edge cases. Run `go test ./...` before submitting changes.

## Commit & Pull Request Guidelines

Recent commits use short Conventional Commit subjects such as `feat: add cli`, `fix: ...`, and `refactor: ...`. Follow that pattern and keep each commit focused. A pull request should explain the behavior change, mention relevant issues, and include test results. Include sample JSON or CLI input and output when changing parsing or user-visible behavior.

## Data and Configuration

Treat files in `assets/` as runtime inputs: preserve valid JSON and Minecraft namespaced identifiers. Keep local environment configuration in the existing `devenv.nix`, `devenv.yaml`, and `justfile` conventions; do not commit machine-specific paths or generated output.
