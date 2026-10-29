# Development guide

How the internship team works on this codebase: Git workflow, code style, recipes for common
changes, testing expectations and a backlog of good next tasks. Set up your PC first with
[SETUP.md](SETUP.md) and read [ARCHITECTURE.md](ARCHITECTURE.md).

## 1. Git workflow

- `main` is always deployable. **Nobody pushes directly to `main`.**
- Create a branch from the latest `main` for each task:

  | Prefix | Use for | Example |
  |---|---|---|
  | `feature/` | new functionality | `feature/related-products` |
  | `fix/` | bug fixes | `fix/coupon-message-codes` |
  | `docs/` | documentation only | `docs/deployment-caddy` |
  | `refactor/` | code changes without behaviour change | `refactor/order-loader` |
  | `chore/` | dependencies, tooling, CI | `chore/github-actions` |

- Commit messages follow the style already in the history (Conventional Commits): `feat: add related products endpoint`, `fix: translate coupon_used_up`, `docs: …`, `test: …`, `chore: …`. One logical change per commit.
- Open a **pull request** into `main`. The **senior student lead reviews and merges** every PR. Peers are encouraged to review each other's PRs too.
- Keep PRs small (ideally under ~400 changed lines). Big features can be split: backend endpoint first, then UI.

### Pull request checklist

Copy into the PR description:

```markdown
## What and why
<one paragraph>

## How I tested
- [ ] backend: go vet ./... && go test ./...
- [ ] frontend: npm run lint && npm run typecheck && npm test && npm run build
- [ ] tried it in the browser (mobile 375 px and desktop width), English and Myanmar
- [ ] screenshots attached for UI changes

## Contract and data
- [ ] docs/API.md updated (if any endpoint or JSON shape changed)
- [ ] frontend/src/types/api.ts updated to match
- [ ] new migration file (never edited an old one)
- [ ] both en and my translations added
```

Reviewer focus: correctness of money and stock logic, authorization (right middleware), SQL
parameters, error codes, both languages, mobile layout.

## 2. Code style

### Frontend (TypeScript / Vue)

- **Prettier** (`frontend/.prettierrc.json`): no semicolons, single quotes, 100-character lines. Run `npm run format` before committing; CI-style check: `npm run format:check`.
- **EditorConfig**: UTF-8, 2 spaces, **LF** line endings, final newline. On Windows set your editor to LF (VS Code: `"files.eol": "\n"`), otherwise Prettier reports every line as changed.
- **ESLint** (`npm run lint`): Vue recommended + TypeScript recommended, plus project rules: `no-explicit-any` is an error, unused variables are errors (prefix with `_` to ignore an argument), `v-html` is forbidden (XSS protection).
- Use `<script setup lang="ts">` and the Composition API. Put API calls in `src/api/*`, never call `fetch` in components. Put shared state in Pinia stores, reusable logic in `composables/`.
- All visible text goes through i18n (`$t('…')` / `t('…')`) — no hard-coded English in templates.
- Money: keep satang integers; format only for display with `formatBaht()`; convert input with `bahtToSatang()`.
- Bilingual API fields: use `useLocalized()` / `pickLocalized()` instead of picking `_en` / `_my` by hand.

### Backend (Go)

- `gofmt` (most editors run it on save) — `gofmt -l .` must print nothing. `go vet ./...` must pass.
- Handlers have the signature `func(w http.ResponseWriter, r *http.Request) error` and return errors from `httpx` (`httpx.Validation`, `httpx.NotFound()`, `httpx.Conflict(code, msg)`, `httpx.E(status, code, msg)`). Do not write error JSON by hand.
- Validate input with `httpx.Fields{}` + `f.Check(cond, field, msg)` and return `f.Err()`.
- Always use SQL **parameters** (`$1`). For search use `db.LikePattern(s)`.
- Several writes that belong together → `db.Tx(ctx, h.DB, func(tx pgx.Tx) error { … })`. Lock rows you will update with `FOR UPDATE`.
- Every staff/admin write calls `audit.Log(ctx, tx, r, "entity.action", "entity", id, details)` **inside** the transaction.
- Stock only changes through `inventory.Move` (it writes the movement row).
- Money is `int64` satang; never `float64`.
- Exported functions get a short doc comment (the codebase follows this everywhere).

## 3. Recipe: adding an API endpoint end-to-end

Example: show **"You may also like"** on the product page — up to 4 popular products from the same
category. New endpoint: `GET /api/v1/products/{slug}/related`.

### Step 1 — Backend handler

Put it next to related code, here `backend/internal/catalog/public.go`:

