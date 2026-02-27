DROP TRIGGER IF EXISTS tasks_updated_at ON tasks;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS priority_level;
DROP EXTENSION IF EXISTS "pgcrypto";
