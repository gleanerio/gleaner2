# CLAUDE.md - Development Guidelines for Nabu/glcon

## Project Overview

Nabu (glcon) is a unified CLI tool that combines Gleaner (data harvesting) and Nabu (graph loading) for the GleanerIO ecosystem. It reads JSON-LD from websites, stores it in S3-compatible object stores, converts to RDF, and loads into SPARQL triplestores.

## Build & Test

```bash
# Build
go build ./cmd/nabu/

# Run tests
go test ./...

# Run specific package tests
go test ./pkg/graph/
go test ./internal/common/

# Test CLI runs without crashing
./nabu --help
./nabu prefix --help
./nabu config init --help
```

## Code Review Checklist

Before committing changes, verify:

### 1. Build and CLI Smoke Test
- [ ] `go build ./cmd/nabu/` succeeds with no errors
- [ ] `./nabu --help` runs without crashing (no config required)
- [ ] `./nabu <subcommand> --help` runs without crashing for each changed subcommand
- [ ] `go vet ./...` reports no issues

### 2. Tests
- [ ] `go test ./...` — all tests that were passing before still pass
- [ ] New code has tests where practical
- [ ] Known pre-existing failures (internal/common missing test data, internal/summoner/acquire test logic) are not made worse

### 3. Logrus Logging
- [ ] Use `log.Errorf` / `log.Infof` / `log.Warnf` / `log.Fatalf` / `log.Tracef` / `log.Debugf` when format verbs (%s, %v, %d, etc.) are used
- [ ] Do NOT use `log.Error`, `log.Info`, `log.Fatal`, etc. with format strings — these pass args as fields, not format parameters
- [ ] Correct: `log.Errorf("failed: %v", err)` / Wrong: `log.Error("failed: %v", err)`

### 4. CLI Command Safety
- [ ] Every command that uses `viperVal` or `mc` calls `requireConfig()` at the start of its Run function
- [ ] Commands that don't need config (like `config init`) must NOT call `requireConfig()`
- [ ] Flag defaults must be empty string `""` unless a meaningful default exists — do not use path-like defaults for URL flags

### 5. Config Package Completeness
- [ ] If adding new code to `internal/` that references `config.*` or `configTypes.*`, verify the referenced function/type exists in `pkg/config/`
- [ ] The `pkg/config/` package is the single source of truth — `gleaner/internal/config/` is the original source but `internal/` imports from `pkg/config/`

## Architecture Notes

- **Entry point**: `cmd/nabu/main.go`
- **CLI commands**: `pkg/cli/` — each file defines a cobra command
- **Config**: `pkg/config/` — shared config types and readers
- **Shared libraries**: `pkg/graph/`, `pkg/storage/` — reusable across gleaner and nabu
- **Core logic**: `internal/` — summoner (harvesting), millers (processing), objects (loading), prune, sparql, services
- **Gleaner source**: `gleaner/` — original gleaner code (subtree merge, reference only)

## Common Patterns

### Adding a new CLI command
1. Create `pkg/cli/mycommand.go`
2. Define cobra command with `requireConfig()` in Run if it needs config
3. Register with `rootCmd.AddCommand()` in `init()`

### Config access
```go
// Get sources
sources, err := config.GetSources(viperVal)
// Get bucket name
bucket, err := config.GetBucketName(viperVal)
// Get SPARQL endpoint
ep := viperVal.GetString("flags.endpoint")
endpoint, err := config.GetEndpoint(viperVal, ep, "bulk")
```