```go
// Related handles GET /products/{slug}/related.
func (h *Handler) Related(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var id int64
	var catID *int64
	if err := h.DB.QueryRow(ctx, `SELECT id, category_id FROM products WHERE slug = $1 AND status = 'active'`,
		r.PathValue("slug")).Scan(&id, &catID); err != nil {
		return err // pgx.ErrNoRows becomes 404 not_found automatically
	}
	if catID == nil {
		return httpx.OK(w, []ProductCard{})
	}
	cards, err := QueryCards(ctx, h.DB, "p.status = 'active' AND p.category_id = $1 AND p.id <> $2",
		"ORDER BY p.sold_count DESC, p.id DESC LIMIT 4", *catID, id)
	if err != nil {
		return err
	}
	return httpx.OK(w, cards)
}
```

### Step 2 — Route

In `backend/internal/server/server.go`, in the right section, with the right middleware:

```go
handle("GET /products/{slug}/related", cat.Related)          // public: no middleware
// customer-only would be:   handle("…", h.X, authed)
// staff/admin:              handle("…", h.X, staff)
// admin only:               handle("…", h.X, admin)
// uploads add:              uploadRL   (after the auth middleware)
```

For a **create** endpoint return `httpx.Created(w, obj)` (201); for delete return `httpx.NoContent(w)` (204).

### Step 3 — Test the backend

Add a unit test if there is pure logic (put pure logic in a function without the database, like
`pricing`). Then try it:

```bash
curl http://localhost:8080/api/v1/products/classic-cotton-tshirt/related
# non-GET requests also need:  -H "X-Requested-With: sawyun" -H "Content-Type: application/json"
```

### Step 4 — Document the contract

Add a row to the right table in [API.md](API.md):
`| GET | /products/{slug}/related | ProductCard[] (max 4, same category, by sold_count) |`

### Step 5 — Frontend types and API module

If the response uses a new shape, add it to `frontend/src/types/api.ts`. Here `ProductCard` already
exists. Add the call to `frontend/src/api/catalog.ts`:

```ts
related: (slug: string) => api.get<ProductCard[]>(`/products/${seg(slug)}/related`),
```

(`seg()` URL-encodes the path segment. Admin calls go in `src/api/admin/index.ts`.)

### Step 6 — Use it in a page

In `frontend/src/pages/store/ProductDetailPage.vue` (sketch):

```ts
import { useAsync } from '@/composables/useAsync'
import { catalogApi } from '@/api/catalog'

const related = useAsync((slug: string) => catalogApi.related(slug))
watch(() => props.slug, (s) => related.run(s), { immediate: true })
```

```vue
<section v-if="related.data.value?.length">
  <h2>{{ $t('product.related') }}</h2>
  <ProductGrid :products="related.data.value" />
</section>
```

`useAsync` returns `{ data, error, loading, run, retry, set }` and drops stale responses, so calling
`run` on every slug change is safe. `ProductGrid` takes `products` (and optional `loading`).

### Step 7 — Translations

Add `product.related` to **both** `en.json` and `my.json` (next section). New backend error codes
need an `errors.<code>` entry in both files, otherwise the UI shows the English server message.

### Step 8 — Verify

`go test ./...`, `npm run lint`, `npm run typecheck`, `npm test`, then check the page on a phone-size
window in both languages.

## 4. Adding a translation key

Message files are in `frontend/src/i18n/locales/`:

| Files | Contains | Loaded |
|---|---|---|
| `en.json`, `my.json` | app, common, nav, product, catalog, cart, auth, status, validation, **errors** … | at start |
| `account.en.json`, `account.my.json` | checkout, address, account pages, orders, pay | at start |
| `admin.en.json`, `admin.my.json` | everything under `admin.*` | lazily, when `/admin` opens |

Rules:

1. Add the key to the **English and Myanmar** file of the same pair, at the same path. The test `src/i18n/__tests__/messages.spec.ts` fails if the key sets differ.
2. If you don't have the Myanmar translation yet, ask the team; do not copy English into `my.json` silently — mark it in the PR so someone translates it before merge.
3. vue-i18n gives special meaning to some characters. Escape them with a literal interpolation:

   | Character | Meaning in vue-i18n | Write instead |
   |---|---|---|
   | `@` | linked message (`@:other.key`) | `{'@'}` — e.g. `"you{'@'}example.com"` (used in `auth.emailPlaceholder`) |
   | `{` `}` | placeholder (`{n}`) | `{'{'}` and `{'}'}` |
   | `\|` (pipe) | separates plural forms | `{'\|'}` |

   The same test compiles every message and fails on invalid syntax, so run it after editing.
4. Plurals: `"items": "{n} item | {n} items"` and call `t('common.items', n)`. Three forms (`zero | one | many`) are also used, e.g. `"No products | {n} product | {n} products"`.
5. Placeholders use names: `"Only {n} left"` → `t('product.lowStock', { n: 3 })`.

Run only the i18n test:

