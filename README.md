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
Returns a paginated list of users, with optional substring search and ordering.

Query parameters:

| param       | type   | default | notes                                               |
|-------------|--------|---------|-----------------------------------------------------|
| `page`      | int    | `1`     | 1-based page number                                 |
| `page_size` | int    | `20`    | items per page, capped at `100`                     |
| `q`         | string | (none)  | case-insensitive substring match on name OR email   |
| `sort`      | string | `id`    | order field: one of `id`, `name`, `email`           |
| `order`     | string | `asc`   | order direction: `asc` or `desc`                    |

Invalid or missing values fall back to the defaults. `q` is trimmed of
surrounding whitespace (an all-whitespace value is ignored) and capped at 200
characters. Unknown `sort`/`order` values fall back to `id`/`asc`. Without any
of these params the endpoint behaves exactly as before: ordered by `id` ascending
with no filter.

Examples:
```bash
# plain pagination
curl "http://localhost:8080/users?page=2&page_size=10"

# search users whose name or email contains "ada", newest ids first
curl "http://localhost:8080/users?q=ada&sort=id&order=desc"

# order by email ascending
curl "http://localhost:8080/users?sort=email&order=asc"
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
