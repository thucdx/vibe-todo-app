-- Speeds up task list queries filtered by date
CREATE INDEX idx_tasks_date ON tasks (date);

-- Speeds up calendar summary query (done/total per day)
CREATE INDEX idx_tasks_date_done ON tasks (date, done);

-- Speeds up session expiry checks
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
