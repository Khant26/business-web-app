# Sawyun Ecommerce

Sawyun Ecommerce is an online shop for a small business that sells everyday goods inside Thailand.
Customers browse products in **English or Myanmar**, pay by **PromptPay QR** (and upload a transfer
slip) or **cash on delivery**, and follow their order until it arrives. Shop staff run everything
from a built-in **admin area**: orders, payment slips, products, stock, coupons, events, shipping
zones and settings.

The whole app is one Go server that serves both the JSON API and the built Vue website, backed by
PostgreSQL. It runs on a laptop with Docker for development and as a single Docker image in
production.

> **Screenshots:** none are committed yet. When the team takes them, put them in `docs/images/` and
> link them here (suggested: home page, product page, checkout, PromptPay page, admin dashboard,
> admin order detail — at both 375 px and 1366 px width).

---

## Features

**Storefront (customers)**

- Bilingual UI: English (default) and Myanmar, switchable at any time; product data is stored in both languages.
- Home page with running events (countdown banners), categories, popular, featured and new products.
- Product list with search, category / price / in-stock filters and sorting; product pages with image gallery, variants (size, colour…) and reviews.
- Cart kept in the browser (works without signing in), re-validated against live stock and prices.
- Password-less sign-in with a 6-digit **email OTP** code.
- Checkout with saved addresses (Thai province → district → subdistrict picker), coupon codes, shipping fee by province zone and **7% VAT** added on top.
- **PromptPay QR** generated for the exact order total, slip upload, then manual verification by staff. **Cash on delivery** for totals up to a configurable limit.
- Order history with a status timeline, cancel button, courier tracking link, and reviews after delivery.
- Wishlist, address book, profile, and a "Security" page that lists signed-in devices and can sign them out.

**Admin area (staff and admins)**

- Dashboard: today's sales, slips to review, orders to pack, low stock, 7-day sales chart.
- Orders: filters, slip approve / reject, status changes (packing → shipped → delivered, cancel), courier tracking.
- Products with variants and images, categories, events, coupons, review moderation.
- Inventory with stock adjustments and a full movement history.
- Admin-only: shipping zones, shop settings (PromptPay ID, VAT, payment window, COD limit), users & staff roles, audit log.

Details: [docs/OVERVIEW.md](docs/OVERVIEW.md).

## Tech stack

