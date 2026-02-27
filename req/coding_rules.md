
# Coding Rules & Conventions

## Frontend (React 18+, Tailwind CSS)

- Use functional components with hooks
- Keep components in `src/components/` organized by feature
- Use camelCase for file names (e.g., `UserCard.jsx`)
- Implement prop validation with TypeScript or PropTypes
- Use Tailwind utility classes; avoid inline styles
- Extract repeated Tailwind patterns into `@apply` classes
- Keep component logic under 200 lines; extract custom hooks
- Use React Query or SWR for data fetching
- Implement error boundaries for error handling

## Backend (Go with net/http or Gin)

- Follow Go conventions: CamelCase for exported, lowercase for unexported
- Place code in `internal/` for private packages
- Use `handlers/`, `models/`, `services/`, `middleware/` directory structure
- Implement proper error handling with custom error types
- Write unit tests alongside code (filename: `*_test.go`)
- Use environment variables for configuration
- Implement structured logging
- Add request/response validation middleware

## Database (PostgreSQL 15+)

- Use snake_case for table and column names
- Include `created_at` and `updated_at` timestamps
- Define foreign keys with `ON DELETE CASCADE` or `SET NULL`
- Create indexes on frequently queried columns
- Use migrations (e.g., Flyway, golang-migrate)
- Document complex queries with comments

## Nginx & Docker Compose

- Configure gzip compression for responses
- Set appropriate cache headers
- Use environment variables in `docker-compose.yml`
- Pin service versions; avoid `latest` tags
- Define health checks for all services
- Use named volumes for persistent data

## Testing

- Write comprehensive test, including unit test and e2e test

## Markdown

- Always include a blank line after every heading before the content that follows (MD022).
- Always surround bullet/numbered lists with blank lines — one blank line before the first item and one after the last (MD032).
- Do not start a list immediately after a paragraph on the very next line; always insert a blank line between them.
