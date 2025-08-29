FROM golang:1.24.5-alpine AS build
RUN apk add --no-cache curl libstdc++ libgcc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
ARG TARGETOS
ARG TARGETARCH
RUN case "$TARGETARCH" in \
    "arm64") curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64-musl -o tailwindcss ;; \
    "amd64") curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64-musl -o tailwindcss ;; \
    *) echo "Unsupported architecture: $TARGETARCH" && exit 1 ;; \
  esac
RUN chmod +x tailwindcss
RUN ./tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css

RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main","server"]


