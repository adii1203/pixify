build:
	go build -o bin/main cmd/api/main.go

run:
	air

migrate:
	go run cmd/migration/main.go