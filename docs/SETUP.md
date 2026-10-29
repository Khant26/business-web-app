# Local setup (Windows)

This guide sets up a development PC from zero. It was written on Windows 11 with Git Bash and
PowerShell; the same commands work on macOS/Linux in a normal terminal. Commands are shown for
**Git Bash** first and **PowerShell** where the syntax differs.

## 1. Install the tools

| Tool | Version | Check with | Notes |
|---|---|---|---|
| Git | any recent | `git --version` | Git for Windows also installs Git Bash. |
| Go | **1.26 or newer** | `go version` | `backend/go.mod` says `go 1.26.0`. |
| Node.js | **22+** (24 recommended) | `node --version` | `frontend/package.json` requires ≥ 20.19; the Docker build uses Node 24. npm 11 is fine. |
| Docker Desktop | recent | `docker compose version` | Enable the WSL 2 backend. Docker must be **running** before step 3. |
| Editor | VS Code recommended | | Useful extensions: Go, Vue - Official (Volar), ESLint, Prettier, EditorConfig. |

## 2. Get the code and create `.env`

```bash
git clone https://github.com/Khant26/business-web-app.git
cd business-web-app
cp .env.example .env
```

PowerShell: `Copy-Item .env.example .env`

The defaults in `.env.example` work for development without changes. `.env` is ignored by Git —
never commit it.

## 3. Start PostgreSQL

```bash
docker compose up -d db
docker compose ps          # "db" should become "healthy" within a few seconds
```

This starts **PostgreSQL 17** with user / password / database all set to `sawyun`, and publishes it
on **host port 5433**. We use 5433 (not the usual 5432) because the development PCs already run a
locally installed PostgreSQL on 5432; using a different port avoids connecting to the wrong
server. Data lives in the Docker volume `db-data`, so it survives `docker compose down`.

To open a SQL shell: `docker compose exec db psql -U sawyun sawyun`

## 4. Run the API

```bash
cd backend
go run ./cmd/api
```

On start the API:

1. reads `.env` (it looks for `.env` in the current folder, then `../.env`);
2. connects to PostgreSQL and applies any pending migrations (`migrations applied` in the log);
3. creates or re-activates the admin account for `ADMIN_EMAIL`;
4. creates `uploads/public` and `uploads/private/slips` under `UPLOAD_DIR`;
5. starts the background jobs and listens on `http://localhost:8080`.

Quick check: open <http://localhost:8080/api/v1/health> — it should return `{"status":"ok"}`.

> Run the API **from the `backend/` folder**. `UPLOAD_DIR=./uploads` and `WEB_DIR=../frontend/dist`
> are relative to the current folder, so running it elsewhere puts uploads in the wrong place.

## 5. Load demo data

In a second terminal:

```bash
cd backend
go run ./cmd/seed
```

