# yakamoz

Yakamoz is a Go modular monolith scaffold for topics and authors. It uses only the
standard `net/http` router and keeps each domain's entity, repository port, service,
and HTTP DTOs together.

## Run locally

Go 1.24+ is required. The zero-dependency mode is the default:

```sh
STORE=memory go run ./cmd/api
```

The API listens on `:8080`. Health checks are `/healthz` and `/readyz`.
`POST /api/v1/topics/{id}/translate` uses a deterministic stub translator by default.

## Run with Postgres

The repository implementation uses `database/sql` with PostgreSQL `$n` placeholders,
so both normal and tagged builds remain self-contained when module downloads are
unavailable. The PostgreSQL driver is intentionally an opt-in runtime dependency.
Once network access is available, enable the PostgreSQL runtime with:

```sh
go get github.com/jackc/pgx/v5@v5.7.5
```

Add this import to `cmd/api/main.go`:

```go
import _ "github.com/jackc/pgx/v5/stdlib"
```

Then run:

```sh
go mod tidy
docker compose up -d postgres
STORE=postgres DATABASE_URL='postgres://yakamoz:yakamoz@localhost:5432/yakamoz?sslmode=disable' go run ./cmd/api
```

Migrations in `migrations/` run at startup. Postgres implementations are behind the
same repository interfaces as the in-memory implementations.

## Extracting a domain

Copy `internal/topic` (or `internal/author`) and its repository interface to a new
service, then replace the composition root with that module's HTTP handler. Keep
business rules in the service and move the repository implementation behind its
interface. The HTTP DTOs are the external contract; UUID references such as
`topic.CreatedBy` do not require importing another domain. Replace the AI
`Translator` port with a client adapter in the new service without changing topic
business logic.
