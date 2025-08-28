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
