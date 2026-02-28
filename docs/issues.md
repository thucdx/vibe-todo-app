# Code Review Issues

Found by automated code review. Organised by severity then layer.

---

## High Priority

### Backend

#### Error leakage to HTTP clients

Handlers pass `err.Error()` directly into JSON responses. This can expose SQL error
details, column names, and driver internals to any caller.

- `backend/internal/handlers/task_handler.go` — lines 41, 57, 78, 94, 113
- `backend/internal/handlers/calendar_handler.go` — line 43
- `backend/internal/handlers/stats_handler.go` — line 37

**Fix:** Return a generic `"internal server error"` message to the client and log the real
error server-side with `slog.Error`.

---

#### Missing input validation

User-supplied strings are forwarded to the database without format or range checks.

- `backend/internal/handlers/task_handler.go` — `date` query param accepts any string
  (e.g. `"2026-13-45"`); validate `YYYY-MM-DD` format before querying.
- `backend/internal/handlers/calendar_handler.go` — `year` and `month` params accept
  negative values and out-of-range months; clamp or validate before use.
- `backend/internal/models/task.go` — `DueTime *string` in `CreateTaskInput` /
  `UpdateTaskInput` has no `HH:MM` format or hour/minute range check.
- `backend/internal/repository/task_repo.go` — non-empty invalid priority strings
  (e.g. `"critical"`) are sent straight to the DB enum without validation.

---

### Frontend

#### Silent mutation failures

Most mutations lack `onError` callbacks. When the server returns an error the user
receives no feedback.

- `frontend/src/components/layout/AppShell.tsx` — `logout.mutate()` (line 20)
- `frontend/src/components/tasks/TaskList.tsx` — `deleteTask.mutate(id)` (line 20),
  `toggleDone.mutate(id)` (line 46)
- `frontend/src/components/sidebar/MiniCalendar.tsx` — `moveTask.mutate()` (line 47)
- No component renders the `isError` / `error` state from `useTasks`, `useCalendar`, or
  `useStats`.

**Fix:** Add `onError` callbacks to mutations and render an inline error message when
queries fail.

---

#### Accessibility — missing labels and focus management

Icon buttons contain only emoji or symbol characters with no `aria-label`, making them
unusable with screen readers.

- `frontend/src/components/tasks/TaskItem.tsx` — edit (✎) and delete (🗑) buttons
  (lines 78–92): add `aria-label="Edit task"` / `aria-label="Delete task"`.
- `frontend/src/components/layout/AppShell.tsx` — logout button 🔒 (line 21): add
  `aria-label="Lock"`.
- `frontend/src/components/sidebar/MiniCalendar.tsx` — prev/next buttons ‹ › (lines
  54–68): add `aria-label="Previous month"` / `aria-label="Next month"`.
- `frontend/src/components/tasks/TaskForm.tsx` — form labels not associated with inputs
  via `htmlFor` / `id`; modal lacks `role="dialog"` and focus trap.

---

## Medium Priority

### Backend

#### Duplicated defaulting logic in TaskRepo

Identical blocks applying default values for `priority`, `points`, and `tags` are
copy-pasted in `Create` and `Update`.

- `backend/internal/repository/task_repo.go` — lines 55–66 (Create) and 81–92 (Update)

**Fix:** Extract a private helper (e.g. `applyTaskDefaults`) in `task_service.go` or as a
method on the input struct.

---

#### Repeated UUID-parsing pattern in task handler

The same four-line block for parsing `:id` from the route path appears three times.

- `backend/internal/handlers/task_handler.go` — lines 66–69, 87–90, 103–106

**Fix:** Extract a `parseTaskID(c *gin.Context) (uuid.UUID, bool)` helper.

---

### Frontend

#### Over-aggressive cache invalidation

Every task mutation unconditionally invalidates `['tasks']`, `['calendar']`, and
`['stats']`, causing the full stats chart to refetch on every checkbox toggle.

- `frontend/src/hooks/useTasks.ts` — lines 17–21, 30–34, 54–58, 66–70, 79–82

**Fix:** Only invalidate `['stats']` from mutations that change `done` or `points`
(toggle, delete, move). Extract an `invalidateTaskQueries(qc, date)` helper to remove the
five identical blocks.

---

#### Unsafe `JSON.parse` on drag-drop data

`MiniCalendar` casts raw drag data to `Task` without any runtime check. Malformed data
can crash the component or cause silent type errors.

- `frontend/src/components/sidebar/MiniCalendar.tsx` — line 44

**Fix:** Wrap in `try/catch` and verify required fields (`id`, `date`) exist before
calling `moveTask.mutate`.

---

## Low Priority

### Backend

#### Expiry-slide error silently discarded in SessionRepo

`Validate` ignores the error returned by the session expiry `UPDATE`.

- `backend/internal/repository/session_repo.go` — line 50

**Fix:** Log the error with `slog.Warn("session expiry slide failed", "err", err)`.

---

#### No graceful shutdown

`r.Run()` blocks indefinitely with no OS signal handler, so SIGTERM skips database
connection cleanup and log flushing.

- `backend/main.go` — line 102

**Fix:** Replace `r.Run` with an `http.Server` and call `Shutdown(ctx)` on SIGINT/SIGTERM.

---

#### `config.go` panics on missing env vars

`panic()` prevents any deferred cleanup from running. Modern Go convention is to return an
`error` from `Load()` and let `main` decide how to exit.

- `backend/internal/config/config.go` — line 27

---

### Frontend

#### `summaryMap` rebuilt on every render in MiniCalendar

`new Map(summary.map(...))` runs unconditionally on every render even when `summary` has
not changed.

- `frontend/src/components/sidebar/MiniCalendar.tsx` — line 38

**Fix:** `const summaryMap = useMemo(() => new Map(summary.map(...)), [summary])`.

---

#### Duplicate toggle-button styling in StatsChart

The `VIEWS` and `METRICS` button groups use identical rendering logic.

- `frontend/src/components/stats/StatsChart.tsx` — lines 29–57

**Fix:** Extract a reusable `ToggleGroup` component.
