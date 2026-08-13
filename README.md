# LocationGrid-go

A small HTTP API for finding points near given coordinates, plus a batch job
for importing points from a CSV file. Point lookups are served from a
Redis-backed spatial grid index, and the matching point records are then
fetched from Postgres.

## How it works

- Coordinates are bucketed into fixed-size grid cells (`internal/cache/key.go`).
  Point IDs for each cell are cached in Redis under a key derived from the
  cell coordinates and the point's `type` (e.g. `point_shop_x:...:y:...`), so
  different point types are indexed independently even at the same location.
- `GET /search?x=<x>&y=<y>&type=<type>&limit=<limit>` resolves the cell for
  `(x, y)`, then searches outward ring by ring (cell, then its neighbors at
  radius 1, 2, ...) through Redis for IDs matching `type`, stopping once at
  least `limit` IDs have been found (it does not trim the result down to
  exactly `limit`). It then loads the full point records for those IDs from
  Postgres.
- The import job (`cmd/import-job`) streams a CSV of points, batches them
  (1000 at a time), bulk-inserts each batch into Postgres, and then indexes
  the newly assigned IDs into the same Redis grid cache so they're
  immediately searchable.

```
Search:  Request → SearchHandler → PointLookup (Redis: coords → point IDs)
                                 → PointRepository (Postgres: IDs → points)

Import:  CSV → Importer → PointRepository.BulkInsert (Postgres: insert, assign IDs)
                        → PointLookup.AddIDsBulk (Redis: index IDs by grid cell)
```

## Project layout

```
cmd/api/                    HTTP API entrypoint (env config, wiring, server)
cmd/import-job/              CSV import job entrypoint
internal/cache/              Redis grid-cell key derivation, point ID lookup + indexing
internal/db/                 Postgres connection pool setup
internal/handler/             HTTP handlers
internal/job/                 CSV parsing and the batch importer
internal/model/               domain types
internal/repository/          Postgres data access (find, bulk insert)
```

## Requirements

- Go (see `go.mod` for the exact version)
- A reachable Postgres instance
- A reachable Redis instance

## Configuration

Environment variables, by entrypoint:

| Entrypoint         | Variable       | Description                                          |
|---------------------|----------------|-------------------------------------------------------|
| `cmd/api`           | `DATABASE_URL` | Postgres connection string                            |
| `cmd/api`           | `REDIS_ADDR`   | Redis address, e.g. `localhost:6379`                   |
| `cmd/import-job`    | `DATABASE_DSN` | Postgres connection string                            |
| `cmd/import-job`    | `REDIS_ADDR`   | Redis connection URL, e.g. `redis://localhost:6379`    |

Note the two entrypoints use different variable names/formats for the same
underlying config — check each `main.go` if you're wiring up deployment.

## Running

### API server

```sh
export DATABASE_URL="postgres://user:password@localhost:5432/locationgrid"
export REDIS_ADDR="localhost:6379"
go run ./cmd/api
```

The server listens on `:8080`.

```sh
curl "http://localhost:8080/search?x=124.5&y=25.0&type=1&limit=10"
```

`x` and `y` are required floats, `type` is a required integer, and `limit`
is a required positive integer (the minimum number of results to search
for). Returns a JSON array of matching points, e.g.:

```json
[
  {"id": 101, "name": "Alpha", "x": 124.5, "y": 25.0, "type": "1"}
]
```

### Import job

```sh
export DATABASE_DSN="postgres://user:password@localhost:5432/locationgrid"
export REDIS_ADDR="redis://localhost:6379"
go run ./cmd/import-job -url "https://example.com/points.csv"
```

Expects CSV rows of `name,x,y,id` (no header row). Rows with fewer than 4
columns, or with non-numeric `x`/`y`, are skipped and counted separately
from successfully imported points.

> **Known gap:** the CSV parser doesn't populate `Point.Type`, so every
> imported point is indexed in Redis under an empty type. Since `/search`
> always queries with a specific numeric `type`, points imported via this
> job are not currently discoverable through the API. Fix this in
> `internal/job/csvparse.go` if/when the CSV format grows a type column.

## Testing

```sh
go test ./...
```

Tests use [`miniredis`](https://github.com/alicebob/miniredis) to fake Redis
and [`pgxmock`](https://github.com/pashagolub/pgxmock) to fake Postgres, so
no external services are required to run the test suite.

## CI

`.github/workflows/ci.yml` runs on every push to `main` and on pull requests:
`gofmt` check, `go vet`, `go build`, and `go test -race`.
