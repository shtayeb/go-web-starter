# Go Web Starter


## Getting Started

### Included
- TemplUI
- Templ
- HTMX
- Postgres
- TailwindCss
- Docker
- Sqlc
- Goose

## MakeFile

Apply migrations to the database
```bash
goose up
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

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
