# RecoLab — Universal Recommendation Engine Demo

A full-stack demo showcasing 4 real-world recommendation scenarios using [Gorse](https://gorse.io), an open-source Go recommendation engine.

## Scenarios

1. **Trending News** — Popular articles across categories
2. **Personalized News Feed** — User-specific news driven by collaborative filtering
3. **E-commerce Product Recommendations** — Personalised products + similar-item discovery
4. **Expo / Trade Booth Matchmaking** — Buyer↔booth affinity scoring

## Architecture

```
Next.js Frontend (port 3000)
        │
        │  /api/* proxy
        ▼
Go Wrapper API (port 3001)
        │  ┌─────────────────┐
        │  │ Business Rules  │
        │  │ Enrichment      │
        │  │ Explanation     │
        │  └─────────────────┘
        │
        ▼
Gorse REST API (port 8087)
        │
    ┌───┴────────┐
    ▼            ▼
PostgreSQL     Redis
(port 5435)  (port 6379)

Gorse Dashboard: http://localhost:8088
```

## Quick Start

### Option A — Docker Compose (all services)

```bash
docker-compose up -d
```

Wait ~30 seconds for Gorse to initialize, then seed data:

```bash
curl -X POST http://localhost:3001/api/demo/load-data
```

Open http://localhost:3000

### Option B — Local development

**Prerequisites:** Go 1.21+, Node.js 20+, Docker (for Gorse/Redis/Postgres)

```bash
# 1. Start infrastructure only
docker-compose up -d redis postgres gorse-master gorse-server gorse-worker

# 2. Start Go backend
cd backend
go run cmd/server/main.go

# 3. Start Next.js frontend (new terminal)
cd frontend
npm install
npm run dev
```

Open http://localhost:3000

## Seed Data

```bash
# Via HTTP
curl -X POST http://localhost:3001/api/demo/load-data

# Via Python helper script
python3 scripts/seed_gorse.py

# Regenerate all JSON files
python3 scripts/generate_demo_data.py
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Backend health check |
| POST | `/api/events` | Submit user feedback event |
| GET | `/api/demo/dashboard` | Stats (users/items/feedback counts) |
| POST | `/api/demo/load-data` | Seed all data into Gorse |
| GET | `/api/demo/trending-news?limit=10` | Trending news articles |
| GET | `/api/demo/personalized-news/:userId?limit=10` | Personalized news for user |
| GET | `/api/demo/products/:userId/recommendations?limit=10` | Product recommendations |
| GET | `/api/demo/products/:itemId/similar?limit=10` | Similar products |
| GET | `/api/demo/expo/buyer/:buyerId/recommendations?limit=10` | Booth recommendations for buyer |
| GET | `/api/demo/expo/seller/:sellerId/buyers?limit=10` | Potential buyers for seller |

### POST /api/events body
```json
{
  "feedbackType": "like",
  "userId": "u_ai_reader",
  "itemId": "news_001"
}
```

### Feedback types and weights
| Type | Weight | Description |
|------|--------|-------------|
| view | 1 | Item viewed |
| click | 2 | Item clicked |
| read | 3 | Article fully read |
| like | 5 | User liked |
| bookmark | 4 | Bookmarked |
| share | 4 | Shared |
| purchase | 8 | Purchased |
| inquiry | 6 | Inquiry sent (expo) |
| meeting_request | 7 | Meeting requested (expo) |
| hide | -2 | User hid item |

## Demo Data

| File | Count | Description |
|------|-------|-------------|
| `data/users.json` | 107 | Users with segmented interests |
| `data/news_items.json` | 198 | News articles (tech/trade/sports/biz) |
| `data/product_items.json` | 95 | Products across 5 categories |
| `data/expo_booths.json` | 50 | Exhibition booths (solar/textile/chem/food/tech) |
| `data/feedback.json` | 2555+ | Interaction events |

## Environment Variables

See `.env` for full list. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `GORSE_API_URL` | `http://localhost:8087` | Gorse server API |
| `GORSE_API_KEY` | `recolab_api_key_2026` | API authentication |
| `BACKEND_PORT` | `3001` | Go wrapper port |
| `TENANT_ID` | `recolab_demo` | Tenant identifier |

## Gorse Dashboard

Access the Gorse management dashboard at http://localhost:8088

Note: Collaborative filtering models need some time (and feedback data) to train. After seeding, wait a few minutes for Gorse to build its models. Until then, endpoints fall back to popularity-based recommendations.

hghgg
