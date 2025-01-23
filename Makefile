build:
	go build -o bin/main cmd/api/main.go

run:
	go run cmd/api/main.go

migrate:
	go run cmd/migration/main.go