# Sawyun Ecommerce — HTTP API contract

This is the single source of truth shared by the Go backend and the Vue frontend.

## Conventions

- Base path: **`/api/v1`**. JSON in and out (`Content-Type: application/json`) except file uploads (`multipart/form-data`, field name **`file`**).
- **Money** is always integer **satang** (`*_satang`, 1 THB = 100 satang). Frontend formats as `฿1,290.00`.
- **Times** are RFC 3339 strings in UTC.
- **Bilingual fields** come as pairs: `name_en` / `name_my`, `title_en` / `title_my`, `description_en` / `description_my`. The frontend picks the current locale and falls back to English when the Myanmar text is empty.
- **Auth** uses an `httpOnly` cookie named **`sy_session`** (SameSite=Lax, Secure when `COOKIE_SECURE=true`). The frontend sends `credentials: 'include'`.
- **CSRF**: every non-GET request must include the header **`X-Requested-With: sawyun`**; otherwise `403 csrf`.
- **Errors**: `{"error": {"code": "string_code", "message": "Human readable English", "fields": {"field": "msg"}}}` (`fields` optional). Status codes: 400 bad input (`bad_request`, `otp_*`), 401 `unauthorized`, 403 `forbidden`/`csrf`, 404 `not_found`, 409 `conflict`, 413 `too_large`, 422 `validation`, 429 `rate_limited` (with `Retry-After` header), 500 `internal`.
- **Pagination**: `?page=1&page_size=20` (max 100) → `{"items": [...], "total": 123, "page": 1, "page_size": 20}`.
- **Uploaded public images** are served at `/uploads/...` (outside `/api`). Payment slips are private and only served through the admin endpoint.
- Frontend error-code → message translation lives in the frontend i18n files, keyed by `error.code`.

## Shared types (TypeScript notation)

```ts
type User = { id: number; email: string; name: string; phone: string; role: 'customer' | 'staff' | 'admin'; created_at: string }

type Category = { id: number; slug: string; name_en: string; name_my: string; image_url: string; sort_order: number; is_active: boolean; product_count: number }

type Variant = { id: number; sku: string; name_en: string; name_my: string; price_satang: number; compare_at_satang: number | null; stock: number; is_active: boolean; sort_order: number }
// admin responses additionally include low_stock_threshold

type ProductCard = {
  id: number; slug: string; name_en: string; name_my: string
  image_url: string                       // first image or ''
  price_satang: number                    // lowest active variant price
  compare_at_satang: number | null        // compare price of that variant
  rating_avg: number; rating_count: number; sold_count: number
  in_stock: boolean; featured: boolean
  category: { slug: string; name_en: string; name_my: string } | null
}

type ProductDetail = ProductCard & {
  description_en: string; description_my: string
  images: { id: number; url: string; sort_order: number }[]
  variants: Variant[]                     // active variants only (public)
}

type Event = { id: number; slug: string; title_en: string; title_my: string; description_en: string; description_my: string; banner_url: string; starts_at: string; ends_at: string; is_active: boolean; sort_order: number }

type ShippingZone = { id: number; name_en: string; name_my: string; fee_satang: number; free_over_satang: number | null; is_default: boolean; sort_order: number; provinces: string[] }

type Address = { id: number; label: string; recipient_name: string; phone: string; line1: string; subdistrict: string; district: string; province: string; postcode: string; is_default: boolean }
// province = English province name, e.g. "Bangkok", "Chiang Mai"

type OrderStatus = 'pending_payment' | 'payment_review' | 'paid' | 'confirmed' | 'packing' | 'shipped' | 'delivered' | 'cancelled' | 'expired'

type OrderItem = { id: number; product_id: number | null; product_slug: string; variant_id: number | null; product_name_en: string; product_name_my: string; variant_name_en: string; variant_name_my: string; sku: string; image_url: string; unit_price_satang: number; quantity: number; line_total_satang: number; can_review: boolean }

type OrderSummary = { code: string; status: OrderStatus; payment_method: 'promptpay' | 'cod'; total_satang: number; item_count: number; created_at: string; first_item_image: string }

type Order = OrderSummary & {
  subtotal_satang: number; discount_satang: number; shipping_satang: number; vat_rate_bp: number; vat_satang: number
  coupon_code: string
  shipping_address: { recipient_name: string; phone: string; line1: string; subdistrict: string; district: string; province: string; postcode: string }
  email: string; note: string
  courier: string; tracking_number: string; tracking_url: string
  payment_due_at: string | null; paid_at: string | null; shipped_at: string | null; delivered_at: string | null; cancelled_at: string | null; cancel_reason: string
  items: OrderItem[]
  history: { status: OrderStatus; note: string; created_at: string }[]
  slips: Slip[]
  can_cancel: boolean; can_upload_slip: boolean
}

type Slip = { id: number; status: 'pending' | 'approved' | 'rejected'; note: string; uploaded_at: string; reviewed_at: string | null }
// admin Slip additionally has: image_url: string  (→ GET /api/v1/admin/slips/{id}/file)

type Quote = {
  lines: { variant_id: number; product_slug: string; product_name_en: string; product_name_my: string; variant_name_en: string; variant_name_my: string; image_url: string; unit_price_satang: number; quantity: number; line_total_satang: number; available: number }[]
  subtotal_satang: number; discount_satang: number; shipping_satang: number; vat_rate_bp: number; vat_satang: number; total_satang: number
  coupon: { code: string; valid: boolean; message_code: string } | null    // message_code e.g. coupon_expired, coupon_min_order
  zone: { name_en: string; name_my: string; free_over_satang: number | null } | null
  problems: { variant_id: number; code: 'out_of_stock' | 'insufficient_stock' | 'unavailable'; available: number }[]
  cod_allowed: boolean
}

type Review = { id: number; rating: number; comment: string; user_name: string; created_at: string }
```

