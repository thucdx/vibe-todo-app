# Todo Vibe — Implementation Plan

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Phase 1: Infrastructure & Docker Compose](#2-phase-1-infrastructure--docker-compose)
3. [Phase 2: Database Migrations](#3-phase-2-database-migrations)
4. [Phase 3: Backend (Go/Gin)](#4-phase-3-backend-gogin)
5. [Phase 4: Frontend (React 18 + TypeScript)](#5-phase-4-frontend-react-18--typescript)
6. [Phase 5: Testing](#6-phase-5-testing)
7. [Build Order](#7-build-order)
8. [Verification](#8-verification)

---

## 1. Project Structure

```
todo_vibe/
├── .env.example
├── docker-compose.yml
├── nginx/
│   ├── Dockerfile
│   └── nginx.conf
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── db/db.go
│   │   ├── errors/errors.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   └── logger.go
│   │   ├── models/
│   │   │   ├── task.go
│   │   │   ├── session.go
│   │   │   └── settings.go
│   │   ├── repository/
│   │   │   ├── task_repo.go
│   │   │   ├── session_repo.go
│   │   │   └── settings_repo.go
│   │   ├── services/
│   │   │   ├── auth_service.go
│   │   │   ├── task_service.go
│   │   │   ├── calendar_service.go
│   │   │   └── stats_service.go
│   │   └── handlers/
│   │       ├── auth_handler.go
│   │       ├── task_handler.go
│   │       ├── calendar_handler.go
│   │       └── stats_handler.go
│   └── migrations/
│       ├── 000001_init.up.sql
│       ├── 000001_init.down.sql
│       ├── 000002_add_indexes.up.sql
│       └── 000002_add_indexes.down.sql
├── frontend/
│   ├── Dockerfile
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── types/index.ts
│       ├── api/client.ts
│       ├── hooks/
│       │   ├── useAuth.ts
│       │   ├── useTasks.ts
│       │   ├── useCalendar.ts
│       │   └── useStats.ts
│       └── components/
│           ├── auth/PinGate.tsx
│           ├── layout/AppShell.tsx
│           ├── sidebar/MiniCalendar.tsx
│           ├── tasks/
│           │   ├── TaskList.tsx
│           │   ├── TaskItem.tsx
│           │   └── TaskForm.tsx
│           └── stats/StatsChart.tsx
└── req/
```

---

## 2. Phase 1: Infrastructure & Docker Compose

### `.env.example`

Define all environment variables with placeholder values:

- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` for the database container.
- `DB_DSN` — the full postgres connection string used by the backend (references the db service hostname).
- `SESSION_SECRET` — a random string used to sign tokens.
- `APP_PIN_RESET` — optional; if set at startup the backend replaces the stored PIN hash and clears all sessions, then ignores the value thereafter.
- `VITE_API_BASE_URL` — injected at frontend build time; defaults to `/api/v1`.

### `docker-compose.yml`

Define four services on a shared internal bridge network:

- **db**: Use `postgres:15.6-alpine`. Mount a named volume for `/var/lib/postgresql/data`. Set a health check using `pg_isready`. No ports exposed to the host.
- **backend**: Build from `./backend`. Set all env vars from `.env`. Declare `depends_on: db` with `condition: service_healthy`. Use `deploy.replicas: 2` so Compose creates two containers. Define a health check hitting `GET /healthz`. No ports exposed to the host — only reachable via Nginx on the internal network.
- **frontend**: Build from `./frontend`. Serves the pre-built React static files via its internal Nginx. No ports exposed directly.
- **nginx**: Build from `./nginx`. Expose port `80` on the host. Depends on both `backend` and `frontend`.

### `nginx/nginx.conf`

Configure the reverse proxy with:

- An `upstream backend_pool` block pointing to the `backend` service on port `8080`. Docker Compose DNS round-robins between the two replicas automatically.
- A `server` block on port 80 with:
  - **Static files**: `location /` serves the frontend's `/usr/share/nginx/html` directory. Use `try_files $uri /index.html` for SPA routing. Add `expires 1y` cache headers for hashed asset filenames (`.js`, `.css`, `.png`, etc.).
  - **API proxy**: `location /api/` proxies to `upstream backend_pool`. Set `proxy_http_version 1.1`, forward `Host`, `X-Real-IP`, `X-Forwarded-For` headers.
  - Enable `gzip` compression for `text/plain`, `text/css`, `application/json`, `application/javascript`.

### `backend/Dockerfile`

Two-stage build:

1. **Builder stage**: Start from `golang:1.22-alpine`. Copy `go.mod`/`go.sum`, run `go mod download`. Copy source and compile a static binary with `CGO_ENABLED=0`.
2. **Runtime stage**: Start from `alpine:3.19`. Copy the binary and the `migrations/` directory. Expose port 8080. Run the binary.

### `frontend/Dockerfile`

Two-stage build:

1. **Builder stage**: Start from `node:20.11-alpine`. Install dependencies with `npm ci`. Copy source and run `npm run build` to produce `/app/dist`.
2. **Runtime stage**: Start from `nginx:1.25.4-alpine`. Copy the `dist/` output into `/usr/share/nginx/html`. Expose port 80.

### `nginx/Dockerfile`

Start from `nginx:1.25.4-alpine`. Copy `nginx.conf` into `/etc/nginx/conf.d/default.conf`. The frontend static files are served from the `frontend` container — Nginx proxies `location /` to the frontend service (or alternatively the frontend dist is baked into the Nginx image via a multi-stage copy; choose one approach and be consistent).

---

## 3. Phase 2: Database Migrations

Use `golang-migrate`. Migrations run automatically at backend startup before the HTTP server starts listening.

### Migration 000001 — Initial Schema (up)

Create the full initial schema:

- Enable the `pgcrypto` extension for `gen_random_uuid()`.
- Create a `priority_level` PostgreSQL ENUM with values `'high'`, `'medium'`, `'low'`.
- Create the **`tasks`** table: `id UUID PK DEFAULT gen_random_uuid()`, `title TEXT NOT NULL` with a length check (1–255 chars), `date DATE NOT NULL`, `due_time TIME` (nullable), `priority priority_level NOT NULL DEFAULT 'medium'`, `tags TEXT[] NOT NULL DEFAULT '{}'`, `points INTEGER NOT NULL DEFAULT 1` with a check (1–100), `done BOOLEAN NOT NULL DEFAULT FALSE`, `created_at` and `updated_at` as `TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
- Create the **`settings`** table: `key TEXT PRIMARY KEY`, `value TEXT NOT NULL`.
- Create the **`sessions`** table: `token UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `expires_at TIMESTAMPTZ NOT NULL`, `last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
- Create a PL/pgSQL trigger function `update_updated_at()` that sets `NEW.updated_at = NOW()` and attach it as a `BEFORE UPDATE` trigger on `tasks`.

### Migration 000001 — Down

Drop the trigger, trigger function, all three tables, and the enum type in reverse dependency order.

### Migration 000002 — Indexes (up)

Add performance indexes:

- `CREATE INDEX idx_tasks_date ON tasks (date)` — speeds up task list queries filtered by date.
- `CREATE INDEX idx_tasks_date_done ON tasks (date, done)` — speeds up the calendar summary query that groups by date and filters on `done`.
- `CREATE INDEX idx_sessions_expires_at ON sessions (expires_at)` — speeds up session cleanup.

### Migration 000002 — Down

Drop all three indexes.

---

## 4. Phase 3: Backend (Go/Gin)

### Dependencies (`go.mod`)

Required packages:

- `github.com/gin-gonic/gin` — HTTP router and middleware framework.
- `github.com/jmoiron/sqlx` — extends `database/sql` with struct scanning.
- `github.com/lib/pq` — PostgreSQL driver.
- `github.com/golang-migrate/migrate/v4` — run migrations programmatically.
- `github.com/google/uuid` — UUID parsing and generation.
- `golang.org/x/crypto/bcrypt` — PIN hashing.

---

### `internal/config/config.go`

Load all configuration from environment variables into a `Config` struct at startup. Panic if any required variable (`DB_DSN`, `SESSION_SECRET`) is absent. Return a sensible default for optional variables (`PORT` defaults to `"8080"`).

### `internal/db/db.go`

Expose two functions:

- `Connect(dsn string) (*sqlx.DB, error)`: open a `sqlx` connection pool, configure `MaxOpenConns=25` and `MaxIdleConns=10`.
- `Migrate(db *sqlx.DB) error`: use `golang-migrate` to run all pending up-migrations from the `file://migrations` source. Treat `ErrNoChange` as success.

### `internal/errors/errors.go`

Define an `AppError` struct with `Code int` (HTTP status) and `Message string`. Pre-declare sentinel errors: `ErrNotFound` (404), `ErrUnauthorized` (401), `ErrBadRequest` (400), `ErrInvalidPIN` (401), `ErrPINAlreadySet` (409).

---

### Models (`internal/models/`)

**`task.go`**

- Define a `Priority` type (string alias) with constants `PriorityHigh`, `PriorityMedium`, `PriorityLow`.
- Define a `Task` struct with all fields matching the DB schema. Use `db:` tags for `sqlx`, `json:` tags for the API. `Tags` field uses `pq.StringArray`. `DueTime` is a `*string` (nullable, formatted as `"HH:MM"`).
- Define `CreateTaskInput` and `UpdateTaskInput` structs for request binding, with `binding:"required"` validation tags on required fields.

**`session.go`**

Define a `Session` struct matching the DB schema with `db:` tags.

**`settings.go`**

No struct needed beyond what the repo uses directly, but define constants for well-known keys (`PinHashKey = "pin_hash"`).

---

### Repositories (`internal/repository/`)

Each repository receives a `*sqlx.DB` and exposes focused DB methods.

**`task_repo.go`**

- `ListByDate(ctx, date string) ([]Task, error)`: query all tasks where `date = $1::date`, ordered by `done ASC`, then priority rank (`CASE` expression: high=1, medium=2, low=3), then `due_time ASC NULLS LAST`, then `created_at ASC`.
- `Create(ctx, CreateTaskInput) (*Task, error)`: insert a new row with defaults applied for empty `priority` (medium) and `points` (1). Return the full row via `RETURNING *`.
- `Update(ctx, id uuid.UUID, UpdateTaskInput) (*Task, error)`: update all mutable fields. Return the updated row via `RETURNING *`.
- `ToggleDone(ctx, id uuid.UUID) (*Task, error)`: flip `done = NOT done` for the given ID. Return the updated row.
- `Delete(ctx, id uuid.UUID) error`: delete by ID. Return `ErrNotFound` if no row was affected.
- `CalendarSummary(ctx, year, month int) ([]DaySummary, error)`: aggregate query that groups tasks by `date` for the given year/month, returning `done` count (filtered) and `total` count per day. Returns a `DaySummary` slice with fields `date string`, `done int`, `total int`.
- `ChartData(ctx, view, metric string) ([]ChartPoint, error)`: dynamic aggregation query. `view` controls the `GROUP BY` granularity (day = `date`, week = `DATE_TRUNC('week', date)`, month = `DATE_TRUNC('month', date)`) and the lookback window (30 days / 12 weeks / 12 months). `metric` controls the aggregate (`COUNT(*) FILTER (WHERE done)` or `SUM(points) FILTER (WHERE done)`). Since `view` and `metric` are validated to a fixed allowed set before calling, string-building the SQL is safe with no injection risk.

**`session_repo.go`**

- `Create(ctx, ttl time.Duration) (*Session, error)`: generate a new UUID token, compute `expires_at = now + ttl`, insert the row, return the session.
- `Validate(ctx, token uuid.UUID) (*Session, error)`: find the session where `token = $1 AND expires_at > NOW()`. If not found, return `ErrUnauthorized`. On success, slide the expiry window forward by 7 days with a separate `UPDATE`.
- `Delete(ctx, token uuid.UUID) error`: delete the session row.

**`settings_repo.go`**

- `Get(ctx, key string) (string, error)`: return `value` for the given key. Returns `sql.ErrNoRows` if absent.
- `Set(ctx, key, value string) error`: upsert — `INSERT ... ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value`.

---

### Services (`internal/services/`)

Services contain business logic and call repositories. They do not touch HTTP.

**`auth_service.go`**

- `IsPINConfigured(ctx) bool`: calls `settings.Get(ctx, "pin_hash")` and returns `true` if found.
- `SetupPIN(ctx, pin string) error`: return `ErrPINAlreadySet` if already configured. Validate pin length (4–8 chars). Hash with `bcrypt.DefaultCost`. Store via `settings.Set`.
- `ResetPIN(ctx, newPIN string) error`: hash and overwrite the pin hash. Delete all sessions.
- `Login(ctx, pin string) (tokenString string, error)`: get the stored hash, compare with `bcrypt.CompareHashAndPassword`. On failure return `ErrInvalidPIN`. On success, create a session with 7-day TTL and return the UUID string.
- `Logout(ctx, tokenStr string) error`: parse the UUID, call `sessions.Delete`.

**`task_service.go`**

Thin delegation layer. `List`, `Create`, `Update`, `ToggleDone`, `Delete` each call the corresponding repository method and return the result. Apply any default values here if not handled by the repo.

**`calendar_service.go`**

`MonthlySummary(ctx, year, month int) ([]DaySummary, error)`: delegates to `task_repo.CalendarSummary`.

**`stats_service.go`**

`ChartData(ctx, view, metric string) ([]ChartPoint, error)`: validate `view` ∈ `{day, week, month}` and `metric` ∈ `{count, points}` — return `ErrBadRequest` for invalid values. Delegate to `task_repo.ChartData`.

---

### Handlers (`internal/handlers/`)

Handlers are Gin handler functions. They bind/validate input, call a service, and write JSON responses.

**`auth_handler.go`**

- `Status`: call `svc.IsPINConfigured` and check whether `"session"` key is set in the Gin context (placed there by auth middleware). Respond `{configured, authenticated}`.
- `Setup`: bind `{pin}` JSON body. Call `svc.SetupPIN`. On `ErrPINAlreadySet` respond 409. On success respond 201.
- `Login`: bind `{pin}`. Call `svc.Login`. On `ErrInvalidPIN` respond 401. On success set an `HttpOnly`, `SameSite=Strict` cookie named `session_token` with 7-day max-age. Respond 200.
- `Logout`: read cookie, call `svc.Logout`. Clear the cookie by setting max-age to -1. Respond 200.

**`task_handler.go`**

- `List`: read `date` query param (required). Call `svc.List`. Respond 200 with task array.
- `Create`: bind `CreateTaskInput` JSON body. Call `svc.Create`. Respond 201 with created task.
- `Update`: parse `:id` UUID from path. Bind `UpdateTaskInput`. Call `svc.Update`. Respond 200.
- `ToggleDone`: parse `:id`. Call `svc.ToggleDone`. Respond 200 with updated task.
- `Delete`: parse `:id`. Call `svc.Delete`. Respond 204 No Content.

**`calendar_handler.go`**

`Summary`: read `year` and `month` query params (default to current year/month). Call `calendarSvc.MonthlySummary`. Respond 200 with array.

**`stats_handler.go`**

`Chart`: read `view` (default `"day"`) and `metric` (default `"count"`) query params. Call `statsSvc.ChartData`. Respond 200 with chart point array.

---

### Middleware (`internal/middleware/`)

**`auth.go`**

- Read the `session_token` cookie. If absent or not a valid UUID, abort with 401.
- Call `sessionRepo.Validate`. If it fails, abort with 401.
- If valid, store the session in the Gin context (`c.Set("session", sess)`) and call `c.Next()`.

**`logger.go`**

A Gin middleware that records the start time before `c.Next()`, then logs method, path, status code, duration, and client IP using `log/slog` structured logging.

---

### `main.go`

Wire everything together in this order:

1. Load config.
2. Connect to DB (retry with backoff if needed).
3. Run migrations.
4. If `cfg.PINReset` is non-empty, call `authSvc.ResetPIN` and log the result.
5. Instantiate all repos → services → handlers.
6. Create a `gin.New()` engine, attach `Logger` and `Recovery` middleware globally.
7. Register `GET /healthz` (unauthenticated, returns 200).
8. Register `/api/v1/auth/*` routes without auth middleware.
9. Register all other `/api/v1/*` routes on a group that uses the `Auth` middleware.
10. Call `r.Run(":" + cfg.Port)`.

---

## 5. Phase 4: Frontend (React 18 + TypeScript)

### Project Setup

- Scaffold with `npm create vite@latest` using the `react-ts` template.
- Install runtime dependencies: `@tanstack/react-query`, `axios`, `recharts`, `date-fns`.
- Configure Tailwind CSS with PostCSS. Set up `tailwind.config.js` to scan `./src/**/*.{ts,tsx}`.
- Configure `vite.config.ts` to proxy `/api` to `http://localhost:8080` in development so the dev server doesn't need Nginx running locally.
- Wrap `<App>` with `<QueryClientProvider>` in `main.tsx`.

### `src/types/index.ts`

Define TypeScript interfaces that mirror the API response shapes: `Task`, `CreateTaskInput`, `UpdateTaskInput`, `DaySummary`, `ChartPoint`, `AuthStatus`. Also define `Priority` as a union type `'high' | 'medium' | 'low'`, and `ViewMode` / `MetricMode` union types for the chart.

### `src/api/client.ts`

- Create an `axios` instance with `baseURL` from `import.meta.env.VITE_API_BASE_URL` and `withCredentials: true` so the session cookie is sent on every request.
- Export four API object groups — `authApi`, `tasksApi`, `calendarApi`, `statsApi` — each exposing typed functions that call the corresponding endpoints and return typed response data. This centralizes all HTTP calls.

---

### Hooks (`src/hooks/`)

**`useAuth.ts`**

- `useAuthStatus()`: React Query query for `GET /auth/status`. Set `retry: false` so the login/setup screen appears immediately without retrying on 401.
- `useLogin()`: mutation for `POST /auth/login`. On success, invalidate the `auth` query key so the app re-evaluates authentication state.
- `useSetupPIN()`: mutation for `POST /auth/setup`. On success, invalidate auth queries.
- `useLogout()`: mutation for `POST /auth/logout`. On success, invalidate all queries to clear cached data.

**`useTasks.ts`**

- `useTasks(date)`: query for `GET /tasks?date=<date>`. Disabled when `date` is empty.
- `useCreateTask(date)`: mutation for `POST /tasks`. On success, invalidate `['tasks', date]` and `['calendar']`.
- `useUpdateTask(date)`: mutation for `PUT /tasks/:id`. Same invalidation.
- `useToggleDone(date)`: mutation for `PATCH /tasks/:id/done`. Implement **optimistic update** — immediately flip `done` in the query cache before the request resolves. Roll back on error.
- `useDeleteTask(date)`: mutation for `DELETE /tasks/:id`. On success, invalidate tasks and calendar.
- `useMoveTask()`: mutation for `PUT /tasks/:id` that updates only the `date` field. On success, invalidate all `['tasks']` and `['calendar']` queries since both the source and target day are affected.

**`useCalendar.ts`**

`useCalendar(year, month)`: query for `GET /calendar?year=&month=`. Keyed by `['calendar', year, month]`.

**`useStats.ts`**

`useStats(view, metric)`: query for `GET /stats?view=&metric=`. Keyed by `['stats', view, metric]`.

---

### Components

**`PinGate.tsx`**

- Calls `useAuthStatus()`. Shows a centered full-screen loading spinner while loading.
- If `authenticated === true`, renders `children`.
- Otherwise, renders a centered card with a password `<input>` and a submit button. If `configured === false`, the button label is "Create PIN" and submit calls `useSetupPIN`. If `configured === true`, label is "Enter PIN" and calls `useLogin`. Show an inline error message below the input on failure.

**`AppShell.tsx`**

- Holds `selectedDate` state (a `Date` object, default `new Date()`).
- Renders a fixed top `<Header>` bar with the app name and a lock icon button (calls logout).
- Renders a two-column layout below: a fixed-width left sidebar containing `<MiniCalendar>`, and a scrollable main area containing the task list and stats chart.
- Passes `selectedDate` and `onSelectDate` down to `MiniCalendar` and the task area.

**`MiniCalendar.tsx`**

- Owns `viewMonth` state (a `Date` representing the month being displayed, default current month).
- Renders prev/next month navigation arrows and the month+year label.
- Renders a 7-column grid of day-of-week headers (Su–Sa).
- Calculates the leading padding cells using `getDay(startOfMonth(viewMonth))`.
- For each day in the month, renders a cell that:
  - Displays the day number.
  - Displays the `done/total` count below the number using data from `useCalendar`. Style green when all tasks are done and total > 0. Show nothing when total is 0.
  - Highlights the cell when it matches `selectedDate`.
  - Marks today's cell with a distinct ring or dot.
  - Calls `onSelectDate(day)` on click.
  - Implements `onDragOver` (call `e.preventDefault()` to allow drop) and `onDrop`: read the task JSON from `e.dataTransfer.getData('application/task')`, call `useMoveTask` with the task ID and the new date string.

**`TaskList.tsx`**

- Calls `useTasks(date)` with the formatted selected date.
- Shows a loading skeleton while loading.
- Shows the "No tasks for this day. Add one!" empty state if the task array is empty.
- Maps tasks to `<TaskItem>` components, passing `onToggle`, `onEdit`, `onDelete` callbacks.
- Renders an "+ Add task" button at the bottom that opens `<TaskForm>` in create mode.

**`TaskItem.tsx`**

- Makes the root element `draggable`. On `dragStart`, serialize the task as JSON and set it on `e.dataTransfer` under the key `'application/task'`.
- Renders a checkbox that calls `onToggle(task.id)` on change.
- Renders the title, struck through and dimmed when `task.done`.
- Renders the due time (if set) as a small text label.
- Renders a priority badge: red for high, yellow for medium, gray for low.
- Renders each tag as a small chip.
- Renders the points value.
- Shows edit and delete icon buttons on hover (`group-hover` in Tailwind). Delete button shows a `confirm()` dialog before calling `onDelete`.

**`TaskForm.tsx`**

- Accepts an optional `task` prop (edit mode if present, create mode otherwise).
- Controlled form with local state for: `title`, `date`, `due_time`, `priority`, `tags`, `points`.
- Pre-populate with task values in edit mode. In create mode, default `date` to the selected day and `priority` to `medium`.
- Tags field: a text input that splits on comma and renders each parsed tag as a removable chip.
- On submit: call `useCreateTask` or `useUpdateTask` depending on mode. Close the form on success.
- Rendered as a modal overlay (fixed backdrop + centered card).

**`StatsChart.tsx`**

- Local state for `view` (`'day' | 'week' | 'month'`) and `metric` (`'count' | 'points'`).
- Calls `useStats(view, metric)`.
- Renders two toggle button groups: one for view, one for metric.
- Renders a Recharts `<AreaChart>` with `<Area>` using a linear gradient fill, `type="monotone"` for smooth curves. Configure `<XAxis dataKey="label">`, `<YAxis>` with integer ticks, `<CartesianGrid>` for subtle grid lines, and `<Tooltip>` showing the period and value.

---

## 6. Phase 5: Testing

### Backend — Unit Tests

- Write unit tests in `*_test.go` files alongside each service file.
- **Auth service**: test `SetupPIN` stores a bcrypt hash, `Login` returns a token on correct PIN and an error on wrong PIN, `IsPINConfigured` returns correct booleans before and after setup.
- **Task service**: test `Create` applies default priority/points, `ToggleDone` flips the boolean, `Delete` returns `ErrNotFound` for unknown IDs.
- Use a real test Postgres database (spin up with `docker compose` in CI) or use interface mocking for the repositories.

### Backend — Handler Tests

- Use `net/http/httptest` and `gin.CreateTestContext` to test handlers in isolation.
- Mock service interfaces so tests don't need a DB.
- Test that invalid JSON bodies return 400, missing auth cookie returns 401, and valid requests return the expected status codes and JSON shapes.

### Frontend — Component Tests

- Use Vitest as the test runner and `@testing-library/react` for rendering.
- **PinGate**: mock `useAuthStatus` to test three states: loading, unauthenticated+unconfigured (setup form), unauthenticated+configured (login form), authenticated (renders children).
- **TaskItem**: verify checkbox calls `onToggle`, drag events set the correct `dataTransfer` data, done tasks have line-through styling.
- **TaskForm**: verify submitting with empty title shows validation, submitting valid data calls the create mutation.
- **StatsChart**: verify toggling view/metric buttons updates the active state and triggers a re-query.

### E2E Tests

Use Playwright targeting `http://localhost` with the full Docker stack running. The smoke test covers the full user journey:

- Visit app → PIN setup screen appears → enter PIN → main app loads.
- Add a task → verify it appears in the list.
- Toggle done → verify the task strikes through and the calendar counter shows `1/1`.
- Drag the task to a different day → verify it disappears from today.
- Click that other day → verify the task appears there.
- Open stats → verify the chart renders with at least one data point.

---

## 7. Build Order

Build incrementally so the app is always in a runnable state:

| Step | What to build |
| ---- | ------------- |
| 1 | Root `docker-compose.yml`, `.env.example`, Dockerfiles (all services with stubs) |
| 2 | Database migration SQL files |
| 3 | Backend: config, db connection, migration runner, `main.go` stub, health check route |
| 4 | Backend: errors, models, settings/session repos, auth service, auth handlers + routes |
| 5 | Backend: task repo, task service, task handlers + routes |
| 6 | Backend: calendar and stats repos + services + handlers |
| 7 | Frontend: Vite + Tailwind + React Query setup, `types/`, `api/client.ts` |
| 8 | Frontend: `useAuth` hook + `PinGate` component |
| 9 | Frontend: `AppShell` layout skeleton (sidebar + main area, no data yet) |
| 10 | Frontend: `useCalendar` hook + `MiniCalendar` component (click to select day) |
| 11 | Frontend: `useTasks` hooks + `TaskList` + `TaskItem` (display and toggle) |
| 12 | Frontend: `TaskForm` modal (create and edit) |
| 13 | Frontend: drag-and-drop from `TaskItem` to `MiniCalendar` day cell |
| 14 | Frontend: `useStats` hook + `StatsChart` component |
| 15 | Tests: backend unit + handler tests |
| 16 | Tests: frontend component tests |
| 17 | Tests: Playwright E2E smoke test |
| 18 | Polish: mobile responsive layout, error boundaries, empty states, loading skeletons |

---

## 8. Verification

### Start the full stack

```bash
cp .env.example .env
# Fill in real values for DB password and SESSION_SECRET
docker compose up --build
```

Docker Compose will start PostgreSQL and wait for its health check, then start two backend replicas (each runs migrations on startup), then the frontend static server, then Nginx.

### Check services are healthy

```bash
docker compose ps
# All services should show "healthy" or "running"
```

### Verify load balancing

Make several API requests and check that the backend access logs show requests distributed across both `backend-1` and `backend-2` containers:

```bash
docker compose logs backend
```

### Verify the API flow

Use `curl` to exercise each endpoint in order:

1. Check auth status — expect `configured: false, authenticated: false`.
2. Set up PIN — expect 201.
3. Login with the PIN — expect 200 and a `Set-Cookie` header with `session_token`.
4. Create a task for today — expect 201 with the task JSON.
5. List tasks for today — expect the task appears.
6. Toggle the task done — expect `done: true` in the response.
7. Fetch the calendar summary for the current month — expect `done: 1, total: 1` for today.
8. Fetch stats for `view=day&metric=count` — expect a data point for today with value 1.
9. Update the task's date to tomorrow — expect the `date` field changes.
10. Delete the task — expect 204.

### Verify the frontend in browser

1. Open `http://localhost` — PIN setup screen appears.
2. Enter a PIN — the main app loads with today's date selected.
3. Add a task via "+ Add task" — it appears in the list.
4. Toggle the checkbox — the task strikes through and the calendar counter shows `1/1`.
5. Drag the task to a different calendar day — it disappears from today.
6. Click that other day — it appears there.
7. Open the Stats section and toggle between Day/Week/Month and Count/Points — chart updates.
8. Click the lock icon — PIN screen reappears.

### Run automated tests

```bash
# Backend
cd backend && go test ./...

# Frontend component tests
cd frontend && npm run test

# E2E (requires stack running)
npx playwright test
```
