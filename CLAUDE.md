# CLAUDE.md — Todo Vibe

Project guide for agentic coding. Read this before making changes.

---

## Key Commands

```bash
# Start full stack (first run builds images)
docker compose up --build -d

# Stop stack
docker compose down

# Full reset (wipes database volume)
docker compose down -v && docker compose up --build -d

# Backend
make build          # compile Go binary
make test           # go test ./...
make test-race      # go test -race ./...

# Frontend
make test-frontend  # vitest run

# E2E (requires stack running)
make test-e2e       # npx playwright test

# All tests
make test-all
```

---

## Architecture

```text
Browser → Nginx (port 80)
            ├── /api/*  → backend_pool (2 × Go replicas)
            └── /*      → frontend (React static files)
                              ↓
                        PostgreSQL (internal only)
```

**Backend layer order:** `handler → service → repository → DB`

- Handlers live in `backend/internal/handlers/` — bind input, call service, write JSON
- Services live in `backend/internal/services/` — business logic only, no HTTP
- Repositories live in `backend/internal/repository/` — raw SQL, no business logic
- Models live in `backend/internal/models/` — shared structs and constants

**Frontend data flow:** `component → hook → api/client.ts → backend`

- Hooks in `src/hooks/` wrap React Query queries/mutations
- All HTTP calls centralised in `src/api/client.ts`
- Components never call `axios` directly

---

## Environment Variables

| Variable | Required | Description |
| --- | --- | --- |
| `DB_DSN` | yes | Full PostgreSQL connection string |
| `SESSION_SECRET` | yes | 32+ char random string |
| `APP_ENV` | no | Set `test` to enable `POST /api/v1/test/reset` |
| `APP_PIN_RESET` | no | Reset stored PIN on next backend startup |
| `VITE_API_BASE_URL` | no | Frontend API prefix (default `/api/v1`) |

---

## API Routes

```text
GET  /healthz                             # unauthenticated health check
POST /api/v1/test/reset                   # test only (APP_ENV=test) — wipe all data

/api/v1/auth        (no session required)
  GET    /status    → {configured, authenticated}
  POST   /setup     → create PIN (first time only)
  POST   /login     → set session cookie
  POST   /logout    → clear session cookie

/api/v1             (session cookie required)
  GET    /tasks?date=YYYY-MM-DD
  POST   /tasks
  PUT    /tasks/:id
  PATCH  /tasks/:id/done
  DELETE /tasks/:id
  GET    /calendar?year=&month=
  GET    /stats?view=day|week|month&metric=count|points
```

---

## Coding Conventions

### Go (Backend)

- Use `CamelCase` for exported identifiers, `lowerCamelCase` for unexported
- All private packages go under `internal/`; follow the existing `handlers/`, `services/`, `repository/`, `models/`, `middleware/` layout
- Return custom error types from `internal/errors` (`ErrNotFound`, `ErrUnauthorized`, etc.); never leak raw DB errors to HTTP responses
- Write unit tests alongside source as `*_test.go` files
- Use structured logging (`log/slog`) — no `fmt.Print` in production paths
- Read configuration exclusively from environment variables via `internal/config`
- Validate all external input (query params, JSON bodies) at the handler layer before passing to services

### React / TypeScript (Frontend)

- Functional components with hooks only — no class components
- Components organised by feature under `src/components/<feature>/`
- File names use `PascalCase` for components (`TaskItem.tsx`), `camelCase` for hooks (`useTasks.ts`)
- Use Tailwind utility classes; no inline `style={}` props; no CSS modules
- Extract repeated Tailwind patterns into `@apply` in `index.css` only when used 3+ times
- Keep component files under ~200 lines; extract logic to custom hooks
- All server state through React Query; no manual `useEffect` for data fetching
- Implement `ErrorBoundary` around major subtrees

### PostgreSQL

- `snake_case` for table and column names
- Every table has `created_at TIMESTAMPTZ` and `updated_at TIMESTAMPTZ`
- Use migrations (golang-migrate numbered files) — never alter the schema manually
- Add indexes for any column used in a `WHERE` or `ORDER BY` on large tables
- Document non-obvious queries with an inline `-- comment`

### Docker / Nginx

- Pin image versions — no `:latest` tags
- Define health checks on every service
- Use named volumes for persistent data
- Enable gzip compression and set cache headers for static assets

### Markdown

- Blank line after every heading before the content that follows
- Blank line before and after every list block
- Never start a list immediately after a paragraph without a blank line

---

## Testing Strategy

### Backend unit tests (no DB)

- Handler tests use stub service interfaces (see `testhelpers_test.go` pattern)
- Service tests mock repositories via struct fields holding `func` values
- Integration tests (real DB) skip when `TEST_DB_DSN` is unset

### Frontend component tests

- Mock hooks with `vi.mock`; never make real HTTP calls in component tests
- Wrap renders in `<QueryClientProvider>` when components use React Query internally
- Use `@testing-library/user-event` for user interaction simulation

### E2E tests

- Require full `docker compose up` before running
- `beforeAll` in the PIN setup describe block calls `POST /api/v1/test/reset` for a clean slate
- Tests are independent of run order after the reset

---

## Important File Paths

| Path | Purpose |
| --- | --- |
| `backend/main.go` | Entry point — wiring of repos/services/handlers/routes |
| `backend/internal/models/settings.go` | `PinHashKey` constant |
| `backend/migrations/` | Numbered SQL migration files |
| `frontend/src/api/client.ts` | All HTTP calls |
| `frontend/src/types/index.ts` | Shared TypeScript interfaces |
| `docker-compose.yml` | Service definitions and env var forwarding |
| `.env` | Local secrets (git-ignored) |
| `.env.example` | Template — keep in sync with `.env` |
| `docs/details.md` | Full feature and data model specification |
| `docs/plan.md` | Original implementation plan |
