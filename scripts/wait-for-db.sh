#!/bin/sh
set -e
HOST="${DB_HOST:-db}"
PORT="${DB_PORT:-5432}"
echo "Waiting for Postgres at ${HOST}:${PORT}..."
until nc -z "$HOST" "$PORT" >/dev/null 2>&1; do
  sleep 1
done
echo "Database is up. Starting application..."
