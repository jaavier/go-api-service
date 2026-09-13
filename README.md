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

## API

### GET /users
Returns a paginated list of users.

Query parameters:

| param       | type | default | notes                          |
|-------------|------|---------|--------------------------------|
| `page`      | int  | `1`     | 1-based page number            |
| `page_size` | int  | `20`    | items per page, capped at `100`|

Invalid or missing values fall back to the defaults.

Example:
```bash
curl "http://localhost:8080/users?page=2&page_size=10"
```

Response envelope:
```json
{
  "data": [
    { "id": 11, "name": "Ada", "email": "ada@example.com" }
  ],
  "page": 2,
  "page_size": 10,
  "total": 57,
  "total_pages": 6
}
```

## Testing
```bash
make test
make test-race
```
