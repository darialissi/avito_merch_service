#!/usr/bin/env bash

# Скрипт для генерации .env из .config.{dev|prod|test}.yml на основе APP_MODE
# Необходим для поднятия сервисов в docker compose

set -euo pipefail

CONFIG_PATH="$(pwd)/configs/.config.$APP_MODE.yml"

if ! command -v yq >/dev/null 2>&1; then
  cat >&2 <<'EOF'
yq not found. Install it for your platform:
  macOS (Homebrew):   brew install yq
  Ubuntu/Debian:      sudo snap install yq
  Fedora:             sudo dnf install -y yq
  Arch:               sudo pacman -S yq
  Windows (choco):    choco install yq
  Windows (scoop):    scoop install yq
EOF
  exit 1
fi

db_host="$(yq -r '.db.host' "$CONFIG_PATH")"
db_port="$(yq -r '.db.port' "$CONFIG_PATH")"
db_name="$(yq -r '.db.name' "$CONFIG_PATH")"
db_user="$(yq -r '.db.user' "$CONFIG_PATH")"
db_password="$(yq -r '.db.password' "$CONFIG_PATH")"

container_db_host="db_$APP_MODE"
container_db_port=5432

redis_host="$(yq -r '.redis.host' "$CONFIG_PATH")"
redis_port="$(yq -r '.redis.port' "$CONFIG_PATH")"
redis_password="$(yq -r '.redis.password' "$CONFIG_PATH")"

cat > .env <<EOF
CONFIG_PATH=$CONFIG_PATH # local

POSTGRES_DB=$db_name
POSTGRES_USER=$db_user
POSTGRES_PASSWORD=$db_password

GOOSE_DRIVER=postgres
GOOSE_MIGRATION_DIR=./migrations
GOOSE_DBSTRING=user=$db_user password=$db_password dbname=$db_name host=$container_db_host port=$container_db_port sslmode=disable

REDIS_PASSWORD=$redis_password
EOF

echo ".env generated from $CONFIG_PATH"

# Если APP_MODE=test/dev — необходимо установить CONFIG_PATH в окружение
if [[ $APP_MODE == "test" || $APP_MODE == "dev" ]]; then
  echo "======================================================================================="
  echo "Для локального запуска приложения необходимо установить переменную в окружение:"
  echo "export CONFIG_PATH=${CONFIG_PATH}"
  echo "======================================================================================="
fi
