# Todo Vibe — Detailed Requirements

## 1. Overview

A minimalist single-user todo application with calendar-based task management, drag-and-drop date assignment, point tracking, and productivity charts. Protected by a PIN set on first use.

---

## 2. Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18+, Tailwind CSS |
| Backend | Golang (net/http or Gin) |
| Database | PostgreSQL 15+ |
| Reverse Proxy / LB | Nginx |
| Containerization | Docker Compose |

---

## 3. Infrastructure Architecture

```
                        ┌─────────────────┐
         Browser ──────▶│  Nginx (port 80) │
                        └────────┬────────┘
                        Round-robin load balance
                   ┌────────────┴────────────┐
                   ▼                         ▼
          ┌────────────────┐       ┌────────────────┐
          │  backend-1     │       │  backend-2     │
          │  Go service    │       │  Go service    │
          │  (port 8080)   │       │  (port 8080)   │
          └────────┬───────┘       └───────┬────────┘
                   └──────────┬────────────┘
                              ▼
                   ┌────────────────────┐
                   │  PostgreSQL DB      │
                   │  (port 5432)        │
                   └────────────────────┘
```

**Docker Compose services:**
- `nginx` — reverse proxy, exposed on host port `80`
- `backend` (scaled to 2 replicas) — Go API server
- `db` — PostgreSQL, internal network only
- `frontend` — React app served as static files via Nginx (or a separate static service)

---

## 4. Authentication

### PIN Setup Flow
1. On first visit, the app detects no PIN has been configured (no record in DB).
2. The user is prompted to **create a PIN** (4–8 digits or alphanumeric).
3. The PIN is hashed (bcrypt) and stored in the `settings` table.
4. A session token (UUID) is issued and stored in an `HttpOnly` cookie.

### Session Flow
- Each request carries the session cookie.
- The backend validates the session token against the `sessions` table (with expiry).
- If no valid session exists, the frontend redirects to the PIN entry screen.
- Session expiry: **7 days** (sliding window on activity).
- On PIN entry success, a new session token is issued.

### Reset
- PIN can be reset by setting the `APP_PIN_RESET` environment variable to a new PIN hash at startup. The backend detects this and replaces the stored PIN, invalidating all existing sessions.

---

## 5. Data Model

### `tasks`
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `title` | TEXT | Required, max 255 chars |
| `date` | DATE | The day this task belongs to |
| `due_time` | TIME | Optional specific time within the day |
| `priority` | ENUM `('high','medium','low')` | Default `'medium'` |
| `tags` | TEXT[] | Array of user-defined tag strings |
| `points` | INTEGER | Default `1`, min `1`, max `100` |
| `done` | BOOLEAN | Default `false` |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

### `settings`
| Column | Type | Notes |
|---|---|---|
| `key` | TEXT (PK) | e.g. `'pin_hash'` |
| `value` | TEXT | |

### `sessions`
| Column | Type | Notes |
|---|---|---|
| `token` | UUID (PK) | |
| `created_at` | TIMESTAMPTZ | |
| `expires_at` | TIMESTAMPTZ | |
| `last_seen_at` | TIMESTAMPTZ | |

---

## 6. Features

### 6.1 Task Management

#### Add Task
- Click a **"+ Add task"** button (in the main task area for the selected day).
- A form/modal appears with fields:
  - **Title** (required)
  - **Date** (pre-filled with the selected day)
  - **Due time** (optional time picker)
  - **Priority** (dropdown: High / Medium / Low, default Medium)
  - **Tags** (comma-separated or tag-chip input, user-defined)
  - **Points** (numeric input, default 1)
- Submit saves the task. The task list refreshes immediately.

#### Edit Task
- Click on any task to open an inline edit form or modal with the same fields.
- All fields are editable.
- Save closes the form and updates the task.

#### Delete Task
- Each task has a delete button (trash icon).
- A brief confirmation prompt appears before deletion.

#### Mark Done / Undone
- Each task has a checkbox.
- Toggling it flips `done` status immediately (optimistic UI update).
- Done tasks are visually struck through and dimmed.

### 6.2 Date Reassignment

#### Drag and Drop
- Tasks in the main task list can be dragged and dropped onto a calendar day in the sidebar mini-calendar.
- The task's `date` is updated to the target day.
- The task list for the original day re-renders (task removed) and the target day counter updates.

#### Manual Date Change
- Inside the task edit form, the **Date** field can be changed to any date.
- Saving moves the task to the new day.

### 6.3 Calendar Sidebar

- A **monthly mini-calendar** is always visible in the left sidebar.
- Navigation arrows allow moving to the previous / next month.
- Each day cell displays: `done / total` (e.g. `3/5`). Days with no tasks show nothing.
- The **currently selected day** is highlighted.
- **Today** is always highlighted with a distinct marker.
- Clicking a day sets it as the active view and loads its tasks in the main area.
- Days with all tasks completed get a visual "complete" indicator (e.g. green dot).

