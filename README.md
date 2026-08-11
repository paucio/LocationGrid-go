# LocationGrid-go

A small HTTP API for finding points near given coordinates. Point lookups are
served from a Redis-backed spatial grid index, and the matching point records
are then fetched from Postgres.

## How it works

- Coordinates are bucketed into fixed-size grid cells (`internal/cache/key.go`).
  Point IDs for each cell are cached in Redis under a key derived from the
  cell coordinates.
- `GET /search?x=<x>&y=<y>` resolves the cell for `(x, y)`, looks up the point
  IDs cached for that cell in Redis, then loads the full point records for
  those IDs from Postgres.

```
Request → SearchHandler → PointLookup (Redis: coords → point IDs)
                        → PointRepository (Postgres: IDs → points)
```

## Project layout

```
cmd/api/                    entrypoint (env config, wiring, HTTP server)
internal/cache/             Redis grid-cell key derivation + point ID lookup
internal/db/                Postgres connection pool setup
internal/handler/           HTTP handlers
internal/model/             domain types
internal/repository/        Postgres data access
```

## Requirements

- Go (see `go.mod` for the exact version)
- A reachable Postgres instance
- A reachable Redis instance

## Configuration

The server is configured via environment variables:

| Variable       | Description                          |
|----------------|---------------------------------------|
| `DATABASE_URL` | Postgres connection string            |
| `REDIS_ADDR`   | Redis address, e.g. `localhost:6379`  |

## Running

```sh
export DATABASE_URL="postgres://user:password@localhost:5432/locationgrid"
export REDIS_ADDR="localhost:6379"
go run ./cmd/api
```

The server listens on `:8080`.

### Example request

```sh
curl "http://localhost:8080/search?x=124.5&y=25.0"
```

Returns a JSON array of matching points, e.g.:

```json
[
  {"id": 101, "name": "Alpha", "x": 124.5, "y": 25.0}
]
```

## Testing

```sh
go test ./...
```

Tests use [`miniredis`](https://github.com/alicebob/miniredis) to fake Redis
and [`pgxmock`](https://github.com/pashagolub/pgxmock) to fake Postgres, so
no external services are required to run the test suite.
