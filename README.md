# Northstar Commerce

A production-shaped ecommerce web app with a Vue 3 + TypeScript frontend, TanStack Query and Table, and a dependency-free Go API. The demo works even when the API is offline (catalog fallback and browser-persisted cart); when the Go service is running, checkout orders are validated and persisted to `backend/data/orders.json`.

## Run locally

Prerequisites: Node.js 22+ and Go 1.23+.

```powershell
cd frontend
npm install
npm run dev
```

In a second terminal:

```powershell
cd backend
go run .
```

Open `http://localhost:5173`. Vite proxies `/api` to Go on port 8080. For a production-like run, build the frontend and let Go serve it:

```powershell
cd frontend
npm run build
cd ..\backend
go run .
```

Then open `http://localhost:8080`. Docker users can instead run `docker build -t northstar .` then `docker run -p 8080:8080 northstar`.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | API/web port |
| `WEB_DIR` | `../frontend/dist` | Built Vue files |
| `DATA_FILE` | `data/orders.json` | Demo order persistence |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed development origin |
| `ADMIN_KEY` | `northstar-demo-admin` | Protects `GET /api/orders`; change outside demos |

## Quality checks

```powershell
cd frontend
npm run build
```

The frontend build performs strict Vue/TypeScript checking before producing the production bundle. The Go API intentionally uses only the standard library, so `go test ./...` also compiles it without downloading dependencies.

## Documentation

- [Functional specification](docs/FUNCTIONAL_SPEC.md)
- [Database model and relationships](docs/DATABASE.md)
- [UX and UI system](docs/UX_UI.md)
- [API contract](docs/API.md)
- [Production checklist](docs/PRODUCTION.md)

## Demo boundaries

Checkout is a safe demo: it creates an order but does not charge a card. Authentication and a payment provider are documented production integrations rather than fake security. The inventory button in the header demonstrates a real TanStack Table-powered admin view.
