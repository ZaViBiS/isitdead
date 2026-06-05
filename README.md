# isitdead

`isitdead` is a website monitoring application. It serves a web frontend, stores application state in PostgreSQL, and is being structured to support notifications through external providers such as Telegram and Discord.

## Features / Roadmap

### Backend

- [x] Go HTTP server
- [x] Static frontend serving from embedded `web/dist`
- [x] Health endpoint at `/api/health`
- [x] PostgreSQL connection setup
- [x] Database migration entry point
- [ ] norifications
  - [ ] telegram
  - [ ] discord
  - [ ] slack
  - [ ] email
- [ ] User auth
- [ ] Monitor CRUD API
- [ ] Public status page API
- [ ] Background monitor/checker worker
- [ ] HTTP website checks
- [ ] TCP service checks
- [ ] Push/heartbeat monitors
- [ ] SSL certificate checks
- [ ] Incident/history storage
- [ ] Billing/subscription API

### Frontend

- [x] Landing page
- [x] Pricing page
- [x] Register page UI
- [x] Login page UI
- [x] Dashboard UI
- [x] Monitor detail page UI
- [x] Public status page UI
- [ ] Fully connected authentication flow
- [ ] Fully connected monitor management flow
- [ ] Push/heartbeat monitor setup UI
- [ ] Fully connected public status pages

## Project Structure

```text
cmd/server
  Application entry point. Creates shared dependencies and starts the HTTP server.

internal/api
  HTTP routing and request handlers.

internal/database
  Database connection and migrations.

internal/notification
  Notification-related interfaces and provider implementations.

web
  Frontend application and embedded build output.
```

## Architecture

```text
frontend
   |
   v
api
   |
   +--> database
   |
   +--> notification service
             |
             +--> telegram notifier
             +--> discord notifier
```

More details:

- [Architecture](docs/architecture.md)

## Configuration

Configuration is provided through environment variables.

```text
DATABASE_URL=postgres://isitdead:isitdead@postgres:5432/isitdead?sslmode=disable
TELEGRAM_TOKEN=
```

See [.env.example](.env.example) for the current list.

## Running Locally

With Docker Compose:

```sh
docker compose up --build
```

With go build

```sh
go build -o bin/isitdead ./cmd/server
```

```sh
./bin/isitdead
```

The Go server listens on port `8080`.
