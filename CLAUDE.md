# CLAUDE.md - Development Guidelines for glcon

## Project Overview

glcon is a **single unified CLI tool** that merges the former Gleaner (data harvesting) and Nabu (graph loading) projects. It reads JSON-LD from websites, stores it in S3-compatible object stores, converts to RDF, and loads into SPARQL triplestores.

**Key design goal:** Configuration is split into separate files so that secrets (service credentials) are never mixed with shareable data (source lists, run settings). See [Configuration Architecture](#configuration-architecture) below.

Reference: https://github.com/gleanerio/gleaner/wiki/Discussion-for-merging-Gleaner-and-Nabu

## Build & Test

```bash
# Build
go build -o nabu ./cmd/nabu/

# Run stable tests (config + graph packages)
make test

# Run all tests (some pre-existing failures in internal/)
make test-all

# Static analysis
make vet

# CLI smoke test — verifies every subcommand's --help works
make smoke

# Full pre-commit check: build + vet + tests + smoke
make check

# Validate a config directory
make validate-config CFG_DIR=configs/local

# Run specific package tests directly
go test -v ./pkg/config/
go test -v ./pkg/graph/
```

## Configuration Architecture

Configuration is split into three files to separate concerns:

```
configs/<name>/
├── services.yaml    # PRIVATE: MinIO keys, SPARQL credentials, headless URL
├── sources.yaml     # SHAREABLE: data sources, context maps, org metadata
└── glcon.yaml       # SHAREABLE: run settings, object prefixes, miller config
```

**Why three files?**
- `services.yaml` contains secrets (S3 keys, SPARQL passwords) — keep private, use env vars in CI
- `sources.yaml` contains the list of data sources — safe to share, version, or generate from CSV
- `glcon.yaml` controls what glcon does (summon, mill, load prefixes) — safe to share

See `configs/template/` for annotated examples of each file.

### Services config (`services.yaml`)
Holds connection details that should NOT be shared publicly:
```yaml
minio:
  address: localhost
  port: 9000
  ssl: false
  accessKey: ""       # or MINIO_ACCESS_KEY env var
  secretKey: ""       # or MINIO_SECRET_KEY env var
  bucket: gleaner
endpoints:
  - service: blazegraph
    baseurl: http://localhost:9999/blazegraph/namespace/kb
    type: blazegraph
    authenticate: false
    modes:
      - action: bulk
        suffix: /sparql
        accept: text/x-nquads
        method: POST
summoner:
  mode: full
  threads: 5
  headless: http://127.0.0.1:9222
```

### Sources config (`sources.yaml`)
Data sources to harvest — safe to share and version:
```yaml
sources:
  - sourcetype: sitemap
    name: samplesearth
    url: https://samples.earth/sitemap.xml
    active: true
    domain: https://samples.earth
context:
  cache: true
contextmaps:
  - prefix: "https://schema.org/"
    file: "./assets/schemaorg-current-https.jsonld"
```

### Run config (`glcon.yaml`)
Controls what operations to perform:
```yaml
gleaner:
  runid: runX
  summon: true
  mill: true
millers:
  graph: true
objects:
  domain: us-east-1
  prefix:
    - summoned/samplesearth
```

### Config generation workflow
The `config generate` command (from gleaner) reads base templates + `sources.csv` and produces merged config files:
1. `config init --cfgName local` — copies templates to `configs/local/`
2. Edit `configs/local/services.yaml` (add your MinIO/SPARQL credentials)
3. Edit `configs/local/sources.csv` (list your data sources)
4. `config generate --cfgName local` — generates `gleaner.yaml`, `nabu.yaml`, `nabu_prov.yaml`

### Config types (pkg/config/)
| Type | File | Purpose |
|------|------|---------|
| `ServicesConfig` | `services.go` | MinIO + Endpoints + Summoner (secrets) |
| `Sources` | `sources.go` | Individual data source definition |
| `SourcesConfig` | `sources.go` | Sources + Objects + ImplNetwork + Context |
| `EndPoint` | `endpoints.go` | SPARQL service with multiple modes |
| `ServiceMode` | `endpoints.go` | Resolved endpoint (baseurl + suffix) |
| `Minio` | `minio.go` | S3-compatible object store connection |
| `Sparql` | `sparql.go` | Legacy flat SPARQL config (deprecated, use EndPoint) |
| `Objects` | `objects.go` | Bucket + domain + prefix list |
| `Summoner` | `summoner.go` | Harvesting behavior settings |

## CLI Commands

glcon provides a unified command set:

| Command | Description | Needs Config? |
|---------|-------------|---------------|
| `summon` | Harvest JSON-LD from data sources | Yes |
| `mill` | Process harvested data through millers | Yes |
| `bulk` | Bulk load RDF to SPARQL endpoints | Yes |
| `prefix` | Load graphs by prefix to triplestore | Yes |
| `release` | Create release graphs | Yes |
| `prune` | Remove orphaned graphs | Yes |
| `graph clear` | Clear all graphs from triplestore | Yes |
| `graph drop <name>` | Drop a specific named graph | Yes |
| `config init` | Initialize config directory | No |

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
- **Config types & readers**: `pkg/config/` — shared across all operations
- **Shared libraries**: `pkg/graph/`, `pkg/storage/` — JSON-LD/RDF conversion, S3 operations
- **Core logic**: `internal/` — summoner (harvesting), millers (processing), objects (loading), prune, sparql, services
- **Config templates**: `configs/template/` — annotated example configs
- **Gleaner source**: `gleaner/` — original gleaner code (subtree merge, reference for config generation)

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
// Get SPARQL endpoint (uses new endpoints[] config, selected by --endpoint flag)
ep := viperVal.GetString("flags.endpoint")
endpoint, err := config.GetEndpoint(viperVal, ep, "bulk")
// Get MinIO settings
minioCfg := config.GetMinioConfig(viperVal)
// Get services config (for the separated services file)
serviceViper, err := config.ReadServicesConfig("services", "configs/local")
```
