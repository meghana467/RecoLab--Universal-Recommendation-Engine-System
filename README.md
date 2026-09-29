# RecoLab

RecoLab is a full-stack recommendation demo built around [Gorse](https://gorse.io). It includes a Next.js interface, a Go API, PostgreSQL event persistence, and Redis-backed Gorse services. Demo data covers news, retail products, and expo matchmaking.

## Features

-Trending and personalized news, including evaluated recommendations.
- User based product recommendations and similar product suggestions.
- Expo booth, supplier, session, and buyer matchmaking.
- User onboarding by interests and feedback event submission.
- Data seeding, dashboard statistics, and offline recommendation evaluation (Precision, Recall, Hit Rate, MRR, NDCG).
- Deterministic label and popularity fallbacks while Gorse models train or when no recommendations are available.
- PostgreSQL persistence for feedback events; the service can still start when PostgreSQL is unavailable.

## Architecture and ports

| Service | Local URL / port | Purpose |
|---|---|---|
| Next.js frontend | http://localhost:3000 | Demo UI; proxies API requests to the Go backend |
| Go backend | http://localhost:3001 | Application API and recommendation/business logic |
| Gorse REST API | http://localhost:8087 | Recommendation and data API |
| Gorse master/dashboard | http://localhost:8088 | Gorse administration UI (default login: `admin` / `password`) |
| PostgreSQL | localhost:5435 | Gorse storage and backend `feedback_events` persistence |
| Redis | localhost:6379 | Gorse cache |

The Compose stack includes Redis, PostgreSQL, Gorse master/server/worker, the Go backend, and the frontend. Compose uses named volumes for database, cache, and Gorse state.

## Requirements

For the Docker workflow, install Docker Desktop (or Docker Engine) with Docker Compose. For local frontend/backend development, install Go 1.21 or later, Node.js 20 or later, and npm. The infrastructure services still run in Docker.

## Quick start with Docker

1. Create a local environment file from the example:

   ```powershell
   if (!(Test-Path .env)) { Copy-Item .env.example .env }
   ```

   Edit `.env` if you need non-default ports or credentials. Never commit `.env`.

2. Start the full stack from the repository root:

   ```powershell
   docker compose up --build -d
   ```

3. Open http://localhost:3000. To load the included demo dataset, use the UI or run:

   ```powershell
   Invoke-RestMethod -Method Post http://localhost:3001/api/demo/load-data
   ```

   Gorse needs time to ingest data and train its models. Recommendations use local fallback scoring while training is underway.

To stop the services, run `docker compose down`. Add `-v` only when you also want to delete the named database and Gorse data volumes.

## Local development

Start the infrastructure services from the repository root:

```powershell
if (!(Test-Path .env)) { Copy-Item .env.example .env }
docker compose up -d redis postgres gorse-master gorse-server gorse-worker
```

Run the backend in a second terminal. It loads the root `.env` automatically when started from `backend`:

```powershell
Set-Location backend
go run ./cmd/server
```

Run the frontend in another terminal:

```powershell
Set-Location frontend
npm ci
npm run dev
```

Open http://localhost:3000. For the local frontend, `NEXT_PUBLIC_API_URL` defaults to `http://localhost:3001`. When running commands from the repository root instead, set the variable explicitly for the frontend process or start the backend and frontend from their respective directories as above.

## Environment configuration

Copy `.env.example` to `.env` for local development. Values shown in the example are local demo defaults. Compose reads the published-port settings and `TENANT_ID` from `.env`; its service addresses and demo database credentials are set in `docker-compose.yml` and `gorse.toml`. If you change database credentials or the Gorse API key, update the matching configuration in those files together before exposing services beyond your machine. The Go backend also has fallback defaults for its configuration when variables are unset.

| Variable | Default | Used by |
|---|---|---|
| `GORSE_API_URL` | `http://localhost:8087` | Backend Gorse API address; Compose sets its internal service address |
| `GORSE_DASHBOARD_URL` | `http://localhost:8088` | Backend dashboard address |
| `GORSE_API_KEY` | `recolab_api_key_2026` | Gorse API authentication; keep in sync with `gorse.toml` |
| `POSTGRES_HOST` | `localhost` | Backend PostgreSQL host; Compose uses `postgres` |
| `POSTGRES_PORT` | `5435` | Backend PostgreSQL port; Compose uses container port `5432` |
| `POSTGRES_USER` | `gorse` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `gorse_pass` | PostgreSQL password |
| `POSTGRES_DB` | `gorse` | PostgreSQL database |
| `POSTGRES_SSLMODE` | `disable` | PostgreSQL connection SSL mode |
| `REDIS_URL` | `redis://localhost:6379` | Redis address; Compose uses `redis://redis:6379` |
| `BACKEND_PORT` | `3001` | Go API listen port |
| `BACKEND_HOST` | `0.0.0.0` | Go API listen host |
| `TENANT_ID` | `recolab_demo` | Tenant identifier attached to API operations |
| `DATA_DIR` | auto-detected (`data`) | Optional path to the JSON demo data directory |
| `NEXT_PUBLIC_API_URL` | `http://localhost:3001` | Frontend API target; Compose uses `http://backend:3001` |

Compose reads `FRONTEND_PORT` (default `3000`), `BACKEND_PORT` (default `3001`), `POSTGRES_PUBLISHED_PORT` (default `5435`), `REDIS_PUBLISHED_PORT` (default `6379`), and `TENANT_ID` (default `recolab_demo`) from `.env`. The PostgreSQL and Redis container ports remain `5432` and `6379` respectively. Compose uses its internal service URLs and the demo credentials declared in `docker-compose.yml` and `gorse.toml`; local host values in `.env` are for running the Go backend outside Docker.

## API overview

All application API routes are served by the Go backend. The frontend proxies `/api/*` to it, except `/api/demo/load-data`, which is handled by a Next.js route handler.

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Backend health and tenant identifier |
| GET | `/api/onboard/interests` | Available onboarding interests |
| POST | `/api/onboard` | Create/onboard a user with selected interests |
| POST | `/api/events` | Submit a feedback event to Gorse and PostgreSQL |
| POST | `/api/demo/load-data` | Seed included demo data into Gorse |
| GET | `/api/demo/dashboard` | User, item, and feedback counts plus recent events |
| GET | `/api/demo/evaluation?category=news&k=10&users=25` | Offline recommendation evaluation |
| GET | `/api/demo/trending-news?limit=10` | Trending news |
| GET | `/api/demo/personalized-news/{userId}?limit=10` | Personalized news for a user |
| GET | `/api/demo/evaluated-news/{userId}?limit=10` | Evaluated news strategy for a user |
| GET | `/api/demo/products/user/{userId}/recommendations?limit=10` | Product recommendations |
| GET | `/api/demo/products/item/{itemId}/similar?limit=10` | Similar products |
| GET | `/api/demo/expo/buyer/{buyerId}/booths?limit=8` | Booth recommendations |
| GET | `/api/demo/expo/buyer/{buyerId}/suppliers?limit=8` | Supplier recommendations |
| GET | `/api/demo/expo/buyer/{buyerId}/sessions?limit=8` | Session recommendations |
| GET | `/api/demo/expo/seller/{sellerId}/buyers?limit=8` | Potential buyers for a seller |
| GET | `/api/demo/expo/trending?limit=10` | Trending expo booths |

Example feedback request:

```json
{
  "feedbackType": "like",
  "userId": "u_ai_reader",
  "itemId": "news_001"
}
```

Feedback types include `view`, `read`, `click`, `like`, `bookmark`, `share`, `purchase`, `cart`, `inquiry`, `meeting_request`, and `hide`. Views and reads are treated as read signals; positive actions train recommendations, while hide is a negative signal.

## Data and scripts

The `data/` directory contains the demo JSON data loaded by the backend: users, feedback, news items, product items, expo booths, and sessions. Keep these input files in version control; do not ignore the directory.

Python utilities in `scripts/` generate demo data, convert external datasets, and seed Gorse. For example:

```powershell
python scripts/generate_demo_data.py
python scripts/seed_gorse.py
```

`seed_gorse.py` requires the Gorse service to be available and uses the configured API URL/key. The separate `convert_datasets.py` helper is for converting source datasets and may require additional input files/dependencies; consult its arguments before use.

## Persistence and evaluation

On backend startup, PostgreSQL's `feedback_events` table is created automatically when the database is reachable. `POST /api/events` writes an event to Gorse and PostgreSQL. The dashboard can fall back to Gorse/seed data if PostgreSQL is unavailable.

The evaluation endpoint withholds each user's last positive interaction and reports Precision@K, Recall@K, Hit Rate@K, MRR@K, and NDCG@K. Read signals (`view`, `read`) are excluded from positive holdout metrics.

## Project layout

```text
backend/       Go API, handlers, Gorse client, persistence, and Dockerfile
frontend/      Next.js app, API proxy, and Dockerfile
data/          Versioned demo JSON inputs
scripts/       Data generation, conversion, and seeding utilities
docker-compose.yml
.env.example   Safe local configuration template
.gitignore
gorse.toml     Gorse recommendation and API-key configuration
```

## Troubleshooting

- Check service state with `docker compose ps` and logs with `docker compose logs -f backend gorse-server`.
- If the frontend cannot reach the API, confirm it uses `http://localhost:3001` locally and `http://backend:3001` inside Compose.
- If PostgreSQL or Redis ports are already taken, set `POSTGRES_PUBLISHED_PORT` or `REDIS_PUBLISHED_PORT` in `.env`.
- If Gorse does not accept the backend key, ensure `GORSE_API_KEY` in `.env` matches the API key configured in `gorse.toml`, then recreate the backend service.
- Gorse model recommendations improve after feedback data has been loaded and its workers have trained the models.
