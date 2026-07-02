#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")"

mode="${1:-}"
if [ -z "$mode" ]; then
  printf "Select deployment mode [local/server]: "
  read -r mode
fi

case "$mode" in
  local|server) ;;
  *)
    echo "Usage: ./init-deploy.sh local|server" >&2
    exit 2
    ;;
esac

if [ -f .env ]; then
  echo ".env already exists. Move it away or delete it before re-initializing." >&2
  exit 1
fi

secret() {
  bytes="${1:-24}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
    return
  fi
  if [ -r /dev/urandom ] && command -v od >/dev/null 2>&1; then
    od -An -N "$bytes" -tx1 /dev/urandom | tr -d ' \n'
    echo
    return
  fi
  date "+%s%N" | sha256sum | awk '{print $1}'
}

data_key="$(secret 32)"
admin_password="$(secret 18)"
backup_key="$(secret 32)"
update_secret="$(secret 32)"
collection_secret="$(secret 32)"
public_base_url="${CBMP_PUBLIC_BASE_URL:-http://localhost:8080}"

if [ "$mode" = "local" ]; then
  cat > .env <<EOF
CBMP_DEPLOY_MODE=local
CBMP_GATEWAY_PORT=8080
CBMP_PUBLIC_BASE_URL=$public_base_url
CBMP_ADDR=0.0.0.0:8088
CBMP_DATA=/data/app.vault
CBMP_DATA_KEY=$data_key
CBMP_BACKUP_KEY=$backup_key
CBMP_INITIAL_ADMIN_PASSWORD=$admin_password
CBMP_ENFORCE_IP_WHITELIST=0
CBMP_UPDATE_SIGNING_SECRET=$update_secret
CBMP_COLLECTION_CALLBACK_SECRET=$collection_secret
CBMP_TAX_GATEWAY_PROVIDER=
CBMP_TAX_GATEWAY_URL=
CBMP_TAX_GATEWAY_TOKEN=
CBMP_TAX_GATEWAY_SECRET=
CBMP_MAP_PROVIDER=
CBMP_MAP_TILE_URL=
CBMP_MAP_API_KEY=
EOF
  echo "Initialized local mode in ERP/deploy/.env"
  echo "Admin user: admin"
  echo "Admin password: $admin_password"
  echo "Start with: docker compose --env-file .env -f docker-compose.local.yml up -d --build"
  exit 0
fi

postgres_password="$(secret 18)"
rabbit_password="$(secret 18)"
minio_password="$(secret 18)"
postgres_user="cbmp"
postgres_db="cbmp"
rabbit_user="cbmp"
minio_user="cbmpadmin"

cat > .env <<EOF
CBMP_DEPLOY_MODE=server
CBMP_PUBLIC_BASE_URL=$public_base_url
CBMP_ADDR=0.0.0.0:8088
CBMP_DATA=/data/app.vault
CBMP_DATA_KEY=$data_key
CBMP_BACKUP_KEY=$backup_key
CBMP_INITIAL_ADMIN_PASSWORD=$admin_password
POSTGRES_DB=$postgres_db
POSTGRES_USER=$postgres_user
POSTGRES_PASSWORD=$postgres_password
CBMP_POSTGRES_DSN=postgres://$postgres_user:$postgres_password@postgres:5432/$postgres_db?sslmode=disable
CBMP_POSTGRES_LOAD_FROM_DOMAIN=0
CBMP_REDIS_ADDR=redis:6379
RABBITMQ_DEFAULT_USER=$rabbit_user
RABBITMQ_DEFAULT_PASS=$rabbit_password
CBMP_RABBITMQ_URL=amqp://$rabbit_user:$rabbit_password@rabbitmq:5672/
CBMP_CLICKHOUSE_HTTP_URL=http://clickhouse:8123
MINIO_ROOT_USER=$minio_user
MINIO_ROOT_PASSWORD=$minio_password
CBMP_ENFORCE_IP_WHITELIST=1
CBMP_UPDATE_SIGNING_SECRET=$update_secret
CBMP_COLLECTION_CALLBACK_SECRET=$collection_secret
CBMP_TAX_GATEWAY_PROVIDER=
CBMP_TAX_GATEWAY_URL=
CBMP_TAX_GATEWAY_TOKEN=
CBMP_TAX_GATEWAY_SECRET=
CBMP_MAP_PROVIDER=
CBMP_MAP_TILE_URL=
CBMP_MAP_API_KEY=
EOF

echo "Initialized server mode in ERP/deploy/.env"
echo "Admin user: admin"
echo "Admin password: $admin_password"
echo "Start with: docker compose --env-file .env -f docker-compose.enterprise.yml up -d --build"
