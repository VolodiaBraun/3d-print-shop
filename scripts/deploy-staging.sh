#!/bin/bash
# Deploy to staging environment (test.avangard-print.ru)
# Usage: bash scripts/deploy-staging.sh [backend|frontend|admin|all]
set -euo pipefail

SERVER="root@194.87.133.93"
STAGING_DIR="/opt/3d-print-shop-staging"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
COMPONENT="${1:-all}"

SSH_OPTS="-o ConnectTimeout=10 -o ServerAliveInterval=10"

deploy_backend() {
  echo "=== Backend: cross-compile ==="
  cd "$PROJECT_DIR/backend"
  GOOS=linux GOARCH=amd64 go build -o /tmp/api-staging ./cmd/api
  echo "Binary built: $(ls -lh /tmp/api-staging | awk '{print $5}')"

  echo "=== Backend: upload ==="
  scp $SSH_OPTS /tmp/api-staging "$SERVER:$STAGING_DIR/api.new"

  echo "=== Backend: migrations ==="
  # Upload migrations
  scp $SSH_OPTS -r "$PROJECT_DIR/backend/migrations" "$SERVER:$STAGING_DIR/app/"

  ssh $SSH_OPTS "$SERVER" bash -c "'
    # Run migrations
    if command -v migrate &>/dev/null; then
      DB_URL=\$(grep \"^DATABASE_URL=\" $STAGING_DIR/.env | cut -d= -f2-)
      migrate -path $STAGING_DIR/app/migrations -database \"\$DB_URL\" up
    else
      echo \"WARNING: migrate CLI not found, skipping migrations\"
    fi

    # Deploy binary
    systemctl stop 3d-print-shop-staging 2>/dev/null || true
    mv $STAGING_DIR/api.new $STAGING_DIR/api
    chmod +x $STAGING_DIR/api
    systemctl start 3d-print-shop-staging
    systemctl enable 3d-print-shop-staging 2>/dev/null || true
    sleep 2
    systemctl is-active 3d-print-shop-staging
  '"
  echo "Backend deployed!"
}

deploy_frontend() {
  echo "=== Frontend: upload source ==="
  # Sync frontend source (excluding node_modules and .next)
  rsync -avz --delete \
    --exclude 'node_modules' \
    --exclude '.next' \
    --exclude '.env.local' \
    -e "ssh $SSH_OPTS" \
    "$PROJECT_DIR/frontend/" "$SERVER:$STAGING_DIR/app/frontend/"

  echo "=== Frontend: build on server ==="
  ssh $SSH_OPTS "$SERVER" bash -c "'
    cd $STAGING_DIR/app/frontend

    # Create .env.local for staging
    echo \"NEXT_PUBLIC_API_URL=https://test.avangard-print.ru/api/v1\" > .env.local

    npm install --legacy-peer-deps 2>/dev/null
    npx next build

    systemctl restart 3d-print-frontend-staging
    systemctl enable 3d-print-frontend-staging 2>/dev/null || true
    sleep 3
    systemctl is-active 3d-print-frontend-staging
  '"
  echo "Frontend deployed!"
}

deploy_admin() {
  echo "=== Admin: upload source ==="
  rsync -avz --delete \
    --exclude 'node_modules' \
    --exclude 'dist' \
    -e "ssh $SSH_OPTS" \
    "$PROJECT_DIR/admin/" "$SERVER:$STAGING_DIR/app/admin/"

  echo "=== Admin: build on server ==="
  ssh $SSH_OPTS "$SERVER" bash -c "'
    cd $STAGING_DIR/app/admin

    # Set API URL for staging build
    export VITE_API_URL=https://test.avangard-print.ru/api/v1

    npm install 2>/dev/null
    npm run build
    cp -r dist/* $STAGING_DIR/admin/
  '"
  echo "Admin deployed!"
}

case "$COMPONENT" in
  backend)
    deploy_backend
    ;;
  frontend)
    deploy_frontend
    ;;
  admin)
    deploy_admin
    ;;
  all)
    deploy_backend
    deploy_frontend
    deploy_admin
    ;;
  *)
    echo "Usage: $0 [backend|frontend|admin|all]"
    exit 1
    ;;
esac

echo ""
echo "=== Verification ==="
echo "API:      curl https://test.avangard-print.ru/api/v1/ping"
echo "Frontend: https://test.avangard-print.ru"
echo "Admin:    https://test.avangard-print.ru/admin/"