### Pricing rules (backend is authoritative)

1. `subtotal` = Σ unit price × qty (current variant prices).
2. `discount` from coupon: percent → `floor(subtotal × value / 100)` capped by `max_discount_satang`; fixed → `min(value, subtotal)`. Coupon must be active, inside `starts_at..ends_at`, `subtotal ≥ min_order_satang`, `used_count < usage_limit`, user's redemptions `< per_user_limit`.
3. `shipping` = zone of the province (or default zone): `fee_satang`, or 0 when `free_over_satang` is set and `subtotal − discount ≥ free_over_satang`.
4. `vat` = `round((subtotal − discount + shipping) × vat_rate_bp / 10000)` (half up).
5. `total` = `subtotal − discount + shipping + vat`.
6. COD allowed only when `total ≤ settings.cod_max_satang`.

### Order lifecycle

| Method | Created as | Flow |
|---|---|---|
| `promptpay` | `pending_payment`, `payment_due_at = now + payment_window_minutes` | slip upload → `payment_review` → admin approve → `paid` (or reject → back to `pending_payment`, due time extended by the window) → `packing` → `shipped` → `delivered` |
| `cod` | `confirmed` | → `packing` → `shipped` → `delivered` |

- Stock is decremented when the order is placed (`inventory_movements.reason = 'order'`), inside one transaction with `SELECT … FOR UPDATE`.
- `cancelled` / `expired` restore stock (`reason = 'cancel' | 'expire'`) and release the coupon redemption. A background job expires `pending_payment` orders past `payment_due_at` every minute.
- Customer can cancel while `pending_payment`, `payment_review` or `confirmed`.
- Admin allowed transitions: `payment_review→paid` (via slip approve), `paid|confirmed→packing`, `packing→shipped` (requires tracking number), `shipped→delivered`, any non-final (`not shipped/delivered/cancelled/expired`) → `cancelled`.
- `delivered` increments `products.sold_count`. Every change writes `order_status_history`.

---

## Public endpoints

