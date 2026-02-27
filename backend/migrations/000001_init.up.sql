-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Priority enum
CREATE TYPE priority_level AS ENUM ('high', 'medium', 'low');

-- Tasks
CREATE TABLE tasks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL CHECK (char_length(title) BETWEEN 1 AND 255),
    date        DATE        NOT NULL,
    due_time    TIME,
    priority    priority_level NOT NULL DEFAULT 'medium',
    tags        TEXT[]      NOT NULL DEFAULT '{}',
    points      INTEGER     NOT NULL DEFAULT 1 CHECK (points BETWEEN 1 AND 100),
    done        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Key-value store for app settings (e.g. pin_hash)
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- User sessions
CREATE TABLE sessions (
    token        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auto-update updated_at on task modifications
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