### 6.4 Main Task Area

- Header shows the selected date (e.g. "Thursday, Feb 27") with a "Today" shortcut button.
- Tasks for the selected day are listed, sorted by:
  1. Done status (undone first)
  2. Priority (High → Medium → Low)
  3. Due time (earliest first, tasks without time last)
- Each task row shows:
  - Checkbox (done toggle)
  - Title
  - Due time (if set)
  - Priority badge (color-coded: red = High, yellow = Medium, gray = Low)
  - Tags (small chips)
  - Points value
  - Edit and Delete actions (visible on hover)
- Empty state: friendly message "No tasks for this day. Add one!"

### 6.5 Productivity Chart

Located below the main task area or accessible via a **"Stats"** tab.

#### Views
| Toggle | Description |
|---|---|
| **Day** | Shows the last 30 days, one data point per day |
| **Week** | Shows the last 12 weeks, one data point per week |
| **Month** | Shows the last 12 months, one data point per month |

#### Metrics
| Toggle | Y-axis shows |
|---|---|
| **Count** | Number of tasks marked done in the period |
| **Points** | Sum of `points` for tasks marked done in the period |

#### Chart Spec
- Type: **Line chart** with area fill below the line.
- X-axis: date labels appropriate to the view (day: `Feb 27`, week: `Week 9`, month: `Feb`).
- Y-axis: numeric, auto-scaled.
- Hover tooltip shows: date range, done count, total points.
- Library suggestion: **Recharts** (React-native, lightweight).

---

## 7. UI Layout

```
┌─────────────────────────────────────────────────────────────┐
│  HEADER:  Todo Vibe                          [Lock / PIN]   │
├──────────────────┬──────────────────────────────────────────┤
│                  │                                          │
│  SIDEBAR         │  MAIN AREA                               │
│                  │                                          │
│  ┌────────────┐  │  Thursday, Feb 27           [Today]      │
│  │ Feb 2026 ◀▶│  │  ─────────────────────────────────────  │
│  │ Mo Tu We..│  │  ☐ Buy groceries    H  #home  1pt  ✎ 🗑  │
│  │  2  3 [4] │  │  ☑ Morning run      M         2pt  ✎ 🗑  │
│  │  9 10  .. │  │                                          │
│  │           │  │  + Add task                              │
│  └────────────┘  │                                          │
│                  │  ─────────────────────────────────────  │
│                  │  CHART (Stats)                           │
│                  │  [Day|Week|Month]  [Count|Points]        │
│                  │  ╭─────────────────────────────────╮    │
│                  │  │  📈 line chart                  │    │
│                  │  ╰─────────────────────────────────╯    │
│                  │                                          │
└──────────────────┴──────────────────────────────────────────┘
```

- Sidebar width: ~220px, fixed.
- Responsive: on mobile (<768px), sidebar collapses into a top-bar calendar icon that opens a drawer.

---

## 8. API Endpoints

All endpoints are prefixed `/api/v1`. Protected by session cookie unless noted.

### Auth
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/auth/status` | Returns `{ configured: bool, authenticated: bool }` |
| `POST` | `/api/v1/auth/setup` | Set PIN for first time. Body: `{ pin: string }` |
| `POST` | `/api/v1/auth/login` | Verify PIN, issue session. Body: `{ pin: string }` |
| `POST` | `/api/v1/auth/logout` | Invalidate session cookie |

### Tasks
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/tasks?date=YYYY-MM-DD` | List tasks for a specific day |
| `POST` | `/api/v1/tasks` | Create a task |
| `PUT` | `/api/v1/tasks/:id` | Update a task (all fields) |
| `PATCH` | `/api/v1/tasks/:id/done` | Toggle done status |
| `DELETE` | `/api/v1/tasks/:id` | Delete a task |

### Calendar
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/calendar?year=YYYY&month=MM` | Returns daily summary `{ date, done, total }` for the month |

### Stats
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/stats?view=day\|week\|month&metric=count\|points` | Returns aggregated chart data |

---

## 9. Non-Functional Requirements

| Requirement | Target |
|---|---|
| **Responsiveness** | Works on desktop and mobile (≥320px wide) |
| **Performance** | Task list loads in < 300ms on localhost |
| **Distributed** | At least 2 backend replicas behind Nginx with round-robin LB |
| **Stateless backend** | Sessions stored in DB; any replica can serve any request |
| **Data persistence** | All data in PostgreSQL; Docker volumes for persistence across restarts |
| **Security** | PIN hashed with bcrypt; session cookie: HttpOnly, SameSite=Strict |
| **Port exposure** | Only Nginx (port 80) and DB (optional, port 5432) exposed to host |

---

## 10. Out of Scope (v1)

- Multi-user support
- Email / push notifications
- Recurring tasks
- Task ordering / manual reordering within a day
- Offline mode / PWA
- Dark mode
