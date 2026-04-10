.PHONY: build-backend build-frontend build-all run-db setup-dev test-all

# Build backend
build-backend:
	cd backend && go build -o ../bin/uptimer ./cmd/uptimer/main.go

# Build frontend
build-frontend:
	cd frontend && npm install && npm run build
	mkdir -p backend/static
	cp -r frontend/dist/* backend/static/

# Build all
build-all: build-frontend build-backend

# Start local database
run-db:
	docker compose up -d

# Initial development setup
setup-dev: run-db
	cd backend && go mod download
	cd frontend && npm install

# Docker
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Run all tests
test-all:
	cd backend && go test ./...
	cd frontend && npm test
