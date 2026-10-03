# myPersonalTechRadar (myPINS)

A personal technology intelligence platform that discovers, understands, filters, and personalizes technology information.

## Project Status

**This is a V0 Personal Open-Source Release (Completed through Phase 9).**
The project is functional for a single user locally, but it intentionally lacks multi-tenancy, pagination, scaling infrastructure, and advanced ML models. It is designed to be deterministic, understandable, and deeply hackable.

## Key Features

- **RSS Ingestion**: Periodically fetches and normalizes content from curated tech sources.
- **AI Enrichment**: Uses Google Gemini 2.5 Flash to generate summaries, extract topics, and explain "why it matters".
- **Deterministic Personalization**: Ranks the feed using a transparent Go algorithm based on:
  - Source Quality
  - Explicit User Interests
  - Active Learning Roadmap (Current vs Next priorities)
  - Time Decay (Freshness)
- **Feedback Learning**: Observes your reading habits (Save, Read, Dismiss) and calculates a bounded behavioral bonus for topics without silently overriding your explicit interests.
- **Explainable Feed**: Every item in the feed tells you *why* it was ranked highly (e.g., "Relevant to your current learning roadmap").
- **Deduplication**: Prevents duplicate content via canonical URLs and title/source matching.

## Architecture & Tech Stack

The project is built as a monolithic repository:
- **Language**: Go 1.22+
- **Database**: PostgreSQL 17
- **Frontend**: React SPA built with Vite and TypeScript
- **AI**: Google Gemini API

```text
React SPA (apps/web)  -->  Go API (cmd/api)  <-- PostgreSQL -->  Go Worker (cmd/worker)
```

## Repository Structure

- `cmd/api`: The Go HTTP REST server (port 8080). Also serves the compiled frontend.
- `cmd/worker`: The batch ingestion and AI enrichment job.
- `cmd/ingest`: A CLI tool for manual ingestion.
- `internal/`: Core domain logic (content, feedback, interests, ranking, roadmap, sources, topics).
- `apps/web`: The React frontend source code.
- `migrations/`: Raw SQL migrations for PostgreSQL.

## Prerequisites

- **Go**: `1.22+` (Built with `1.27.1`)
- **Node.js**: For building the React frontend.
- **Docker**: For running the PostgreSQL database.
- **Google Gemini API Key**: Free tier is sufficient. Get one at [Google AI Studio](https://aistudio.google.com/).

## Setup & Running Locally

### 1. Environment Configuration
```bash
cp .env.example .env
```
Edit `.env` to add your `GEMINI_API_KEY`. The default `DATABASE_URL` matches the included `docker-compose.yml`.

### 2. Start PostgreSQL
```bash
docker-compose up -d
```

### 3. Apply Migrations
Migrations are applied manually via `psql`.
```bash
for file in migrations/*.sql; do psql $DATABASE_URL -f "$file"; done
```

### 4. Build the Frontend
The Go API serves the frontend from `apps/web/dist`. You must build it first:
```bash
cd apps/web
npm install
npm run build
cd ../..
```

### 5. Start the API Server
```bash
go run ./cmd/api
```
The application will be available at `http://localhost:8080`.

### 6. Run the Ingestion Worker
The worker is a one-shot batch process. Run it periodically (via cron or manually) to fetch RSS feeds and enrich them with AI:
```bash
go run ./cmd/worker
```

## Docker Usage

A minimal multi-stage `Dockerfile` is included. It builds both Go binaries (`/app/api` and `/app/worker`) and packages the pre-built React frontend.

**Important:** You must build the React frontend locally (`cd apps/web && npm install && npm run build`) *before* running `docker build`.

```bash
docker build -t mypersonaltechradar .
docker run -p 8080:8080 --env-file .env mypersonaltechradar
```
To run the worker in Docker:
```bash
docker run --env-file .env mypersonaltechradar /app/worker
```

## Development & Testing

Format the code, tidy dependencies, and run the deterministic test suite:
```bash
gofmt -w .
go mod tidy
go test ./...
go vet ./...
```

## Limitations (V0)

- **No Pagination**: The `ListPersonalized` API fetches all unread items and ranks them in memory. This is highly testable and fast for a single user, but will not scale.
- **No Observability**: Lacks OpenTelemetry tracing or structured logging.
- **Single-Tenant**: Database unique constraints enforce a single user.

## License

This project is open-source under the [MIT License](LICENSE).
