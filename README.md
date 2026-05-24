# AI Marketing Content Engine

A distributed AI content-generation platform that automates competitor research, keyword extraction, LLM content generation, ranking, deduplication, scheduling, and analytics.

Built as a hands-on learning project for backend / distributed-systems roles — Go services orchestrate a Kafka-decoupled pipeline, a C++ ranking engine handles the CPU-bound text math, and a RAG layer grounds LLM output in real crawled data.

> **Status:** Phase 3 / 9 — API Gateway in progress. Infrastructure and DB schema are up.

---

## Architecture

```
                                    ┌──────────────────┐
                                    │   Client / UI    │
                                    └────────┬─────────┘
                                             │ HTTPS + JWT
                                             ▼
                                  ┌────────────────────┐
                                  │  API Gateway (Go)  │   Gin · JWT · rate limit
                                  └────────┬───────────┘
                                           │ produce
                                           ▼
                                  ┌────────────────────┐
                                  │       Kafka        │   topics · DLQ · retries
                                  └────────┬───────────┘
                          ┌────────────────┼─────────────────┐
                          ▼                ▼                 ▼
                  ┌──────────────┐ ┌──────────────┐ ┌─────────────────┐
                  │ Crawler (Go) │ │ Ranking (C++)│ │ Content Gen (Go)│
                  │  worker pool │ │  TF-IDF · cos│ │  LLM + RAG      │
                  └──────┬───────┘ └──────┬───────┘ └────────┬────────┘
                         │                │                  │
                         └────────────────┼──────────────────┘
                                          ▼
                              ┌───────────────────────┐
                              │  PostgreSQL + pgvector│
                              └───────────┬───────────┘
                                          │
                              ┌───────────▼───────────┐
                              │  Scheduler (cron)     │   publish · schedule
                              └───────────────────────┘

  Redis caches LLM responses across the pipeline.
  Prometheus + Grafana observe every Go service.
```

Services communicate over Kafka so each one scales independently and retries are first-class. The C++ ranking engine is exposed over gRPC/HTTP — no cgo.

---

## Tech stack

| Layer            | Tech                                                                 |
| ---------------- | -------------------------------------------------------------------- |
| Services         | **Go** (API, crawler, content gen, scheduler), **C++** (ranking)     |
| Messaging        | Apache Kafka (Confluent 7.5) + Zookeeper, kafka-ui                   |
| Storage          | PostgreSQL 16 + **pgvector**, Redis 7                                |
| AI               | OpenAI or Groq for LLM; pgvector for embeddings (1536-d)             |
| Infra            | Docker Compose                                                       |
| Observability    | Prometheus, Grafana (Phase 9)                                        |
| Key Go libraries | `gin`, `pgx/v5`, `go-redis/v9`, `golang-jwt`, `colly`, `robfig/cron` |

---

## Quick start

### Prerequisites

- Docker Desktop
- Go 1.26+
- (later phases) a C++ toolchain, an OpenAI / Groq API key

### 1. Clone and configure

```powershell
git clone https://github.com/ilovecroissant/ai-marketing-engine.git
cd ai-marketing-engine
Copy-Item .env.example .env   # then edit values
```

`.env` keys used today:

```
DB_USER=admin
DB_PASSWORD=secret
DB_HOST=localhost
DB_PORT=5432
DB_NAME=marketing_engine
REDIS_ADDR=localhost:6379
```

### 2. Bring up infrastructure

```powershell
docker compose up -d
```

This launches:

| Service     | Port  | Notes                                |
| ----------- | ----- | ------------------------------------ |
| Postgres    | 5432  | pgvector preinstalled; schema auto-loads from `db/init.sql` |
| Redis       | 6379  | AOF persistence on                   |
| Kafka       | 9092  |                                      |
| kafka-ui    | 8080  | http://localhost:8080                |
| Zookeeper   | 2181  |                                      |

### 3. Run the API gateway

```powershell
cd api-gateway
go run .
```

You should see `Connected to PostgreSQL`, `Connected to Redis`, `api-gateway ready`.

### Tearing down

```powershell
docker compose down       # stop containers
docker compose down -v    # also drop the Postgres volume (wipes data)
```

---

## Database schema

Defined in [db/init.sql](db/init.sql) and applied automatically the first time Postgres starts.

- **users** — accounts (`free` / `pro` / `enterprise`)
- **campaigns** — a content request (product, industry, tone, platform, status)
- **contents** — generated articles with ranking subscores and publish state
- **analytics** — per-content engagement metrics + API latency
- **embeddings** — pgvector chunks used for RAG retrieval (Phase 7)

To reset the schema, drop the volume: `docker compose down -v`.

---

## Repository layout

```
ai-marketing-engine/
├── api-gateway/           # Go service (Phase 3) — Gin router, JWT, Kafka producer
│   ├── main.go
│   ├── db/connect.go      # shared pgxpool
│   └── cache/connect.go   # shared go-redis client
├── db/
│   └── init.sql           # schema, indexes, pgvector extension
├── docker-compose.yml     # Postgres, Redis, Kafka, Zookeeper, kafka-ui
├── CLAUDE.md              # context for AI coding assistants
└── README.md
```

Planned: `crawler/`, `ranking/` (C++), `content-gen/`, `scheduler/`.

---

## Build phases

The project follows a strict 9-phase roadmap. Each phase depends on the previous.

| #  | Phase                       | Highlights                                                             | Status |
| -- | --------------------------- | ---------------------------------------------------------------------- | ------ |
| 1  | Prerequisites               | Go concurrency, C++ STL, REST/JSON                                     | done   |
| 2  | Infrastructure              | docker-compose, DB schema, Postgres + Redis wired                      | done   |
| 3  | **API Gateway (Go)**        | Gin router, JWT, rate limiter, Kafka producer                          | **in progress** |
| 4  | Kafka                       | producer/consumer, DLQ, exponential backoff, idempotency               | todo   |
| 5  | Concurrent Crawler          | bounded worker pool, `colly`/`goquery`, context timeouts, robots.txt   | todo   |
| 6  | C++ Ranking Engine          | TF-IDF, cosine similarity, `0.4·SEO + 0.3·Engagement + 0.3·Readability`| todo   |
| 7  | AI Content Gen + RAG        | chunk → embed → retrieve → LLM; Redis cache by semantic input hash     | todo   |
| 8  | Scheduler + Analytics       | `robfig/cron`, latency middleware, optional React + Chart.js dashboard | todo   |
| 9  | Observability + Polish      | Prometheus, Grafana, Kafka consumer lag, load test < 120 ms avg        | todo   |

---

## Design decisions

- **Why Kafka?** Decouple services, support retries, enable asynchronous workflows, scale each worker independently.
- **Why Redis?** Eliminate repeated LLM API calls and reduce response latency for frequently requested content.
- **Why C++ for ranking?** CPU-intensive TF-IDF and cosine similarity benefit from lower-level memory control that Go can't match.
- **Why a worker pool, not unlimited goroutines?** Bounded concurrency prevents memory exhaustion and gives predictable resource usage under load.
- **Why gRPC/HTTP between Go and C++?** Cleaner boundary than cgo; keeps the interface explicit and minimal.

---

## License

MIT
