# include variables from the .envrc file
include .env

##############
##  HELPERS ##
##############

## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	 @sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


##################
##  DEVELOPMENT ##
##################

# Build the application
all: build test

templ-install:
	go install github.com/a-h/templ/cmd/templ@latest

sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

tailwind-install:
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64 -o tailwindcss; fi
	@chmod +x tailwindcss

install-deps: templ-install sqlc-install goose-install tailwind-install
	@go mod tidy

# Watch Tailwind CSS changes
tailwind: tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

.PHONY: templ
templ:
    templ generate --watch --proxy="http://localhost:8090" --open-browser=false

.PHONY: migrate
migrate:
	goose up

build: tailwind-install templ-install
	@echo "Building..."
	@templ generate
	@./tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Run tests with testcontainers (slower but isolated)
test-container:
	@echo "Testing with testcontainers..."
	@go test ./... -v -count=1

# Start a local test database for faster testing
test-db-start:
	@echo "Starting test PostgreSQL database..."
	@docker run -d \
		--name test-postgres \
		-e POSTGRES_USER=testuser \
		-e POSTGRES_PASSWORD=testpass \
		-e POSTGRES_DB=testdb \
		-p 5433:5432 \
		postgres:16-alpine
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Test database started on port 5433"
	@echo "Run tests with: make test-fast"

# Stop the test database
test-db-stop:
	@echo "Stopping test database..."
	@docker stop test-postgres || true
	@docker rm test-postgres || true

# Run tests with existing test database (faster)
test-fast:
	@echo "Testing with existing database (fast mode)..."
	@TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" go test ./... -v -count=1

# Run specific test
test-one:
	@echo "Running specific test..."
	@TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable" go test -v ./... -run $(TEST_NAME) -count=1

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch tailwind-install templ-install test-container test-db-start test-db-stop test-fast test-one

# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

###########################
### QUALITY CONTROL ######
###########################

## audit: tidy dependencies and format, ver and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying modules dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vetting code....'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vender: tidy and vendor dependencies
.PHONY: vendor
vender:
	@echo 'Tidying and verifiying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vender


# TODO
#############################
#### PRODUCTION #############
#############################

production_host_ip = '45.55.49.87'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	shh greelight@${production_host_ip}

## production/deploy/api: deploy the api production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -rP --delete ./bin/linux_amd64/api ./migrations greenlight@${production_host_ip}:~
	ssh -t greenlight@${production_host_ip} 'migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up'


## production/configure/api.service: configure the production systemd api.service file
.PHONY: production/configure/api.service 
production/configure/api.service: 
	rsync -P ./remote/production/api.service greenlight@${production_host_ip}:~ 
	ssh -t greenlight@${production_host_ip} '\ 
	  sudo mv ~/api.service /etc/systemd/system/ \ 
	  && sudo systemctl enable api \ 
	  && sudo systemctl restart api \
	  '


## production/configure/caddyfile: configure the production Caddyfile 
.PHONY: production/configure/caddyfile 
production/configure/caddyfile: 
	rsync -P ./remote/production/Caddyfile greenlight@${production_host_ip}:~ 
	ssh -t greenlight@${production_host_ip} '\ 
		sudo mv ~/Caddyfile /etc/caddy/ \ 
		&& sudo systemctl reload caddy \
	'




