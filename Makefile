.PHONY: build test test-race test-frontend test-e2e test-all up down reset

build:
	cd backend && go build ./...

test:
	cd backend && go test ./...

test-race:
	cd backend && go test -race ./...

test-frontend:
	cd frontend && npm run test

test-e2e:
	npx playwright test

test-all: test test-frontend test-e2e

up:
	docker compose up --build -d

down:
	docker compose down

reset:
	docker compose down -v && docker compose up --build -d
