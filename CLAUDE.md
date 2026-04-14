# Crayfish Travel -- Reverse-matching travel platform

## What This Project Does
A travel platform where users describe their ideal trip and get matched with suitable travel products. Monorepo with Go backend, Next.js web app, and Alipay mini program.

## Tech Stack
- Backend: Go 1.26 + Gin + PostgreSQL + Redis + Asynq (async tasks)
- Web: Next.js 16 + React 19 + Tailwind 4 + shadcn/ui
- Mini Program: Taro 4 + React (Alipay target)
- Infra: Docker Compose, Swagger docs, DB migrations
- E2E: Playwright (separate e2e-tests/ package)

## Key Directories
- backend/ -- Go API server (Gin), migrations, mocks
- web/ -- Next.js consumer-facing web app
- miniprogram/ -- Taro-based Alipay mini program
- landing/ -- Static landing page
- shared/ -- Shared assets (compliance-terms.json)
- e2e-tests/ -- Playwright E2E test suite
- docker/ -- Docker Compose config
- scripts/ -- Compliance checks and utilities
- docs/ -- PRD, analysis, API requirements

## Development Commands
- Backend dev: `make dev` (or `cd backend && go run cmd/server/main.go`)
- Backend build: `make build`
- Backend test: `make test`
- Backend lint: `make lint`
- Swagger: `make swagger`
- Web dev: `make web-dev`
- Web build: `make web-build`
- Mini program dev: `make mini-dev`
- DB setup: `make setup` (docker-up + migrate-up)

## Notes
- The web/ subdirectory has its own CLAUDE.md referencing AGENTS.md with Next.js 16 breaking-change warnings.
- Mini program targets Alipay via Taro, not WeChat.