| Method | Path | Response |
|---|---|---|
| GET | `/health` | `{status:"ok"}` |
| GET | `/settings/public` | `{shop_name, vat_rate_bp, payment_window_minutes, cod_max_satang, support_email, support_phone}` |
| GET | `/home` | `{events: Event[] (currently running, active), categories: Category[] (active), popular: ProductCard[] (top 8 by sold_count), featured: ProductCard[] (8), new_arrivals: ProductCard[] (8)}` |
| GET | `/categories` | `Category[]` (active, by sort_order) |
| GET | `/categories/{slug}` | `Category` |
| GET | `/products` | paged `ProductCard`. Query: `q`, `category` (slug), `event` (slug), `sort` = `popular`(default) \| `newest` \| `price_asc` \| `price_desc` \| `rating`, `min_price`, `max_price` (satang), `in_stock=1`, `page`, `page_size`. Only `status=active`. |
| GET | `/products/{slug}` | `ProductDetail` |
| GET | `/products/{slug}/reviews` | paged `Review` (approved only) |
| GET | `/events` | `Event[]` currently running |
| GET | `/events/{slug}` | `Event & {products: ProductCard[]}` |
| GET | `/provinces` | `{name_en: string, name_th: string}[]` (77 Thai provinces) |
| GET | `/shipping/zones` | `ShippingZone[]` |
| POST | `/cart/validate` | body `{items: {variant_id, quantity}[]}` → `{lines: Quote['lines'], problems: Quote['problems']}` (no auth; used to refresh the local cart) |

## Auth

| Method | Path | Body → Response |
|---|---|---|
| POST | `/auth/otp/request` | `{email}` → `204`. Always 204 for valid email format (no account enumeration). Creates a 6-digit code (5 min TTL), emails it. The code is emailed when `SMTP_HOST` is set; in development it is also written to the API log (see Implementation notes). |
| POST | `/auth/otp/verify` | `{email, code}` → `{user: User}` + sets `sy_session`. Creates the user (role customer) if new. Errors: `otp_invalid`, `otp_expired`, `otp_too_many_attempts` (after 5 wrong codes the code is burned), `account_disabled`. |
| POST | `/auth/logout` | → `204` |
| GET | `/auth/me` | → `{user: User}` or `401` |

Rate limits (429 `rate_limited`): OTP request — 1 / 60 s and 5 / hour per email, 20 / hour per IP. OTP verify — 10 / 15 min per IP. Orders — 10 / hour per user. Uploads — 20 / hour per user. Everything else — 300 / min per IP.

## Customer account (requires login)

| Method | Path | Body → Response |
|---|---|---|
| PATCH | `/me` | `{name, phone}` → `{user}` |
| GET | `/me/sessions` | `{id, ip, user_agent, created_at, last_seen_at, current: boolean}[]` (`id` is an opaque short id) |
| DELETE | `/me/sessions/{id}` | `204` |
| POST | `/me/sessions/revoke-others` | `204` (sign out all other devices) |
| GET | `/me/addresses` | `Address[]` |
| POST | `/me/addresses` | Address without id → `Address` (first address becomes default) |
| PUT | `/me/addresses/{id}` | → `Address` |
| DELETE | `/me/addresses/{id}` | `204` |
| GET | `/me/wishlist` | `ProductCard[]` |
| PUT | `/me/wishlist/{productId}` | `204` (add) |
| DELETE | `/me/wishlist/{productId}` | `204` |
| GET | `/me/wishlist/ids` | `number[]` |
| POST | `/checkout/quote` | `{items: {variant_id, quantity}[], province?, coupon_code?}` → `Quote` (no province → `zone: null`, shipping 0) |
| POST | `/orders` | `{items, address_id?, address?: Address-without-id, payment_method, coupon_code?, note?, save_address?: boolean}` → `Order` (201). Errors: `validation`, `out_of_stock` (with `fields`), `coupon_invalid`, `cod_not_allowed`, `empty_cart`. |
| GET | `/me/orders` | paged `OrderSummary`, query `status` optional |
| GET | `/me/orders/{code}` | `Order` |
| POST | `/me/orders/{code}/cancel` | `{reason?}` → `Order` |
| GET | `/me/orders/{code}/promptpay` | `{payload: string, amount_satang, promptpay_name, due_at}` — EMVCo PromptPay payload with amount; frontend renders the QR. 409 `not_payable` if not `pending_payment`. 409 `promptpay_not_configured` if no PromptPay ID set. |
| POST | `/me/orders/{code}/slip` | multipart `file` (jpg/png/webp, ≤ 5 MB) → `Order` (status becomes `payment_review`) |
| POST | `/products/{slug}/reviews` | `{rating 1-5, comment ≤ 1000}` → `201 {id}`. Requires a delivered order with this product; one review per product per user (`409 already_reviewed`, `403 not_purchased`). Review starts `pending`. |

