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

templ:
    templ generate --watch --proxy="http://localhost:8090" --open-browser=false

# Watch Tailwind CSS changes
tailwind: tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi

sqlc-install:
	@if ! command -v sqlc> /dev/null; then \
		read -p "Go's 'sqlc' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "sqlc installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install sqlc. Exiting..."; \
			exit 1; \
		fi; \
	fi

goose-install:
	@if ! command -v goose> /dev/null; then \
		read -p "Go's 'goose' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/pressly/goose/v3/cmd/goose@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "goose installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install goose. Exiting..."; \
			exit 1; \
		fi; \
	fi

tailwind-install:
	
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64 -o tailwindcss; fi
	@chmod +x tailwindcss

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

.PHONY: all build run test clean watch tailwind-install templ-install

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




