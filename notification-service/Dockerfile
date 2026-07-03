# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo o código-fonte
COPY . .

# Build do binário
RUN go build -o notification-service ./cmd

# Run stage
FROM alpine:latest

WORKDIR /app

# Copiar o binário compilado
COPY --from=builder /app/notification-service .

# Executar a aplicação
CMD ["./notification-service"]
