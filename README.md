# Gleaner

## About

`gleaner` is the unified command-line tool that merges the former **Gleaner**
(data harvesting) and **Nabu** (graph loading) projects into a single application.

It reads structured data (JSON-LD) from websites via sitemaps and APIs, stores it
in an S3-compatible object store (MinIO, AWS S3, Google Cloud Storage, Wasabi),
converts it to RDF, and loads/synchronizes it into a SPARQL triplestore.

## Quick start

```bash
# Build
make build              # produces ./gleaner

# Initialize a config directory (see configs/template/ for annotated examples)
./gleaner config init configs/local

# Harvest, process, and load
./gleaner summon --cfgPath configs --cfgName local
./gleaner mill   --cfgPath configs --cfgName local
./gleaner prefix --cfgPath configs --cfgName local
```

Configuration is split into three files so secrets stay separate from shareable settings:

| File | Contents | Share? |
|------|----------|--------|
| `services.yaml` | MinIO keys, SPARQL credentials | **Private** |
| `sources.yaml` | Data sources, context maps | Shareable |
| `gleaner.yaml` | Run settings, object prefixes | Shareable |

Further information can be found in the [documentation directory](./docs/README.md).

## Development

```bash
make check    # build + vet + tests + CLI smoke test
make test     # stable package tests
make smoke    # verify every subcommand's --help works
```
