# go-api-service

A simple Go REST API service for users.

## Requirements
- Go 1.21+
- PostgreSQL

## Running
```bash
export DATABASE_URL=postgres://user:pass@localhost/db?sslmode=disable
export PORT=8080
go run main.go
```

## Testing
```bash
make test
make test-race
```
