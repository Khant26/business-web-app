# Deployment

> **Status:** the hosting setup (which provider, domain name, server size, backup storage) has
> **not been decided yet** — it is a decision for the team and supervisor. This document describes
> a setup that works on any Linux server (VPS or cloud VM) with Docker. Adapt it once hosting is
> chosen, and record the final choice here.

## 1. What gets deployed

```
            Internet (HTTPS 443)
                   │
        ┌──────────▼──────────┐
        │  Reverse proxy       │  Caddy or Nginx: TLS certificates, HTTP→HTTPS,
        │  (on the host)       │  sets X-Forwarded-For
        └──────────┬──────────┘
                   │ http://127.0.0.1:8080
        ┌──────────▼──────────┐      volume "uploads" → /app/uploads
        │  app container       │◄──── (product images + private slips)
        │  Go API + built SPA  │
        └──────────┬──────────┘
                   │ db:5432 (Docker network)
        ┌──────────▼──────────┐      volume "db-data"
        │  db container        │◄──── (PostgreSQL 17 data)
        └─────────────────────┘
```

There is **one application image**. The root `Dockerfile` has three stages:

1. `node:24-alpine` — `npm ci` and `npm run build` the frontend.
2. `golang:1.26-alpine` — build a static binary `sawyun` from `./cmd/api` (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`).
3. `alpine:3.21` — copy the binary to `/app/sawyun` and the frontend to `/app/web`; run as a non-root `app` user; `ENV PORT=8080 WEB_DIR=/app/web UPLOAD_DIR=/app/uploads TZ=Asia/Bangkok`; expose 8080.

Migrations are inside the binary and run automatically on start.

## 1b. Split hosting: frontend on Vercel, API on the Google VM

This is the setup used for the presentation (3–5 users).

| Part | Where | How |
|---|---|---|
| Frontend | Vercel project with **Root Directory = `frontend`** (uses `frontend/vercel.json`) | Env var `VITE_API_ORIGIN=https://<api-subdomain>` |
| API | Google VM, container from `backend/Dockerfile` (API only), behind the existing Caddy | Reuses the shared Postgres (own database + role) |

Use two subdomains of the **same domain** (e.g. `shop.example.com` → Vercel CNAME and
`api-shop.example.com` → VM A record) so the login cookie works with the default `COOKIE_SAMESITE=lax`.
If the frontend stays on a `*.vercel.app` address instead, set `COOKIE_SAMESITE=none` (needs `COOKIE_SECURE=true`).

API environment for this setup:

```env
APP_ENV=production
DATABASE_URL=postgres://sawyun:<password>@<postgres-host>:5432/sawyun?sslmode=disable
OTP_SECRET=<long random string>
OTP_SHOW_CODE=true               # presentation: login code shown in a pop-up, no email needed
ADMIN_EMAIL=<your email>
COOKIE_SECURE=true
TRUST_PROXY=true                 # behind Caddy
CORS_ORIGIN=https://shop.example.com
UPLOAD_DIR=/app/uploads          # mount a volume here
```

Run the seed once inside the container for demo data: `docker exec <container> ./sawyun-seed`.
Turn `OTP_SHOW_CODE` off (and configure SMTP) before anyone outside the team uses the shop.

## 2. Build the image

```bash
docker build -t sawyun-ecommerce:latest .
```

or let Compose build it (next section). `.dockerignore` keeps `node_modules`, `dist`, `.env`,
uploads, logs and `.exe` files out of the build context.

## 3. Required environment

Create `.env` on the server (never commit it). Minimum for production:

```dotenv
APP_ENV=production
ADMIN_EMAIL=owner@yourshop.com
OTP_SECRET=<output of: openssl rand -hex 32>
COOKIE_SECURE=true
TRUST_PROXY=true            # only with a proxy that overwrites X-Forwarded-For (see below)

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=...
SMTP_PASSWORD=...
SMTP_FROM="Sawyun Ecommerce <no-reply@yourshop.com>"

# Optional (defaults shown)
SESSION_TTL_HOURS=720
SESSION_IDLE_HOURS=168
SESSION_MAX_PER_USER=5
```

`DATABASE_URL`, `WEB_DIR` and `UPLOAD_DIR` are set by `docker-compose.yml` for the `app` service.
`CORS_ORIGIN` is ignored in production. All variables are explained in [SETUP.md](SETUP.md#8-env-reference).

The API **refuses to start** if `APP_ENV=production` and `OTP_SECRET` is missing, shorter than 16
characters, or still the example value.

## 4. Run with Docker Compose

`docker-compose.yml` defines:

- `db` — `postgres:17-alpine`, volume `db-data`, health check `pg_isready`, published on host port 5433.
- `app` — built from the `Dockerfile`, in the **`app` profile** (so plain `docker compose up` in development does not start it), `env_file: .env`, forced `APP_ENV=production`, `DATABASE_URL` pointing at `db:5432`, port `8080:8080`, volume `uploads`, starts after `db` is healthy.

Start / update everything:

```bash
docker compose --profile app up -d --build
docker compose --profile app logs -f app      # watch start-up: "migrations applied", "listening"
```

### Production hardening of the compose file

The committed compose file is tuned for development (simple password, database port open on the
host). For a server, add a `docker-compose.prod.yml` next to it (this file does not exist yet —
create it during deployment) and start with both files:

```yaml
# docker-compose.prod.yml
services:
  db:
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports: !reset []                 # do not publish PostgreSQL on the host
    restart: unless-stopped
  app:
    environment:
      DATABASE_URL: postgres://sawyun:${POSTGRES_PASSWORD}@db:5432/sawyun?sslmode=disable
    ports: !override
      - "127.0.0.1:8080:8080"         # only the reverse proxy can reach the app
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/api/v1/health"]
      interval: 30s
      timeout: 5s
      retries: 3
```

```bash
# POSTGRES_PASSWORD is read from the shell or from .env
docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile app up -d --build
```

Notes:

- `POSTGRES_PASSWORD` only takes effect when the `db-data` volume is **created**. To change the password of an existing database, run `ALTER USER sawyun PASSWORD '…';` in `psql` first.
- `!reset` / `!override` need Docker Compose v2.24 or newer.
- Alpine includes BusyBox `wget`, so the health check works without extra packages.

## 5. Reverse proxy and HTTPS

The app speaks plain HTTP on port 8080. Put a reverse proxy in front for TLS. Then set
`COOKIE_SECURE=true` (secure cookie + HSTS) and, if the proxy **overwrites** `X-Forwarded-For`,
`TRUST_PROXY=true` so rate limits and logs use the real client IP.

> The app trusts the **first** address in `X-Forwarded-For`. If a proxy *appends* to a header sent
> by the client, the client controls the first entry and can fake its IP. Configure the proxy to
> **replace** the header, as in the examples below.

### Option A: Caddy (simplest — automatic Let's Encrypt certificates)

```caddyfile
# /etc/caddy/Caddyfile
shop.example.com {
    encode zstd gzip
    reverse_proxy 127.0.0.1:8080
}
```

Caddy obtains and renews the certificate, redirects HTTP to HTTPS, and sets `X-Forwarded-For` to the
real client IP (it ignores client-supplied values unless `trusted_proxies` is configured).

### Option B: Nginx

```nginx
server {
    listen 80;
    server_name shop.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name shop.example.com;

    ssl_certificate     /etc/letsencrypt/live/shop.example.com/fullchain.pem;   # from certbot
    ssl_certificate_key /etc/letsencrypt/live/shop.example.com/privkey.pem;

    client_max_body_size 6m;          # uploads are up to 5 MB

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;   # replace, do not append
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }
}
```

Get the certificate with `certbot --nginx -d shop.example.com`.

If a CDN or load balancer (e.g. Cloudflare) sits in front of the proxy, the proxy must take the real
client IP from that service's header before setting `X-Forwarded-For`; otherwise leave
`TRUST_PROXY=false` (rate limits will then apply per proxy IP).

## 6. Uploads volume

The `uploads` volume (mounted at `/app/uploads`) contains:

- `public/products`, `public/categories`, `public/events` — images referenced from the database;
- `private/slips` — customers' payment slips (personal financial data).

Treat it as **part of the database**: back it up together, restore it together, never delete it.
The container runs as user `app`; if you replace the volume with a host folder, make it writable by
that user.

## 7. Backups and restore

Back up **both** the database and the uploads volume, daily, and copy them **off the server**.
The volume names start with the Compose project name (by default the folder name, e.g.
`business-web-app_db-data`, `business-web-app_uploads`; check with `docker volume ls`).

### Database

```bash
# Backup (custom format, compressed)
docker compose exec -T db pg_dump -U sawyun -Fc sawyun > backups/sawyun-$(date +%F).dump

# Restore into the running database (overwrites existing objects)
docker compose --profile app stop app
docker compose exec -T db pg_restore -U sawyun -d sawyun --clean --if-exists < backups/sawyun-2026-10-29.dump
docker compose --profile app start app
```

### Uploads

```bash
# Backup
docker run --rm -v business-web-app_uploads:/data:ro -v "$PWD/backups":/backup alpine \
  tar czf /backup/uploads-$(date +%F).tgz -C /data .

# Restore
docker run --rm -v business-web-app_uploads:/data -v "$PWD/backups":/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/uploads-2026-10-29.tgz -C /data"
```

Automate both with a daily cron job, keep at least 7 daily and 4 weekly copies, and **test a
restore** on another machine before launch and every few months. A backup that has never been
restored is not a backup.

## 8. Upgrading

1. **Back up** the database and uploads (section 7).
2. Get the new code: `git pull` (or check out the release tag).
3. Rebuild and restart: `docker compose [-f … -f …] --profile app up -d --build`.
4. Watch the logs: you should see `migrations applied` and `listening`. New migrations run automatically before the server accepts requests.
5. Check the health endpoint and click through: home page, sign in, admin dashboard.

**If a migration fails**, the app exits and Docker keeps restarting it; the error is in the log. Fix
forward with a new commit, or roll back by restoring the backup and starting the previous image.
There is no automatic "down" migration in production.

Downtime is a few seconds while the container restarts. Orders in progress are safe: each change is
a database transaction.

## 9. Health check and monitoring

- `GET /api/v1/health` → `200 {"status":"ok"}`. It proves the process is running and serving requests; it does **not** query the database (a database outage shows up as `500` errors on other endpoints).
- Logs are JSON lines on stdout in production (`docker compose logs app`). Each request logs method, path, status, duration, IP and user id; errors are logged at `ERROR` level.
- Suggested external checks once hosting is chosen: an uptime monitor on `/api/v1/health` and on `/` (the home page), disk space alerts (uploads and database grow over time), and an alert if the daily backup file is missing.

## 10. First launch checklist

1. Server with Docker, firewall allowing only 22 (SSH), 80 and 443.
2. DNS `A` record for the shop domain → server IP.
3. `.env` created as in section 3; `docker-compose.prod.yml` added.
4. `docker compose … --profile app up -d --build`; proxy configured; HTTPS works.
5. Sign in with `ADMIN_EMAIL` (the code arrives by email).
6. **Settings:** shop name, real PromptPay ID and account name, VAT, payment window, COD limit, support contacts. Test-scan a QR with a real banking app.
7. **Shipping zones**, **categories**, **products** (with images and stock).
8. Add staff on **Users & staff**.
9. Place one real test order end-to-end (PromptPay, slip, approve, ship, deliver) and one COD order, then cancel them.
10. Set up and test backups. Go through the production checklist in [SECURITY.md](SECURITY.md#12-production-checklist).

Do **not** run `cmd/seed` on production: it creates demo products, demo coupons and a demo PromptPay ID.
