#!/bin/sh
set -e

echo "🚀 Running database migrations..."
/app/roaming-document migrate up

echo "✅ Migrations complete. Starting application..."
exec /app/roaming-document
