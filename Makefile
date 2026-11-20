.PHONY: run build test migrate-up migrate-down tidy

run:
	go run ./cmd/api/main.go

build:
	go build -o bin/carhistory ./cmd/api/main.go

test:
	go test ./... -v

tidy:
	go mod tidy

migrate-up:
	psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f migrations/001_initial_schema.sql

migrate-down:
	psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f migrations/002_drop_schema.sql

lint:
	golangci-lint run ./...

docker-build:
	docker build -t carhistory .

docker-run:
	docker-compose up -d
