# --- Build stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /code
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/DropKey ./cmd/api

# --- Final stage ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/DropKey .
COPY cmd/api/.env .env

CMD ["./DropKey"]
