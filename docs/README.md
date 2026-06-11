# gleaner - GleanerIO Unified Tool

## About

`gleaner` is the unified command-line tool that merges the former
**Gleaner** (data harvesting) and **Nabu** (graph loading) projects into a single binary.

It reads structured data (JSON-LD) from websites via sitemaps and APIs, stores it
in an S3-compatible object store (MinIO, AWS S3, Google Cloud Storage, Wasabi),
converts it to RDF, and loads it into any standards-compliant SPARQL triplestore.

Triplestore requirements:
* SPARQL 1.1 with Update support
* SPARQL 1.1 over HTTP

## Building

```bash
make build                       # builds the gleaner binary
go build -o gleaner ./cmd/gleaner/
```

## Quick Start

```bash
# 1. Initialize a configuration directory
gleaner config init myproject

# 2. Edit configs (see Configuration section below)
#    myproject/nabu.yaml      - service connections (MinIO, SPARQL)
#    myproject/gleaner.yaml   - data sources

# 3. Harvest JSON-LD from configured sources
gleaner summon --cfgPath myproject --cfgName gleaner

# 4. Process harvested data into RDF (N-Quads)
gleaner mill --cfgPath myproject --cfgName gleaner

# 5. Load RDF into triplestore
gleaner prefix --cfgPath myproject --cfgName local
```

## Configuration

gleaner requires a YAML configuration file. A template can be found in
[example.yaml](../config/example.yaml).

There are three ways to specify a config file:

```bash
# 1. Full path to a config file
gleaner <command> --cfg /path/to/config.yaml

# 2. Config directory + name (looks for <cfgPath>/<cfgName>/nabu.yaml)
gleaner <command> --cfgPath configs --cfgName local

# 3. URL-based configuration
gleaner <command> --cfgURL https://example.org/nabuconfig.yaml
```

### Minimal Configuration Example

```yaml
minio:
  address: localhost
  port: 9000
  ssl: false
  accesskey: myaccesskey
  secretkey: mysecretkey
  bucket: gleaner

objects:
  domain: us-east-1
  prefix:
    - summoned/providera
    - prov/providera

endpoints:
  - service: local_blazegraph
    baseurl: http://localhost:9090/blazegraph/namespace/earthcube
    type: blazegraph
    authenticate: false
    username: ""
    password: ""
    modes:
      - action: sparql
        suffix: /sparql
        accept: application/sparql-results+json
        method: GET
      - action: update
        suffix: /sparql
        accept: application/sparql-update
        method: POST
      - action: bulk
        suffix: /sparql
        accept: text/x-nquads
        method: POST
```

### Environment Variables

MinIO credentials can be set via environment variables instead of the config file:

```bash
export MINIO_ACCESS_KEY=myaccesskey
export MINIO_SECRET_KEY=mysecretkey
```

## Global Flags

These flags are available on all commands:

| Flag | Default | Description |
|------|---------|-------------|
| `--cfg` | | Full path to config file |
| `--cfgPath` | `configs` | Base directory for config files |
| `--cfgName` | `local` | Config subdirectory name |
| `--cfgURL` | | URL to fetch config from |
| `--nabuConfName` | `nabu` | Name of nabu config file |
| `--prefix` | | Limit operations to a specific S3 prefix |
| `--endpoint` | | Select a named SPARQL endpoint from config |
| `--address` | `localhost` | MinIO server address |
| `--port` | `9000` | MinIO server port |
| `--access` | env `MINIO_ACCESS_KEY` | MinIO access key |
| `--secret` | env `MINIO_SECRET_KEY` | MinIO secret key |
| `--bucket` | `gleaner` | S3 bucket name |
| `--ssl` | `false` | Use SSL for MinIO |
| `--dangerous` | `false` | Enable destructive operations |

---

## Commands

### config init - Initialize Configuration

Create a new configuration directory with template files.

```bash
gleaner config init                # creates ./configs/
gleaner config init myproject      # creates ./myproject/
```

---

### summon - Harvest JSON-LD from Data Sources

Reads configured data sources and harvests JSON-LD from websites via sitemaps,
APIs, or headless browser rendering. Stores harvested data in MinIO/S3.

```bash
# Using config path
gleaner summon --cfgPath configs --cfgName local

# Using full config path
gleaner summon --cfg /path/to/gleaner.yaml
```

---

### mill - Process Harvested Data

Processes harvested JSON-LD through the milling pipeline: converts JSON-LD to
RDF (N-Quads) and optionally runs SHACL validation.

```bash
gleaner mill --cfgPath configs --cfgName local
gleaner mill --cfg /path/to/gleaner.yaml
```

---

### prefix - Load Objects to Triplestore

Reads all objects from configured S3 prefixes and loads them into the triplestore.

```bash
# Load all configured prefixes
gleaner prefix --cfg example.yaml

# Load a specific prefix only
gleaner prefix --cfg example.yaml --prefix summoned/amgeo

# Using config path
gleaner prefix --cfgPath configs --cfgName local
```

---

### prune - Synchronize Graphs

Syncs the triplestore with the object store: removes orphaned graphs and adds
new ones. Updated objects get new SHA256-based names, so updates are treated
as new objects and old versions are pruned.

```bash
gleaner prune --cfg example.yaml
gleaner prune --cfg example.yaml --prefix summoned/amgeo
gleaner prune --cfgPath configs --cfgName local
```

