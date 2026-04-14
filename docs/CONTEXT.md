# Crayfish Travel -- Project Context

## Tech Stack
- Go backend
- Next.js 16 / React web frontend
- Taro (Alipay miniprogram)
- Docker deployment
- GitHub Actions CI

## Current State (2026-04-14)
- Last commit: 32002e8 (date fallback fix + confirm page cleanup)
- Recent work: SSE streaming + skeleton screens (perf 3/10 -> 10/10)
- Git clean

## Key Directories
- backend/ -- Go API server
- web/ -- Next.js frontend
- miniprogram/ -- Taro Alipay miniprogram
- landing/ -- landing page
- deploy/ -- deployment configs
- docker/ -- Docker configs
- e2e-tests/ -- E2E test suite
- shared/ -- shared types/utils

## Verification Commands
- Web: `cd web && npm run dev`
- Backend: `cd backend && make run`
- E2E: `cd e2e-tests && npx playwright test`
- Health: http://localhost:3000

## Known Issues
- Missing CLAUDE.md (now added)
- No unit tests for backend
