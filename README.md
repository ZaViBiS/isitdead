# isitdead

## Quick Start With Docker

```sh
docker compose up --build
```

App: http://localhost:8080

PostgreSQL runs on `localhost:5432`:

```txt
postgres://isitdead:isitdead@localhost:5432/isitdead?sslmode=disable
```

## Local Backend

Start PostgreSQL only:

```sh
docker compose up postgres
```

Run the backend against it:

```sh
DATABASE_URL='postgres://isitdead:isitdead@localhost:5432/isitdead?sslmode=disable' go run ./cmd/server
```

## Tests

Run DB-backed tests through Docker:

```sh
make test-postgres
```

Or against an existing PostgreSQL database:

```sh
TEST_DATABASE_URL='postgres://isitdead:isitdead@localhost:5432/isitdead?sslmode=disable' make test
```