```bash
cd frontend
npx vitest run src/i18n
```

## 5. Adding a database migration

Short version: add `backend/internal/db/migrations/00002_<what>.sql` with `-- +goose Up` and
`-- +goose Down` sections, restart the API, and **never edit an existing migration**. Full rules
and an example are in [DATABASE.md](DATABASE.md#4-migrations).

## 6. Testing expectations

| Change | Required tests |
|---|---|
| Pricing, coupons, shipping, VAT | Table-driven cases in `backend/internal/pricing/pricing_test.go`. Include an edge case (exact threshold, cap, zero). |
| Order status rules | `backend/internal/orders/status_test.go`. |
| New middleware / security behaviour | `backend/internal/middleware/middleware_test.go` style tests with `httptest`. |
| Order placement, stock, expiry, anything with transactions | Extend `internal/server/integration_test.go` and run it with `TEST_DATABASE_URL` (see [SETUP.md](SETUP.md#10-running-tests)). |
| Frontend utils, stores, API client | Vitest spec in a `__tests__` folder next to the code (`*.spec.ts`). |
| Components with logic | Vue Test Utils spec, like `components/__tests__/ProductCard.spec.ts`. |
| New translation keys | The i18n spec runs automatically. |
| UI changes | Manual check at 375 px and desktop width, English and Myanmar; screenshots in the PR. |

All of these must pass before merge:

```bash
cd backend  && go vet ./... && go test ./...
cd frontend && npm run lint && npm run typecheck && npm test && npm run build
```

Bug fixes should come with a test that fails before the fix.

## 7. Known issues to fix first

Found while writing these docs (2026-10-29). Coupon/zone error translations, the PromptPay ID hint and
SMTP sending in development were fixed the same day. Still open:

1. **`PUBLIC_URL` is not used yet.** It is reserved for links in future emails; either use it (backlog item 1) or remove it from `.env.example`.
2. **Smoke and browser tests are not in the repo.** The API smoke test and headless-browser flow were run by hand; committing them (e.g. Playwright in CI) is a good task.

## 8. Suggested backlog

Ordered roughly from easiest to hardest. Each is a good self-contained internship task.

| # | Task | Notes / hints |
|---|---|---|
| 1 | **Order emails** (placed, payment approved/rejected, shipped) in English + Myanmar | Reuse `auth.SMTPMailer.Send`. Send **after** the transaction commits, in a goroutine or a small `email_outbox` table processed by a job (retries). Use `PUBLIC_URL` for links. Respect the development rule (log instead of send). |
| 2 | **CI with GitHub Actions** | Jobs: Go (`go vet`, `gofmt -l`, `go test`, integration test with a `postgres:17` service and `TEST_DATABASE_URL`), frontend (`npm ci`, lint, typecheck, test, build), and `docker build`. Require green CI before merge. |
| 3 | **Commit the end-to-end tests** | Turn the manual browser flow (OTP login → coupon → PromptPay order → slip → admin approve) into Playwright tests in `frontend/e2e/` (already excluded from Vitest), run in CI at 375 px and 1366 px. Read the OTP from a test mailer or the log. Also commit the API smoke test as a script. |
| 4 | **Product search improvements** | Search `description_my` too; PostgreSQL full-text search or `pg_trgm` for typo tolerance; search suggestions in `SearchBox`. Add indexes in a migration. |
| 5 | **Image resizing / thumbnails** | On upload, create e.g. 400 px and 1200 px WebP versions (Go `image` packages or a small library) and serve the right size with `srcset`. Saves mobile data. Mind the 5 MB limit and memory use. |
| 6 | **Low-stock and new-order notifications for staff** | Email or LINE Notify / Telegram message to staff when a slip is uploaded or stock drops below threshold. |
| 7 | **Courier API integration** | Start with one courier (e.g. Thailand Post tracking API): fetch status by tracking number in a background job and auto-mark *Delivered*. Keep manual entry as fallback. |
| 8 | **Sales reports and CSV export** | Monthly sales, VAT totals for accounting, best sellers; CSV download for orders. |
| 9 | **Refund tracking** | Record refunds for cancelled paid orders (amount, date, reference) so accounting matches. |
| 10 | **Automatic payment verification** | Investigate slip-verification or bank APIs (cost, contract). Keep manual approval as fallback. |
| 11 | **Cloud object storage for uploads** | S3-compatible storage behind an interface so local disk still works in development; needed before running more than one app instance. |
| 12 | **Shared rate limiting** | Redis or PostgreSQL-based limiter if the app ever runs on more than one instance. |
| 13 | **Thai language** | Add `th` locale files and `name_th` columns, or a translation table if more languages are planned. |

Before starting a backlog item, write a short plan in the PR or an issue (endpoints, tables,
screens) and get the senior's OK — it saves rework.
