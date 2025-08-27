# Go Web Starter


## Getting Started

### Included
- TemplUI
- Templ
- HTMX
- PostgreSQL or SQLite
- TailwindCss
- Docker
- Sqlc
- Goose

## Database Configuration

This application supports both PostgreSQL and SQLite databases. Preferred configuration uses the `BLUEPRINT_DB_*` environment variables. For database type detection, the application reads `BLUEPRINT_DB_TYPE` first, falling back to `DATABASE_TYPE` for backward compatibility.

### PostgreSQL (Default)
Set `BLUEPRINT_DB_TYPE=postgres` (or legacy `DATABASE_TYPE=postgres`) in your `.env` file and configure the PostgreSQL connection settings.

### SQLite
Set `BLUEPRINT_DB_TYPE=sqlite` (or legacy `DATABASE_TYPE=sqlite`). SQLite uses a local file database and requires minimal configuration.

Example SQLite configuration:
```bash
BLUEPRINT_DB_TYPE=sqlite
BLUEPRINT_DB_URL=./database.sqlite
```

## MakeFile

Apply migrations to the database
```bash
make migrate
# or
go run cmd/api/main.go migrate
```

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```

Live reload the application:

```bash
make watch
```

Mailhog test mail server
- Install mailhog for your OS
- localhost:8025
```bash
mailhog
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
