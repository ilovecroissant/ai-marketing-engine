# AI Marketing Content Engine

Distributed AI content-generation platform built as a learning project for backend / distributed-systems roles. Go + C++ services orchestrated via Kafka, backed by PostgreSQL (pgvector) and Redis, generating marketing content with an LLM grounded in crawled competitor data (RAG).

## Architecture

```
Client → API Gateway (Go/Gin) → Kafka → { Crawler (Go) , Ranking (C++) , Content Gen (Go+LLM) } → Postgres
                                  ↑                                                                  ↓
                                  └────────────── Scheduler (Go/cron) ←──────────────────────────────┘
                              Redis cache fronts LLM calls. Prometheus + Grafana observe everything.
```

Services are decoupled through Kafka topics so each worker scales independently and retries are first-class.

## Tech stack

- **Languages:** Go (services), C++ (ranking engine)
- **Infra:** Docker Compose, PostgreSQL 16 + pgvector, Redis 7, Kafka (Confluent 7.5) + Zookeeper, kafka-ui
- **Go libs:** `pgx/v5`, `go-redis/v9`, `gin`, `golang-jwt`, `golang.org/x/time/rate`, `colly`/`goquery`, `robfig/cron`, `godotenv`
- **AI:** OpenAI or Groq for LLM calls; pgvector for embeddings (1536-dim ada-002)
- **Observability (Phase 9):** Prometheus, Grafana

## Repository layout

- [docker-compose.yml](docker-compose.yml) — Postgres, Redis, Kafka, Zookeeper, kafka-ui
- [db/init.sql](db/init.sql) — schema: `users`, `campaigns`, `contents`, `analytics`, `embeddings` (auto-runs on first Postgres start)
- [api-gateway/](api-gateway/) — Go service, Phase 3 work-in-progress
  - [main.go](api-gateway/main.go) — entrypoint; loads `.env`, connects DB + Redis
  - [db/connect.go](api-gateway/db/connect.go) — pgxpool wrapper, exposes `db.Pool`
  - [cache/connect.go](api-gateway/cache/connect.go) — go-redis client, exposes `cache.Client`
- `.env` — local secrets (gitignored). Required keys: `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `REDIS_ADDR`
- Future services (not yet scaffolded): `crawler/`, `ranking/` (C++), `content-gen/`, `scheduler/`

Go module path: `github.com/ilovecroissant/ai-marketing-engine` (the whole repo is one module today; may split per service later).

## Common commands

```powershell
# Bring infra up / down
docker compose up -d
docker compose down

# Watch Kafka via UI
# → http://localhost:8080

# Run API gateway (from api-gateway/)
go run .
go build ./...
go test ./...

# Reset DB schema (drops volume — wipes data)
docker compose down -v ; docker compose up -d
```

## Build phases (roadmap)

The project follows a strict 9-phase order; each phase depends on the previous.

1. **Prerequisites** — Go concurrency (goroutines, channels, `sync.WaitGroup`, `context`), C++ STL, REST/JSON
2. **Infrastructure** — docker-compose, DB schema, Postgres + Redis wired from Go
3. **API Gateway (Go)** — Gin router, JWT auth, rate limiter (`x/time/rate`), Kafka producer 🚧
4. **Kafka** — producer/consumer, DLQ, exponential-backoff retries, idempotent handling
5. **Concurrent Crawler** — bounded worker pool over URL channel, `colly`/`goquery`, `context.WithTimeout`, respect robots.txt
6. **C++ Ranking Engine** — TF-IDF, cosine similarity, score = `0.4×SEO + 0.3×Engagement + 0.3×Readability`, exposed over gRPC/HTTP (not cgo)
7. **AI Content Gen + RAG** — chunk crawled text (~500 tokens) → embeddings in pgvector → retrieve → inject as LLM context; cache by `hash(industry+keywords+tone+platform)`
8. **Scheduler + Analytics** — `robfig/cron` publisher, request/latency middleware → `analytics` table, optional React+Chart.js dashboard
9. **Observability** — Prometheus metrics (request count, latency histograms, error rate, Kafka consumer lag), Grafana dashboards, load test for sub-120ms avg latency

When in doubt about scope, finish the current phase before pulling work forward from later phases.

## Key design decisions (for interviews / future me)

- **Kafka** — decouple services, async workflows, independent scaling, retries with DLQ.
- **Redis** — cache LLM responses; they are slow and expensive, and identical requests are common.
- **C++ for ranking** — TF-IDF / cosine similarity are CPU-bound and benefit from manual memory control.
- **Bounded worker pool over unbounded goroutines** — predictable resource use, no OOM under spiky load.
- **gRPC/HTTP for Go↔C++** — cleaner boundary than cgo; keep the interface minimal.
- **pgvector in main Postgres** — one fewer system to run; good enough for this scale.

## Conventions

- All services read config from env vars (loaded via `godotenv` in dev). Never commit `.env`.
- Postgres connection is a single shared `pgxpool.Pool` exposed as `db.Pool`; same pattern for `cache.Client`.
- IDs are UUIDs (`gen_random_uuid()`); timestamps default to `NOW()`.
- Kafka consumers must be idempotent (at-least-once delivery).
- Validate every input at the API gateway — bad data in Kafka causes silent downstream failures.
- Cache LLM calls aggressively, keyed by a hash of the semantic inputs (not the raw prompt).
