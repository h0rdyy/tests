# Go backend migration

The Flask application is replaced by `main.go`. Existing Jinja templates and all
frontend assets remain unchanged. Rendering uses Pongo2, a Go Jinja-compatible
engine, with a small `url_for` compatibility function; this avoids converting
Jinja syntax and preserves the current HTML.

## Run

```sh
go run ./cmd/server
```

Runtime configuration: `PORT` (default 5000), `DB_PATH`, `LOG_PATH`, and SMTP
variables `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`,
`RECEIVER_EMAIL`. Leads are stored in SQLite at `data/cybersec.db`; email is
sent asynchronously after persistence. The admin CMS remains intentionally
out of scope; the database/repository boundary is the extension point for it.

## Docker

```sh
docker compose up --build
```

The container persists database and rotating logs in named volumes and runs as
non-root with dropped capabilities.
