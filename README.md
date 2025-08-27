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

This application supports both PostgreSQL and SQLite databases. You can choose your database by setting the `DATABASE_TYPE` environment variable.

### PostgreSQL (Default)
Set `DATABASE_TYPE=postgres` in your `.env` file and configure the PostgreSQL connection settings.

### SQLite
Set `DATABASE_TYPE=sqlite` in your `.env` file. SQLite uses a local file database and requires minimal configuration.

Example SQLite configuration:
```bash
DATABASE_TYPE=sqlite
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
