#!/bin/bash
# Setup staging environment on the server
# Run once: ssh root@194.87.133.93 'bash -s' < scripts/setup-staging.sh
set -euo pipefail

STAGING_DIR="/opt/3d-print-shop-staging"
PROD_ENV="/opt/3d-print-shop/.env"

echo "=== 1. Create staging database ==="
docker exec shop-postgres psql -U shop_user -d postgres -c \
  "SELECT 1 FROM pg_database WHERE datname='shop_db_staging'" | grep -q 1 && \
  echo "Database already exists" || \
  docker exec shop-postgres psql -U shop_user -d postgres -c \
    "CREATE DATABASE shop_db_staging OWNER shop_user;"

echo "=== 2. Create staging directory structure ==="
mkdir -p "$STAGING_DIR"/{admin,app/frontend,app/admin}

echo "=== 3. Create staging .env ==="
if [ ! -f "$STAGING_DIR/.env" ]; then
  # Extract S3 and SMTP values from prod .env
  S3_ENDPOINT=$(grep '^S3_ENDPOINT=' "$PROD_ENV" | cut -d= -f2-)
  S3_REGION=$(grep '^S3_REGION=' "$PROD_ENV" | cut -d= -f2-)
  S3_BUCKET=$(grep '^S3_BUCKET=' "$PROD_ENV" | cut -d= -f2-)
  S3_ACCESS_KEY=$(grep '^S3_ACCESS_KEY=' "$PROD_ENV" | cut -d= -f2-)
  S3_SECRET_KEY=$(grep '^S3_SECRET_KEY=' "$PROD_ENV" | cut -d= -f2-)
  S3_PUBLIC_URL=$(grep '^S3_PUBLIC_URL=' "$PROD_ENV" | cut -d= -f2-)

  cat > "$STAGING_DIR/.env" <<ENVEOF
ENV=staging
PORT=8081
DATABASE_URL=postgresql://shop_user:shop_pass@localhost:5432/shop_db_staging?sslmode=disable
REDIS_URL=redis://localhost:6379/1
JWT_SECRET=staging-secret-$(openssl rand -hex 16)
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
ALLOWED_ORIGINS=https://test.avangard-print.ru
TELEGRAM_WEBAPP_URL=https://test.avangard-print.ru
S3_ENDPOINT=$S3_ENDPOINT
S3_REGION=$S3_REGION
S3_BUCKET=$S3_BUCKET
S3_ACCESS_KEY=$S3_ACCESS_KEY
S3_SECRET_KEY=$S3_SECRET_KEY
S3_PUBLIC_URL=$S3_PUBLIC_URL
ENVEOF
  echo "Created $STAGING_DIR/.env"
else
  echo ".env already exists, skipping"
fi

echo "=== 4. Create systemd services ==="

cat > /etc/systemd/system/3d-print-shop-staging.service <<'EOF'
[Unit]
Description=3D Print Shop API (Staging)
After=network.target docker.service

[Service]
Type=simple
WorkingDirectory=/opt/3d-print-shop-staging
ExecStart=/opt/3d-print-shop-staging/api
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/3d-print-frontend-staging.service <<'EOF'
[Unit]
Description=3D Print Frontend (Staging)
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/3d-print-shop-staging/app/frontend
ExecStart=/usr/bin/npm start -- -p 3001
Restart=on-failure
RestartSec=5
Environment=PORT=3001

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
echo "Systemd services created"

echo "=== 5. Create nginx config ==="

cat > /etc/nginx/sites-available/test.avangard-print.ru <<'NGINX'
server {
    listen 80;
    server_name test.avangard-print.ru;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name test.avangard-print.ru;

    # SSL certs will be added by certbot
    ssl_certificate /etc/letsencrypt/live/avangard-print.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/avangard-print.ru/privkey.pem;

    client_max_body_size 10M;

    # Frontend (Next.js on port 3001)
    location / {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # API
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Webhook
    location /webhook/ {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Admin SPA
    location /admin/ {
        alias /opt/3d-print-shop-staging/admin/;
        try_files $uri $uri/ /admin/index.html;
    }
}
NGINX

# Enable site
ln -sf /etc/nginx/sites-available/test.avangard-print.ru /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
echo "Nginx configured"

echo "=== 6. SSL certificate ==="
certbot --nginx -d test.avangard-print.ru --non-interactive --agree-tos --redirect 2>/dev/null || \
  echo "Certbot failed or cert already exists. Run manually: certbot --nginx -d test.avangard-print.ru"

echo ""
echo "=== Setup complete! ==="
echo "Next: deploy backend, frontend, admin, and run migrations"
echo "Use: bash scripts/deploy-staging.sh"
