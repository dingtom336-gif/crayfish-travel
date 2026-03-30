.PHONY: dev build test lint swagger compliance migrate-up migrate-down docker-up docker-down

# Backend
dev:
	cd backend && go run cmd/server/main.go

build:
	cd backend && go build -o bin/server cmd/server/main.go

test:
	cd backend && go test ./... -v -cover -coverprofile=cover.out

lint:
	cd backend && golangci-lint run ./...

swagger:
	cd backend && swag init -g cmd/server/main.go -o docs

# Compliance
compliance:
	bash scripts/compliance-check.sh

# Database
migrate-up:
	cd backend && migrate -path migrations -database "postgres://crayfish_user:crayfish_dev_password@localhost:5432/crayfish_travel?sslmode=disable" up

migrate-down:
	cd backend && migrate -path migrations -database "postgres://crayfish_user:crayfish_dev_password@localhost:5432/crayfish_travel?sslmode=disable" down 1

# Docker
docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

# Frontend
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Miniprogram
mini-dev:
	cd miniprogram && npm run dev:alipay

# All
setup: docker-up migrate-up
	@echo "Dev environment ready."
