#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
ENV_FILE="$ROOT/.env"

random_secret() {
  openssl rand -base64 "$1" | tr '+/' '-_' | tr -d '=\n'
}

if [ ! -f "$ENV_FILE" ]; then
  admin_password="$(random_secret 24)"
  cat >"$ENV_FILE" <<EOF
APP_ADMIN_USERNAME=admin
APP_ADMIN_PASSWORD=$admin_password
JWT_SECRET=$(random_secret 48)
SCA_SERVICE_TOKEN=$(random_secret 32)
POSTGRES_DB=devsecurity
POSTGRES_USER=devsecurity
POSTGRES_PASSWORD=$(random_secret 24)
S3_ACCESS_KEY=devsecurity
S3_SECRET_KEY=$(random_secret 24)
S3_BUCKET=devsecurity
MINIO_ROOT_USER=devsecurity
MINIO_ROOT_PASSWORD=$(random_secret 24)
GIT_ALLOWED_HOSTS=localhost,gitlab.com
GRAFANA_ADMIN_PASSWORD=$(random_secret 24)
EOF
  chmod 600 "$ENV_FILE"
  printf 'Created .env with random secrets.\nAdmin username: admin\nAdmin password: %s\n' "$admin_password"
fi

if ! grep -q '^SCA_SERVICE_TOKEN=' "$ENV_FILE"; then
  printf 'SCA_SERVICE_TOKEN=%s\n' "$(random_secret 32)" >>"$ENV_FILE"
  chmod 600 "$ENV_FILE"
fi

cd "$ROOT"
exec docker compose up --build