---

### bulk - Bulk Load RDF

Generates all triples into a temporary file and bulk-loads them into the
triplestore via SPARQL UPDATE. The temp file is removed after loading.

The bulk endpoint configuration varies by triplestore:

**GraphDB** ([docs](https://graphdb.ontotext.com/documentation/10.2/)):
```yaml
endpointBulk: http://example.org:7200/repositories/testing/statements
endpointMethod: PUT
contentType: application/n-quads
```

**Jena** ([docs](https://jena.apache.org/tutorials/index.html)):
```yaml
endpointBulk: http://example.org:3030/testing/data
endpointMethod: PUT
contentType: application/n-quads
```

**Blazegraph** ([docs](https://github.com/blazegraph/database/wiki/REST_API)):
```yaml
endpointBulk: http://example.org:9090/blazegraph/namespace/kb/sparql
endpointMethod: POST
contentType: text/x-nquads
```

```bash
# Bulk load a specific prefix
gleaner bulk --cfg example.yaml --prefix summoned/providera

# Bulk load all configured prefixes
gleaner bulk --cfg example.yaml
```

---

### release - Build Release Graphs

Creates release graphs: the entire set of RDF objects for a provider rolled into
one file as N-Quads, with named graphs following the URN pattern defined in
[ADR 0001-URN-decision](https://github.com/gleanerio/nabu/blob/dev/decisions/0001-URN-decision.md).

```bash
# Release for a specific source
gleaner release --cfg example.yaml --prefix summoned/providera

# Release for all configured sources
gleaner release --cfg example.yaml
```

---

### object - Load a Single Object

Loads a single S3 object into the triplestore by its path.

```bash
gleaner object --cfg example.yaml milled/opentopography/ffa0df033bb3a8fc9f600c80df3501fe1a2dbe93.rdf

gleaner object --cfgPath configs --cfgName local milled/opentopography/abc123.rdf
```

---

### drain - Remove Objects from S3

Removes all objects from an S3 bucket prefix. Use `--prefix` to limit scope.

```bash
gleaner drain --cfg example.yaml --prefix summoned/providera
```

---

### clear - Clear All Graphs

Removes ALL graphs from the triplestore. Requires `--dangerous` flag.

```bash
gleaner clear --cfg example.yaml --dangerous
```

---

### graph - Graph Management

Parent command for graph operations.

#### graph clear

Clear all graphs from the triplestore (requires `--dangerous`):

```bash
gleaner graph clear --cfg example.yaml --dangerous
```

#### graph drop

Drop a specific named graph:

```bash
gleaner graph drop "http://example.org/mygraph" --cfg example.yaml
```

---

### meili - Load into MeiliSearch

Loads JSON-LD data into a MeiliSearch instance for full-text search indexing.

```bash
gleaner meili --cfg example.yaml
```

---

## Tools Examples

### Example 1: End-to-End Harvest and Load

This example shows a complete workflow: harvest data from a source, convert it
to RDF, and load it into a Blazegraph triplestore.

```bash
# Step 1: Harvest JSON-LD from the configured sitemap sources
gleaner summon --cfgPath configs --cfgName local

# Step 2: Mill the harvested JSON-LD into N-Quads RDF
gleaner mill --cfgPath configs --cfgName local

# Step 3: Load milled RDF into the triplestore
gleaner prefix --cfgPath configs --cfgName local --prefix milled/opentopography

# Step 4: Build a release graph for the provider
gleaner release --cfgPath configs --cfgName local --prefix summoned/opentopography
```

### Example 2: Prune and Reload a Source

When a data source has been updated and you want to sync the triplestore:

```bash
# Prune removes graphs that no longer have corresponding objects in S3,
# and adds any new objects that appeared since the last sync.
gleaner prune --cfg myconfig.yaml --prefix summoned/amgeo

# If you want a clean reload instead, drain + re-harvest:
gleaner drain --cfg myconfig.yaml --prefix summoned/amgeo
gleaner summon --cfg myconfig.yaml
gleaner mill --cfg myconfig.yaml
gleaner prefix --cfg myconfig.yaml --prefix milled/amgeo
```

### Example 3: Using URL-Based Configuration

Useful for CI/CD or shared team configurations hosted on a web server:

```bash
gleaner release \
  --cfgURL https://provisium.io/data/nabuconfig.yaml \
  --prefix summoned/dataverse \
  --endpoint localoxi
```

### Example 4: Working with Multiple Endpoints

When your config defines multiple SPARQL endpoints, use `--endpoint` to select one:

```yaml
# In your config file:
endpoints:
  - service: dev_blazegraph
    baseurl: http://localhost:9090/blazegraph/namespace/dev
    # ...
  - service: prod_graphdb
    baseurl: http://graphdb.example.org:7200/repositories/production
    # ...
```

```bash
# Load to the dev endpoint
gleaner prefix --cfg example.yaml --endpoint dev_blazegraph

# Load to production
gleaner prefix --cfg example.yaml --endpoint prod_graphdb
```

## Backward Compatibility

`gleaner` is the only binary; the former `nabu` and `glcon` entry points were
removed. All commands that worked with those tools work identically with
`gleaner`:

```bash
gleaner prefix --cfg example.yaml
```
