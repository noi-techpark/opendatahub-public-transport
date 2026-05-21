# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Microservices for Open Data Hub public-transport data ingestion and transformation, written in Go. Two service kinds:

- **Collectors** (`collectors/`) — poll a data provider and publish the raw response, as-is, to RabbitMQ.
- **Transformers** (`transformers/`) — consume raw data from a queue, transform it, and push the result to its destination (here, an nginx fileserver).

The event/queue/raw-storage infrastructure (RabbitMQ, raw-data-bridge, fileserver) lives in a separate repo: https://github.com/noi-techpark/infrastructure-v2

## Repository layout & module model

This is a **multi-module Go workspace**. `go.work` ties the modules together for local development, but **it is gitignored** (a fresh clone won't have it; recreate with `go work init && go work use ./lib ./collectors/*/src ./transformers/*/src` if missing). The repo root is *not* a module — `go build ./...` from root fails. Run go commands **from within a module directory**.

- `lib/` — one module (`github.com/noi-techpark/opendatahub-public-transport/lib`) holding all shared packages. **Versioned and published**; services depend on a tagged version, not a `replace` directive.
- `collectors/<name>/src/`, `transformers/<name>/src/` — each service is its own module (`opendatahub.com/<name>`) with code under `src/`.

Consequence: **services do not pick up local `lib` changes through their `go.mod`** (no replace directives). The Docker build copies only `src/`, runs `go mod download`, and pulls the published `lib`. To ship a `lib` change you must tag/publish `lib` and bump the version in each consuming `go.mod`. Locally, `go.work` makes services see the on-disk `lib`.

### Shared library packages (`lib/`)
- `compress/` — gzip+base64 for shoving binary data through a JSON string field. Marked by metadata `compressed: gzip+base64`; use `compress.IsCompressed/Encode/Decode`.
- `go-siri/siri/` — SIRI Lite JSON codecs for VM (vehicle monitoring), ET (estimated timetable), SE (situation exchange). Handles the SIRI-Lite quirk where a field may be a single object *or* an array. JSON only; XML is stubbed.
- `go-gtfsrt/gtfsrt/` — GTFS-RT feed model with `Serialize(gtfsrt.FormatProtobuf | gtfsrt.FormatJSON)`. `pb/` holds generated protobuf from `proto/`.
- `gtfs-query/gtfs/` — in-memory GTFS store + query API (routes/trips/stops/calendar, service-on-date, stop-time matching). CLI: `gtfs-inspect/`.
- `netex-query/netex/` — streaming NeTEx XML parser with a **pluggable profile system** (`epip`, `it-l2`). Profiles register via `init()` and control both parsing and CSV output — import them for side effects: `_ ".../netex/profile"`. CLI: `netex2csv/`.

## Common commands

Run from inside the relevant module dir (`lib/` or a service's `src/`):

```bash
go build ./...                       # build a module
go test ./...                        # test a module
go test -run TestName ./...          # single test
go vet ./...                         # vet (note: lib currently has unkeyed-struct vet warnings)

# CLI tools (from lib/)
go run ./gtfs-query/gtfs-inspect -zip feed.zip -route 240 -date 20260521
go run ./netex-query/netex2csv -input feed.xml -profile epip -output out/
```

Local service run (RabbitMQ via compose, `dev` profile):
```bash
cd collectors/feed-fetcher        # or a transformer dir that has docker-compose.yml
cp .env.example .env
docker compose --profile dev up
```
The Dockerfile (`infrastructure/docker/Dockerfile`) is multi-stage: `build` (published image), `dev` (`go run .`), `test` (`go test .`).

## Service patterns

All services bootstrap with `ms.InitWithEnv(ctx, "", &env)` where `env` embeds `dc.Env` (collectors) or `tr.Env` (transformers) from `github.com/noi-techpark/opendatahub-go-sdk`. Telemetry is OpenTelemetry via the SDK's `tel`/`tel/logger`; always `defer tel.FlushOnPanic()`. Use `ms.FailOnError` for fatal startup errors.

- **Collector** (`dc.NewDc` → `StartCollection`/`Publish`): publishes `rdb.RawAny{Provider, Timestamp, Rawdata}`. `feed-fetcher` wraps each response in a `SiriPayload{Format, Protocol, Metadata, Payload}`.
- **Transformer** (`tr.NewTr[T]` → `listener.Start(ctx, handler)`): receives `*rdb.Raw[T]`; the SDK fetches the raw blob from the raw-data-bridge (`RAW_DATA_BRIDGE_ENDPOINT`). Handlers decompress with `compress`, deserialize, transform, and `PUT` to the fileserver (`FILESERVER_HOST` + path).

### Two flavors of config, deliberately

- **Single-purpose, env-var configured** — the `sta-{vm,et,se}-to-gtfsrt` transformers. One deployment per feed per provider, with custom transform logic. This is the **preferred default** for new transformers.
- **ConfigMap-driven multi-entry** — `feed-fetcher` (many endpoints) and `fileserver-saver` (many queue→path sinks). These spawn one goroutine per entry, each with panic-recovery + restart. Only reach for a ConfigMap when the config is genuinely a list; **do not** build generic config-driven dispatch when env vars on a dedicated binary suffice.

### The `sta-*-to-gtfsrt` transformers (the interesting ones)

They convert SIRI Lite → GTFS-RT, which requires mapping provider IDs to GTFS IDs. At startup they download **NeTEx + GTFS static data over FTP** (`NETEX_FTP_URL`, `GTFS_FTP_URL`), build a `Resolver`, and refresh every `REFRESH_HOURS` via an atomic swap (`StaticData` with RWMutex). To keep memory sane, NeTEx parsing loads only needed entity types (`store.OnlyTypes(...)`) and GTFS excludes `shapes.txt`/`translations.txt`.

`resolver.go` does the ID matching and is the trickiest code: it maps SIRI `LineRef`→GTFS route (via NeTEx public_code), and resolves trips in two phases — (A) extract the embedded NeTEx ServiceJourney ID and traverse SJ→JP→Route→Line→GTFS trip, scored by departure-time proximity; (B) fall back to matching scheduled stop time at the current stop. Output is written as both `.pb` and `.json`.

## Deployment & CI

- All services deploy via the **shared Helm chart `helm/generic-collector/`**; each service supplies its own `infrastructure/helm/values.yaml` (image, `env`, secret refs, optional `configMap.files`).
- Each service has a GitHub workflow (`.github/workflows/<svc>.yml`) triggered on path changes to that service **or `lib/**`**. It builds/pushes to `ghcr.io/...`, then deploys to **test on `main`** and **prod on the `prod` branch** (AWS EKS, namespace `pt`).

## Licensing — REUSE compliance (enforced in CI)

Every file needs an SPDX header or a `.license` sidecar. Code is `AGPL-3.0-or-later`; docs/config/generated files are `CC0-1.0`. A `reuse` pre-commit hook and the REUSE CI check will fail otherwise. Match the existing header style when adding files.