## Admin (`/admin/*`, role staff or admin; ★ = admin only)

All admin writes are recorded in `audit_logs`.

| Method | Path | Body → Response |
|---|---|---|
| GET | `/admin/dashboard` | `{today_sales_satang, today_orders, pending_slips, to_pack, low_stock_count, sales_7d: {date: "YYYY-MM-DD", total_satang, orders}[], recent_orders: AdminOrderSummary[] (10), low_stock: InventoryRow[] (10)}` |
| GET | `/admin/orders` | paged `AdminOrderSummary` = `OrderSummary & {customer_email, recipient_name, has_pending_slip}`. Query `status`, `q` (code/email/name/phone), `payment_method`, `from`, `to` (dates) |
| GET | `/admin/orders/{code}` | `Order & {customer: User, slips: AdminSlip[]}` |
| POST | `/admin/orders/{code}/status` | `{status, note?}` → `Order` (409 `invalid_transition`) |
| PUT | `/admin/orders/{code}/tracking` | `{courier, tracking_number, tracking_url?}` → `Order` |
| POST | `/admin/slips/{id}/approve` | `{note?}` → `Order` |
| POST | `/admin/slips/{id}/reject` | `{note}` → `Order` |
| GET | `/admin/slips/{id}/file` | image bytes |
| GET | `/admin/products` | paged `AdminProductRow` = `{id, slug, name_en, name_my, status, featured, category_name_en, image_url, variant_count, total_stock, min_price_satang, updated_at}`. Query `q`, `status`, `category` (id) |
| POST | `/admin/products` | `ProductInput` → `AdminProduct` |
| GET | `/admin/products/{id}` | `AdminProduct` = `ProductDetail` (all variants incl. inactive, with `low_stock_threshold`) `& {status, category_id, event_ids: number[]}` |
| PUT | `/admin/products/{id}` | `ProductInput` → `AdminProduct` |
| DELETE | `/admin/products/{id}` | `204` (sets `status=archived`) |
| POST | `/admin/products/{id}/images` | multipart `file` → `{id, url, sort_order}` |
| PUT | `/admin/products/{id}/images/order` | `{image_ids: number[]}` → `204` |
| DELETE | `/admin/products/{id}/images/{imageId}` | `204` |
| POST | `/admin/products/{id}/variants` | `VariantInput` → `Variant` |
| PUT | `/admin/variants/{id}` | `VariantInput` (without stock) → `Variant` |
| DELETE | `/admin/variants/{id}` | `204` (deactivates if it has order history, otherwise deletes) |
| GET | `/admin/inventory` | paged `InventoryRow` = `{variant_id, sku, product_id, product_name_en, variant_name_en, stock, low_stock_threshold, is_low, image_url}`. Query `q`, `low=1` |
| POST | `/admin/inventory/{variantId}/adjust` | `{change: int (≠0), reason: 'restock'\|'adjust'\|'damage', note?}` → `InventoryRow` (stock can't go below 0 → 422) |
| GET | `/admin/inventory/movements` | paged `{id, variant_id, sku, product_name_en, change, reason, note, order_code, user_email, stock_after, created_at}`. Query `variant_id` |
| GET | `/admin/categories` | `Category[]` (all) |
| POST / PUT | `/admin/categories`, `/admin/categories/{id}` | `{slug, name_en, name_my, sort_order, is_active}` → `Category` |
| DELETE | `/admin/categories/{id}` | `204` (409 `in_use` if products reference it) |
| POST | `/admin/categories/{id}/image` | multipart `file` → `Category` |
| GET | `/admin/events` | `(Event & {product_ids: number[]})[]` (all) |
| POST / PUT | `/admin/events`, `/admin/events/{id}` | `{slug, title_en, title_my, description_en, description_my, starts_at, ends_at, is_active, sort_order, product_ids: number[]}` → event |
| DELETE | `/admin/events/{id}` | `204` |
| POST | `/admin/events/{id}/banner` | multipart `file` → event |
| GET | `/admin/coupons` | `Coupon[]` = `{id, code, type, value, min_order_satang, max_discount_satang, starts_at, ends_at, usage_limit, per_user_limit, used_count, is_active, created_at}` |
| POST / PUT | `/admin/coupons`, `/admin/coupons/{id}` | coupon fields (code is upper-cased) → `Coupon` |
| DELETE | `/admin/coupons/{id}` | `204` (deactivates if used, otherwise deletes) |
| GET | `/admin/reviews` | paged `{id, product_id, product_name_en, user_email, rating, comment, status, created_at}`. Query `status` |
| POST | `/admin/reviews/{id}/approve` \| `/reject` | `204` (recomputes product rating) |
| GET ★ | `/admin/shipping-zones` | `ShippingZone[]` |
| POST / PUT ★ | `/admin/shipping-zones`, `/admin/shipping-zones/{id}` | `{name_en, name_my, fee_satang, free_over_satang, is_default, sort_order, provinces: string[]}` → `ShippingZone` (a province can belong to one zone only → 409 `province_taken`) |
| DELETE ★ | `/admin/shipping-zones/{id}` | `204` (can't delete the default zone) |
| GET ★ | `/admin/settings` | `{shop_name, promptpay_id, promptpay_name, vat_rate_bp, payment_window_minutes, cod_max_satang, support_email, support_phone}` |
| PUT ★ | `/admin/settings` | same object → same |
| GET ★ | `/admin/users` | paged `User & {is_active, order_count}`. Query `role`, `q` |
| POST ★ | `/admin/users` | `{email, name, role: 'staff'\|'admin'}` → `User` (creates or promotes; they log in with OTP) |
| PATCH ★ | `/admin/users/{id}` | `{role?, is_active?}` → `User & {is_active, order_count}` (cannot demote/disable yourself → 403; disabling revokes sessions) |
| GET ★ | `/admin/audit-logs` | paged `{id, user_email, action, entity, entity_id, details, ip, created_at}` |

```ts
type ProductInput = { slug: string; category_id: number | null; name_en: string; name_my: string; description_en: string; description_my: string; status: 'draft' | 'active' | 'archived'; featured: boolean; event_ids?: number[] }
type VariantInput = { sku: string; name_en: string; name_my: string; price_satang: number; compare_at_satang: number | null; stock?: number /* create only; writes a 'restock' movement */; low_stock_threshold: number; is_active: boolean; sort_order: number }
```

---

## Implementation notes

Reviewed against `backend/internal/server/server.go` and the handlers on 2026-10-29. Where this
section is more specific than the tables above, **this section describes what the backend actually
does**.

### General

- Unknown `/api/*` paths return JSON `404 not_found`. `GET /health` only proves the process is up; it does not query the database.
- JSON bodies are limited to 1 MB (`413 too_large`). Invalid JSON or a missing body → `400 bad_request`.
- `page_size` defaults to 20 (max 100). Rows not found by id/code/slug → `404 not_found`; a customer asking for another customer's order also gets `404`.
- Generic `409 conflict` is used for duplicates, with `fields` naming the field: product/category/event `slug`, variant `sku`, coupon `code` (e.g. `{"code": "already exists"}`). It is also returned when a user already has 20 addresses. An unknown `category_id` on a product → `422 validation`.
- **All create endpoints return `201 Created`**: `POST /orders`, `POST /me/addresses`, `POST /products/{slug}/reviews`, and admin `POST /admin/products`, `/admin/products/{id}/variants`, `/admin/products/{id}/images`, `/admin/categories`, `/admin/events`, `/admin/coupons`, `/admin/shipping-zones`, `/admin/users`. The image endpoints for categories (`/image`) and events (`/banner`) return `200` with the updated object. `PUT`/`PATCH` updates return `200`; deletes and review moderation return `204`.

### Error codes chosen by the backend

| Code | Status | When |
|---|---|---|
| `otp_invalid` | 400 | Wrong code, no code requested, code already used, or code not 6 characters. |
| `otp_expired` | 400 | Code older than 5 minutes. |
| `otp_too_many_attempts` | 400 | 5 wrong attempts (the code is burned). |
| `account_disabled` | 403 | Correct code but `users.is_active = false`. |
| `csrf` | 403 | Non-GET `/api/*` request without `X-Requested-With: sawyun`. |
| `forbidden` | 403 | Wrong role; also when an admin tries to change their own role or disable themselves. |
| `not_purchased` | 403 | Review without a delivered order containing the product. |
| `invalid_transition` | 409 | Customer cancel not allowed in the current status; admin status change not allowed; slip approve/reject when the slip is not `pending` or the order is not `payment_review`; tracking update while the order is `pending_payment`, `payment_review`, `cancelled` or `expired`. |
| `not_payable` | 409 | `GET …/promptpay` or `POST …/slip` when the order is not a PromptPay order in `pending_payment`. |
| `promptpay_not_configured` | 409 | Settings have no PromptPay ID, **or** the stored ID is invalid. |
| `out_of_stock` | 409 | `POST /orders` when any line has a problem. `fields` is keyed by **variant id** (as a string) with the problem code as value, e.g. `{"12": "insufficient_stock", "15": "unavailable"}`. |
| `already_reviewed` | 409 | Second review of the same product by the same user. |
| `in_use` | 409 | Deleting a category that still has products (any status). |
| `province_taken` | 409 | Saving a shipping zone with a province that belongs to another zone (`fields.provinces` lists them). |
| `default_zone` | 409 | Deleting the default shipping zone. |
| `empty_cart` | 422 | `items` empty on `POST /checkout/quote` or `POST /orders`. (`POST /cart/validate` with no items returns `200` with empty arrays.) |
| `coupon_invalid` | 422 | `POST /orders` with a coupon that cannot be used; `fields.coupon_code` holds the coupon message code. |
| `cod_not_allowed` | 422 | `payment_method: "cod"` and total > `cod_max_satang`. |
| `validation` | 422 | Field errors. Notable cases: quantity outside 1–999 or more than 50 distinct variants (`fields.items`); unknown province (`fields.province` or `fields["address.province"]`); unknown `address_id`; shipping without a tracking number (`fields.tracking_number`); slip reject without `note`; inventory change that would make stock negative (`fields.change`); admin order list `from`/`to` not `YYYY-MM-DD`. |
| `rate_limited` | 429 | See rate limits; `Retry-After` is in whole seconds (minimum 1). |
| `internal` | 500 | Unexpected error (logged), or the OTP email could not be sent. |

### Auth

- `POST /auth/otp/request` returns `204` for any well-formed email (no account enumeration), but may return `422 validation` (`fields.email`), `429` or `500` (mail failure). Rate-limit order: per IP, then per email per minute, then per email per hour.
- The OTP is emailed whenever `SMTP_HOST` is set. With `SMTP_HOST` empty it is only written to the log. In development the code is always logged too, and an SMTP failure is logged but does not fail the request; in production an SMTP failure returns `500`.
- `POST /auth/otp/verify` checks only the **newest** code for the email. The verify limit (10 / 15 min per IP) counts every attempt, successful or not.
- Sessions: the cookie `Max-Age` equals `SESSION_TTL_HOURS`; a session is also invalid after `SESSION_IDLE_HOURS` without requests. Signing in beyond `SESSION_MAX_PER_USER` deletes the oldest sessions. `DELETE /me/sessions/{id}` returns `404` for an unknown id and clears the cookie when it deletes the current session.

### Quote, cart and orders

- `province` is optional in `POST /checkout/quote`. Empty → no zone: `zone: null` and `shipping_satang: 0`. English names are matched case-insensitively and Thai names are accepted; both become the canonical English name. A known province that is in no zone uses the default zone; if there is no default zone, `zone` is `null` and shipping is 0.
- Duplicate variants in `items` are merged (quantities added, capped at 999).
- Lines with problems stay in `lines` but are **excluded** from the subtotal; they are listed in `problems` (`out_of_stock`, `insufficient_stock` with `available`, or `unavailable` for inactive variants, non-active products and unknown ids).
- Coupon `message_code` values: `coupon_not_found`, `coupon_inactive`, `coupon_not_started`, `coupon_expired`, `coupon_used_up` (total `usage_limit` reached), `coupon_user_limit` (`per_user_limit` reached), `coupon_min_order`. Checks run in that order (minimum order last). `ends_at` is exclusive. `usage_limit: null` = unlimited; `per_user_limit: 0` = unlimited (default 1). Codes are trimmed and upper-cased.
- `item_count` in order summaries is the **total quantity** of all lines, not the number of lines.
- Order codes look like `SY-7K2M9Q` (6 characters from `23456789ABCDEFGHJKMNPQRSTUVWXYZ`).
- `POST /orders` with `address` (instead of `address_id`) reports field errors with the prefix `address.` (e.g. `address.postcode`). `save_address: true` saves it to the address book in the same transaction.
- `can_review` on an order item is true only when the order is `delivered`, the product still exists, and the user has not reviewed it yet.

### Payments

- Slips can be uploaded **only while the order is `pending_payment`** (PromptPay). After an upload the order is `payment_review`, so a second upload returns `409 not_payable` until staff reject the first one.
- Rejecting a slip moves the order back to `pending_payment` and sets `payment_due_at = now + payment_window_minutes`. Approving sets `paid` and `paid_at`. Both write an audit entry (`slip.approve` / `slip.reject`).
- Customers never receive slip image URLs (`image_url` is present only in admin responses).
- The PromptPay payload is a dynamic EMVCo QR (tag `01` = `12`) with tag `54` = amount in baht with 2 decimals and a CRC-16/CCITT-FALSE checksum. Accepted PromptPay IDs: 10-digit mobile starting with `0`, 11 digits starting with `66`, 13-digit tax ID, 15-digit e-wallet ID (non-digits are ignored).

### Admin

- `POST /admin/orders/{code}/status` returns `422 validation` for an unknown status before checking the transition. `delivered` increases `products.sold_count`. `cancelled` uses `note` as the reason (default "Cancelled by shop"), restores stock, releases the coupon and rejects pending slips.
- `PUT /admin/orders/{code}/tracking`: `courier` and `tracking_number` are required (≤ 100 characters); `tracking_url` is optional but must be `http(s)` with a host.
- `GET /admin/orders` `from` / `to` are Bangkok calendar dates (inclusive).
- `GET /admin/inventory` excludes archived products; `low=1` (or `true`) shows only active low-stock variants. Adjust `change` must be non-zero and within ±1,000,000.
- `POST /admin/users` returns the plain `User` (201); `PATCH /admin/users/{id}` returns `User & {is_active, order_count}`.
- `POST /admin/reviews/{id}/approve|reject` may be called again to change a decision; the product rating is recomputed each time.
- `PUT /admin/products/{id}/images/order` silently ignores image ids that do not belong to the product.
- `DELETE /admin/products/{id}` archives the product and also clears `featured`.
- Settings validation: `vat_rate_bp` 0–10000, `payment_window_minutes` 5–10080, `cod_max_satang` ≥ 0, `shop_name` required (≤ 100), `support_email` must be a valid email if set.

### Public catalog

- `GET /products?q=` searches `name_en`, `name_my` and `description_en`.
- `GET /events/{slug}` returns any **active** event, including scheduled and ended ones (only `/events` and `/home` filter to running events).
- `GET /products/{slug}/reviews` requires an `active` product; `POST /products/{slug}/reviews` accepts any non-archived product.
- `PUT /me/wishlist/{productId}` returns `404` if the product is not `active`; `GET /me/wishlist` lists only active products.
- `GET /provinces` is cached by browsers for 1 day (`Cache-Control: public, max-age=86400`).