| Layer | Technology | Notes |
|---|---|---|
| Backend | Go 1.26 (`net/http` router), [pgx v5](https://github.com/jackc/pgx), [goose v3](https://github.com/pressly/goose) | No web framework. Migrations are embedded and run on startup. |
| Database | PostgreSQL 17 | Runs in Docker (`docker-compose.yml`), host port **5433**. |
| Frontend | Vue 3 + TypeScript, Vite 7, Pinia, Vue Router, vue-i18n 11, zod | Built into static files that the Go server serves. |
| Testing | `go test`, Vitest + Vue Test Utils | See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md#6-testing-expectations). |
| Packaging | Multi-stage `Dockerfile` (Node 24 → Go 1.26 → Alpine 3.21) | One image: API + website. |

## Quick start

**Prerequisites:** Go **1.26+**, Node.js **22+** (24 recommended; `package.json` requires ≥ 20.19),
Docker Desktop (with WSL 2 on Windows), Git.

```bash
# 1. Configuration (Git Bash; in PowerShell use: Copy-Item .env.example .env)
cp .env.example .env

# 2. Database (PostgreSQL 17 on localhost:5433)
docker compose up -d db

# 3. API on http://localhost:8080 (applies migrations, creates the first admin)
cd backend
go run ./cmd/api

# 4. Demo data — in a second terminal
cd backend
go run ./cmd/seed

# 5. Website on http://localhost:5173 — in a third terminal
cd frontend
npm ci
npm run dev
```

Open <http://localhost:5173>. Full Windows walkthrough and troubleshooting: [docs/SETUP.md](docs/SETUP.md).

To try the production image instead (database + app on <http://localhost:8080>):

```bash
docker compose --profile app up --build
```

This needs a `.env` with a real `OTP_SECRET` (see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)).

## Development workflow

The usual loop is **db → api → seed → frontend**:

1. `docker compose up -d db` — start PostgreSQL once; it keeps running in the background.
2. `go run ./cmd/api` in `backend/` — restart it after Go changes. Migrations apply automatically.
3. `go run ./cmd/seed` in `backend/` — only needed once (safe to run again; it does not duplicate data).
4. `npm run dev` in `frontend/` — Vite hot-reloads Vue changes and proxies `/api` and `/uploads` to `localhost:8080`, so the browser only talks to one origin.

## Demo logins

There are no passwords. Every sign-in uses an emailed 6-digit code. In development **the code is
not emailed; it is printed in the API terminal**, for example:

```
level=INFO msg="OTP email (not sent, logged)" to=admin@example.com code=482913
```

| Who | Email | How you get the role |
|---|---|---|
| Admin | `admin@example.com` | The value of `ADMIN_EMAIL` in `.env` is made an admin every time the API (or the seed) starts. |
| Customer | any email, e.g. `customer@example.com` | A new email becomes a `customer` account on first sign-in. |
| Staff | any email | An admin adds it on **Admin → Users & staff**. |

Demo coupons from the seed: **WELCOME10** (10% off, min ฿300, max ฿200 discount, once per
customer) and **SAVE50** (฿50 off, min ฿500, up to 3 times per customer). The seed also sets a
demo PromptPay ID (`0812345678`, "Sawyun Shop"). **Never pay real money to the demo QR.**

## Project structure

```
business-web-app/
├── .env.example              # all configuration variables, copy to .env
├── docker-compose.yml        # db (always) + app (profile "app")
├── Dockerfile                # builds the single production image
├── backend/                  # Go module github.com/Khant26/business-web-app/backend
│   ├── cmd/
│   │   ├── api/              # main server: config → DB → migrate → admin → jobs → HTTP
│   │   └── seed/             # idempotent demo data
│   └── internal/
│       ├── server/           # routes, middleware order, static files, SPA fallback
│       ├── auth/             # OTP, sessions, role checks, mailer
│       ├── orders/           # cart validation, quote, checkout, payments, admin orders
│       ├── pricing/          # pure price calculation (unit tested)
│       ├── promptpay/        # EMVCo QR payload + CRC16
│       ├── catalog/  coupons/  inventory/  shipping/  settings/  users/  audit/
│       ├── db/               # pool, transactions, migrations/*.sql
│       ├── httpx/  middleware/  ratelimit/  jobs/  geo/  config/
└── frontend/                 # Vue 3 + Vite app
    ├── public/data/thai-address.json   # province/district/subdistrict tree
    ├── scripts/build-thai-address.mjs  # regenerates that file
    └── src/
        ├── api/              # fetch client + one module per area (admin/ separate)
        ├── pages/store/  pages/admin/  layouts/
        ├── components/ui/  components/store/  components/admin/
        ├── stores/           # Pinia: auth, cart, wishlist, settings, ui
        ├── i18n/locales/     # en/my × (main, account, admin)
        ├── composables/  utils/  types/  styles/
        └── router/
```

## Commands

| Where | Command | What it does |
|---|---|---|
| root | `docker compose up -d db` | Start PostgreSQL 17 on `localhost:5433` |
| root | `docker compose --profile app up --build` | Build and run the production image with the database |
| root | `docker compose down` | Stop containers (data volumes are kept) |
| `backend/` | `go run ./cmd/api` | Run the API (port 8080) |
| `backend/` | `go run ./cmd/seed` | Insert demo data |
| `backend/` | `go test ./...` | Unit tests (integration test skips without `TEST_DATABASE_URL`) |
| `backend/` | `go vet ./...` / `gofmt -l .` | Static checks / list unformatted files |
| `frontend/` | `npm ci` | Install exact dependency versions |
| `frontend/` | `npm run dev` | Dev server on port 5173 with API proxy |
| `frontend/` | `npm run build` | Type-check and build to `frontend/dist` |
| `frontend/` | `npm run preview` | Serve the build on port 4173 (also proxies the API) |
| `frontend/` | `npm run lint` / `npm run lint:fix` | ESLint |
| `frontend/` | `npm run typecheck` | `vue-tsc` type check |
| `frontend/` | `npm test` / `npm run test:watch` | Vitest |
| `frontend/` | `npm run format` / `npm run format:check` | Prettier |
| `frontend/` | `npm run build:address` | Rebuild `public/data/thai-address.json` |

## Documentation

| Document | Read it when you want to… |
|---|---|
| [docs/OVERVIEW.md](docs/OVERVIEW.md) | understand what the product does, roles, order lifecycle and pricing |
| [docs/SETUP.md](docs/SETUP.md) | set up a development PC, understand `.env`, run tests, fix common problems |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | learn how the code is organised and why |
| [docs/DATABASE.md](docs/DATABASE.md) | see every table, the ER diagram and how to add a migration |
| [docs/API.md](docs/API.md) | look up an endpoint (the API contract) |
| [docs/SECURITY.md](docs/SECURITY.md) | understand OTP, sessions, rate limits and the production checklist |
| [docs/USER_GUIDE.md](docs/USER_GUIDE.md) | help a customer use the shop |
| [docs/ADMIN_GUIDE.md](docs/ADMIN_GUIDE.md) | run the shop day to day as staff or admin |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | build, deploy, back up and upgrade |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | contribute: branches, code style, adding features, backlog |

## Verification status

On **2026-10-29** the following passed: Go unit tests and the database integration test; a
95-check API smoke test; frontend lint, type check, 38 Vitest tests and production build; a
headless-browser end-to-end run (customer OTP login → coupon → PromptPay order → slip upload →
admin approval) at 375 px (mobile) and 1366 px (desktop); Myanmar text rendering; and building and
running the production Docker image. The smoke and browser tests were run by hand and are not yet
committed to the repository (see the backlog in [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md#8-suggested-backlog)).
