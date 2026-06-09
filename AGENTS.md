# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the application entrypoint and imports plugins for registration. Feature code lives under `plugin/<name>/`, with one Go package per plugin, such as `plugin/rsshub` or `plugin/minecraftobserver`. Shared runtime assets, databases, fonts, and bundled data live in `data/`; keep plugin-specific static resources in `data/<Name>` when they must be versioned. Build and resource helpers live in `winres/`, with additional utilities in directories such as `abineundo/`, `custom/`, and `kanban/`. Place tests beside implementation files as `*_test.go`.

## Build, Test, and Development Commands
- `go mod tidy`: sync module dependencies and clean unused requirements.
- `go generate main.go`: regenerate runtime or embedded assets used by the main build.
- `go generate ./...`: regenerate all package assets after broad resource changes.
- `go run -ldflags "-s -w" main.go`: run the bot locally.
- `go build -trimpath -ldflags "-s -w" .`: build a release-style binary.
- `go test ./...`: run all unit tests.
- `golangci-lint run`: run the lint gate used by CI.

The quick-start scripts `run.sh` and `run.bat` follow the same flow: tidy, generate, run.

## Coding Style & Naming Conventions
Use standard Go formatting with `gofmt` and `goimports`; the local import prefix is `github.com/FloatTech/ZeroBot-Plugin`. Package and directory names should be short lowercase identifiers. Use `CamelCase` for exported names and `camelCase` for internal names. Prefer descriptive lowercase file names such as `model.go`, `store.go`, and `*_test.go`. Lint forbids direct `fmt.Errorf`; use `github.com/pkg/errors` helpers such as `errors.Wrap` or `errors.Errorf` where wrapping is needed.

## Testing Guidelines
Use Go's standard `testing` package. Prefer table-driven tests for parsers, storage logic, command helpers, and deterministic plugin behavior. Run targeted tests during iteration, for example `go test ./plugin/rsshub/...`, and run `go test ./...` before opening a PR.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit style: `feat(scope): ...`, `fix(scope): ...`, `chore(lint): ...`, `chore(nix): ...`, and `doc(README): ...`. Use plugin names or areas such as `main`, `ci`, or `nix` as scopes.

PRs should include a clear behavior summary, affected plugin(s), linked issue when available, and validation evidence such as `go test ./...` or `golangci-lint run`. Include screenshots or log snippets for user-visible output changes. Do not include `.go` in the PR title; the CI workflow may close those PRs automatically.
