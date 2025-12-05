# --- build stage ---
FROM golang:1.24-bullseye AS builder
WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарник из cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api

# --- runtime stage ---
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/server .
COPY scripts/wait-for-db.sh /usr/local/bin/wait-for-db.sh

# Install runtime dependencies
RUN apk add --no-cache ca-certificates netcat-openbsd \
    && chmod +x /usr/local/bin/wait-for-db.sh

EXPOSE 8080

CMD ["/bin/sh","-c","wait-for-db.sh && ./server"]