The seed is **idempotent**: running it again does not create duplicates and does not reset stock.
It creates 6 categories, 16 products with variants and stock, 2 events (one running now, one
starting in 20 days — their dates are reset each time the seed runs), 3 shipping zones, coupons
`WELCOME10` and `SAVE50`, and fills empty settings with a demo PromptPay ID `0812345678`, account
name "Sawyun Shop", support email and phone. Seeded products have **no images**; a placeholder is
shown until you upload some in the admin area. See [DATABASE.md](DATABASE.md#5-seed-data).

## 6. Run the website

In a third terminal:

```bash
cd frontend
npm ci
npm run dev
```

Open <http://localhost:5173>. Vite proxies `/api/*` and `/uploads/*` to `http://localhost:8080`
(set `VITE_BACKEND_URL` to proxy somewhere else). Because the browser sees only one origin, cookies
and CSRF work without any CORS setup.

## 7. Sign in

1. Click **Sign in**, enter `admin@example.com`, click **Send code**.
2. Look at the **API terminal**. You will see a line like
   `level=INFO msg="OTP email (not sent, logged)" to=admin@example.com code=482913`.
3. Type the 6 digits. You are now signed in as admin; the **Admin** link appears in the header.

For a customer, use any other email (for example `customer@example.com`).

## 8. `.env` reference

Values are read from real environment variables first; `.env` only fills in variables that are not
already set. Empty values fall back to the default. In `.env`, a ` #` after a value starts a comment,
and values may be wrapped in quotes.

### App

| Variable | Default | Meaning |
|---|---|---|
| `APP_ENV` | `development` | `production` switches on: JSON logs, real SMTP sending, the strict `OTP_SECRET` check, and always serving the SPA. Anything other than `production` counts as development. |
| `PORT` | `8080` | HTTP port of the Go server. |
| `PUBLIC_URL` | `http://localhost:5173` | The public address of the shop. Loaded into config for future use (e.g. links in emails); currently not used by any handler. |
| `WEB_DIR` | `../frontend/dist` | Folder with the built frontend. In development it is served only if the folder exists; in production it is always served (`/app/web` in Docker). |
| `UPLOAD_DIR` | `./uploads` | Root folder for uploads: `public/` (product, category, event images) and `private/slips/` (payment slips). |

### Database

| Variable | Default | Meaning |
|---|---|---|
| `DATABASE_URL` | `postgres://sawyun:sawyun@localhost:5433/sawyun?sslmode=disable` | pgx connection string. Inside Docker Compose the app uses `db:5432` instead (set in `docker-compose.yml`). |

### Auth

| Variable | Default | Meaning |
|---|---|---|
| `ADMIN_EMAIL` | `admin@example.com` | This email is created or promoted to an **active admin** on every API and seed start. Changing it later does **not** demote the old admin. |
| `COOKIE_SECURE` | `false` | Adds the `Secure` flag to the session cookie **and** sends the HSTS header. Must be `true` behind HTTPS and `false` on plain `http://localhost` (otherwise the browser drops the cookie). |
| `SESSION_TTL_HOURS` | `720` (30 days) | Absolute lifetime of a session. |
| `SESSION_IDLE_HOURS` | `168` (7 days) | A session not used for this long is invalid. |
| `SESSION_MAX_PER_USER` | `5` | Signing in on a 6th device removes the oldest session. Minimum 1. |
| `OTP_SECRET` | `change-me-to-a-long-random-string` | HMAC key for hashing OTP codes. In production it must be **at least 16 characters and not the default**, or the API refuses to start. |

### SMTP (sign-in code emails)

| Variable | Default | Meaning |
|---|---|---|
| `SMTP_HOST` | empty | Mail server, e.g. `smtp.gmail.com`. Empty = codes are only printed in the API log. |
| `SMTP_PORT` | `587` | `587` = STARTTLS, `465` = implicit TLS. Other ports use STARTTLS if offered (required except for `localhost`). |
| `SMTP_USERNAME` | empty | Login; if empty, no authentication is attempted. |
| `SMTP_PASSWORD` | empty | Password or app password. |
| `SMTP_FROM` | `Sawyun Ecommerce <no-reply@example.com>` | Sender, must be a valid address (`Name <addr>` form is fine). |

> **How sending works:** when `SMTP_HOST` is set, real emails are sent in every mode, so you can
> test your mail setup locally. In development the code is **also** printed in the API log, and if the
> mail server rejects the message the error is logged but sign-in still works. In production an SMTP
> error makes the OTP request fail, so misconfiguration is noticed.

### CORS / proxy

| Variable | Default | Meaning |
|---|---|---|
| `CORS_ORIGIN` | empty in code (`.env.example` has `http://localhost:5173`) | One origin allowed to make credentialed cross-origin requests. Active **only in development**. Not needed when you use the Vite proxy. |
| `TRUST_PROXY` | `false` | When `true`, the client IP is taken from the **first** `X-Forwarded-For` entry. Turn on only behind a reverse proxy that overwrites that header (see [DEPLOYMENT.md](DEPLOYMENT.md#5-reverse-proxy-and-https)). |

### Frontend build-time variable

| Variable | Default | Meaning |
|---|---|---|
| `VITE_BACKEND_URL` | `http://localhost:8080` | Where `npm run dev` / `npm run preview` proxy `/api` and `/uploads`. Read by `vite.config.ts`, not by the browser code. |

## 9. Sending real emails with Gmail (optional)

You only need this to test real delivery or for production.

1. Use a Google account with **2-Step Verification** turned on.
2. Go to Google Account → Security → **App passwords**, create one (name it "Sawyun"), and copy the 16-character password.
3. Set in `.env`:

   ```dotenv
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=yourshop@gmail.com
   SMTP_PASSWORD=abcdefghijklmnop        # the app password, without spaces
   SMTP_FROM="Sawyun Ecommerce <yourshop@gmail.com>"
   ```

4. Run the API in production mode to actually send, for example in Git Bash:

   ```bash
   cd backend
   APP_ENV=production OTP_SECRET=$(openssl rand -hex 32) go run ./cmd/api
   ```

   PowerShell: `$env:APP_ENV="production"; $env:OTP_SECRET="<64 random hex chars>"; go run ./cmd/api`

   In production mode the API also serves `../frontend/dist` on port 8080, so run `npm run build`
   first or keep using the Vite dev server on 5173.

Gmail replaces the From address with your Gmail address unless you have set up an alias. Gmail has
daily sending limits, so for a real shop consider a transactional mail service (any SMTP provider
works).

## 10. Running tests

### Backend

```bash
cd backend
go vet ./...
gofmt -l .            # prints nothing when all files are formatted
go test ./...
```

`go test ./...` runs the unit tests (pricing, coupon validation, order transitions, PromptPay
payload/CRC, middleware, rate limiter, uploads, config parsing, provinces). The **integration test**
(`internal/server/integration_test.go`) is skipped unless `TEST_DATABASE_URL` is set. It places a
real order, expires it and checks that stock and coupon usage are restored. It writes test rows, so
point it at a **separate, disposable database**:

```bash
docker compose exec db createdb -U sawyun sawyun_test     # once
TEST_DATABASE_URL="postgres://sawyun:sawyun@localhost:5433/sawyun_test?sslmode=disable" \
  go test ./internal/server/ -run Integration -v
```

PowerShell:

```powershell
$env:TEST_DATABASE_URL = "postgres://sawyun:sawyun@localhost:5433/sawyun_test?sslmode=disable"
go test ./internal/server/ -run Integration -v
```

### Frontend

```bash
cd frontend
npm run lint
npm run typecheck
npm test              # Vitest (jsdom)
npm run build
npm run format:check
```

The i18n test fails if English and Myanmar files have different keys or a message has invalid
vue-i18n syntax — see [DEVELOPMENT.md](DEVELOPMENT.md#4-adding-a-translation-key).

## 11. Common problems

| Symptom | Cause | Fix |
|---|---|---|
| API log: `ping database: … connection refused` | Docker Desktop or the `db` container is not running. | Start Docker Desktop, then `docker compose up -d db`. |
| API connects but tables are missing / login fails strangely | You connected to the **local** PostgreSQL on 5432 instead of Docker. | Keep `DATABASE_URL` on port **5433**. If 5433 is also taken, change the left side of `"5433:5432"` in `docker-compose.yml` **and** the port in `DATABASE_URL`. |
| `docker compose up` says port 5433 is already allocated | Another container or program uses it. | `docker ps` to find it, stop it, or change the port as above. |
| `npm ci` warns that install scripts were skipped, or `npm run dev` fails with an esbuild error | Newer npm versions only run dependency install scripts that are allowed. `package.json` allows `"esbuild@0.28.2"` in `allowScripts`. After a dependency update the esbuild version may change. | Run `npm ls esbuild` to see the version, update the `allowScripts` entry to match, then `npm ci` again (or `npm rebuild esbuild`). |
| `npm run dev` fails: port 5173 is in use | `strictPort` is on, so Vite will not pick another port (the CORS and proxy setup assume 5173). | Close the other Vite process. |
| No OTP email arrives | In development codes are **never emailed**. | Read the code from the API terminal (`OTP email (not sent, logged)`). |
| `Too many attempts. Try again in N seconds` when requesting a code | OTP request limits: **1 per 60 s per email**, 5 per hour per email, 20 per hour per IP. The login page also shows a 60-second resend timer. | Wait. Limits are kept in memory, so restarting the API clears them in development. |
| `Too many wrong codes` | 5 wrong attempts burn the code. Verify is also limited to 10 tries per 15 minutes per IP. | Request a new code (after the 60-second wait). |
| Signed in successfully but immediately signed out / `/auth/me` returns 401 | `COOKIE_SECURE=true` on plain HTTP, so the browser refuses the cookie. | Use `COOKIE_SECURE=false` locally. |
| `403 csrf` from Postman/curl | Every non-GET API request needs the header `X-Requested-With: sawyun`. | Add the header. The frontend client adds it automatically. |
| CORS error in the browser console | You opened the site from an origin other than the Vite dev server (for example a different port) and it calls the API directly. | Use <http://localhost:5173> (the proxy avoids CORS), or set `CORS_ORIGIN` to that exact origin. CORS is disabled in production because the API and site share one origin. |
| `OTP_SECRET must be set to a long random value in production` | `APP_ENV=production` with the default or a short secret. | Set `OTP_SECRET` to 16+ random characters, e.g. `openssl rand -hex 32`. |
| Uploaded images return 404 | The API was started from another folder, so `UPLOAD_DIR=./uploads` points somewhere else. | Start the API from `backend/`, or use an absolute `UPLOAD_DIR`. |
| "frontend not built" on <http://localhost:8080> | Production mode serves `WEB_DIR`, which is empty. | `cd frontend && npm run build`, or use the dev server on 5173. |
| Myanmar text shows boxes | Fonts are bundled (Noto Sans Myanmar via `@fontsource`), so this usually means `npm ci` did not finish. | Re-run `npm ci`. |
