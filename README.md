# Todo Vibe

A full-stack task management application with PIN-based authentication, an interactive calendar, and productivity stats.

---

## Features

- **PIN authentication** — first-time setup creates a PIN; subsequent visits require it to unlock
- **Daily task list** — create, edit, toggle done, and delete tasks for any date
- **Task attributes** — title, due time, priority (high / medium / low), tags, and point value
- **Mini calendar** — sidebar calendar shows done/total counts per day; drag a task onto a day to reschedule it
- **Stats chart** — area chart of completed tasks or points, aggregated by day, week, or month
- **Load-balanced backend** — two Go replicas behind Nginx with round-robin routing
- **Self-resetting E2E tests** — smoke suite resets state via a test-only endpoint before each run

---

## Architecture

```
Browser
  │
  ▼
┌─────────────────────────────────────┐
│  Nginx  (port 80)                   │
│  /api/*  →  backend_pool (2×Go)     │
│  /*      →  frontend (static)       │
└────────────┬──────────┬─────────────┘
             │          │
     ┌───────▼──┐   ┌───▼──────────────┐
     │ Backend  │   │  Frontend        │
     │ Go + Gin │   │  React 18 + Vite │
     │ ×2 replicas│  │  Tailwind CSS   │
     └───────┬──┘   └──────────────────┘
             │
     ┌───────▼──────┐
     │  PostgreSQL  │
     │  (pgdata vol)│
     └──────────────┘
```

### Backend (`backend/`)

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| HTTP router | Gin 1.9 |
| Database driver | sqlx + lib/pq |
| Migrations | golang-migrate |
| Auth | bcrypt PIN + session cookies |

**Package layout:**

```
backend/
├── main.go                  # Wiring: repos → services → handlers → routes
├── internal/
│   ├── config/              # Env-var config loader
│   ├── db/                  # Connection pool + migration runner
│   ├── models/              # Task, Session, Settings, priority enum
│   ├── repository/          # SQL — TaskRepo, SessionRepo, SettingsRepo
│   ├── services/            # Business logic — Auth, Task, Calendar, Stats
│   ├── handlers/            # HTTP handlers
│   └── middleware/          # Auth cookie validation, structured logger
└── migrations/              # Numbered SQL migration files
```

**API routes:**

```
GET  /healthz                        # Health check (unauthenticated)

/api/v1/auth
  GET    /status                     # {configured, authenticated}
  POST   /setup                      # First-time PIN creation
  POST   /login                      # PIN login → sets session cookie
  POST   /logout                     # Clears session cookie

/api/v1  (session cookie required)
  GET    /tasks?date=YYYY-MM-DD      # List tasks for a day
  POST   /tasks                      # Create task
  PUT    /tasks/:id                  # Update task
  PATCH  /tasks/:id/done             # Toggle done
  DELETE /tasks/:id                  # Delete task
  GET    /calendar?year=&month=      # Day-by-day done/total counts
  GET    /stats?view=day&metric=count# Chart data

/api/v1/test/reset  (APP_ENV=test only)
  POST   /test/reset                 # Wipe tasks, sessions, and PIN
```

### Frontend (`frontend/`)

| Concern | Technology |
|---|---|
| Framework | React 18 + TypeScript |
| Build tool | Vite 5 |
| Styling | Tailwind CSS 3 |
| Server state | TanStack React Query 5 |
| Charts | Recharts |
| HTTP client | Axios |

**Component tree:**

```
App
└── PinGate          (auth check → setup form or login form)
    └── AppShell     (header + two-column layout)
        ├── MiniCalendar   (sidebar — month navigation, drag target)
        ├── TaskList       (main — skeleton, empty state, task rows)
        │   └── TaskItem   (checkbox, edit, delete, drag handle)
        ├── TaskForm       (modal — create / edit)
        └── StatsChart     (area chart with view/metric toggles)
```

### Database schema

```sql
tasks     (id uuid PK, title, date, due_time, priority, tags[], points, done, created_at, updated_at)
sessions  (token uuid PK, created_at, expires_at, last_seen_at)
settings  (key text PK, value text)   -- stores pin_hash
```

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Docker Compose)
- [Node.js 20+](https://nodejs.org/) — only needed to run E2E / frontend tests locally
- [Go 1.22+](https://go.dev/dl/) — only needed to run backend tests locally

---

## Build & Run

### 1. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set a strong `SESSION_SECRET` (32+ random characters). The defaults work for local development.

To expose the test-reset endpoint (required for E2E tests), ensure:

```
APP_ENV=test
```

is present in `.env` (the example already includes it).

### 2. Start the stack

```bash
docker compose up --build
```

Add `-d` to run in the background:

```bash
docker compose up --build -d
```

The first run downloads base images and compiles both services. Subsequent runs use the Docker layer cache and are much faster.

### 3. Open the app

Navigate to **http://localhost** — you will be prompted to create a PIN on the first visit.

### 4. Stop the stack

```bash
docker compose down          # keeps the database volume
docker compose down -v       # also deletes the database volume (full reset)
```

---

## Running Tests

### Backend unit tests

```bash
cd backend
go test ./...
```

Run with verbose output and race detector:

```bash
go test -v -race ./...
```

### Frontend component tests

```bash
cd frontend
npm run test          # run once
npm run test:watch    # watch mode
```

### E2E smoke tests (Playwright)

Requires the full Docker stack to be running (`docker compose up -d`).

Install Playwright browsers once:

```bash
npx playwright install chromium
```

Run the suite:

```bash
npx playwright test
```

The first test (`first-time PIN setup`) calls `POST /api/v1/test/reset` in a `beforeAll` hook, wiping tasks, sessions, and the stored PIN so the suite is fully repeatable without manual database resets.

Run with the Playwright UI (headed, interactive):

```bash
npx playwright test --ui
```

Show the HTML report after a run:

```bash
npx playwright show-report
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `POSTGRES_DB` | yes | PostgreSQL database name |
| `POSTGRES_USER` | yes | PostgreSQL user |
| `POSTGRES_PASSWORD` | yes | PostgreSQL password |
| `DB_DSN` | yes | Full PostgreSQL DSN used by the backend |
| `SESSION_SECRET` | yes | Secret used for session integrity (32+ chars) |
| `APP_PIN_RESET` | no | Set to a PIN value to reset the stored PIN on next backend startup |
| `APP_ENV` | no | Set to `test` to enable the `POST /api/v1/test/reset` endpoint |
| `VITE_API_BASE_URL` | no | Frontend build-time API prefix (default `/api/v1`) |

---

## Project Structure

```
todo_vibe/
├── .env.example
├── docker-compose.yml
├── playwright.config.ts
├── package.json            # Playwright dev dependency
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── errors/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   ├── repository/
│   │   └── services/
│   └── migrations/
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── src/
│       ├── api/
│       ├── components/
│       ├── hooks/
│       └── types/
├── nginx/
│   ├── Dockerfile
│   └── nginx.conf
└── e2e/
    └── smoke.spec.ts
```
