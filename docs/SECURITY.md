# Security

This document describes how Sawyun Ecommerce protects accounts, payments and data, and what must be
done before going live. Code references point to `backend/internal/…`.

## 1. Sign-in with email OTP (`auth/auth.go`, `auth/otp.go`)

There are no passwords. A customer proves they own an email address by typing a code sent to it.

| Property | Value | Why |
|---|---|---|
| Code | 6 random digits from `crypto/rand` (uniform, `000000`–`999999`). | Unpredictable. |
| Storage | Only `HMAC-SHA256(OTP_SECRET, email + "|" + code)` is stored in `otp_codes.code_hash`. | A database leak does not reveal usable codes; the email is part of the hash, so a code for one email cannot be used for another. |
| Comparison | Constant-time (`hmac.Equal`). | No timing leaks. |
| Lifetime | **5 minutes** (`OTPTTL`). | Short window for guessing or interception. |
| Attempts | **5 wrong tries** per code (`OTPMaxAttempts`); then the code is burned (`consumed_at` set) and `otp_too_many_attempts` is returned. | With 10⁶ possible codes, 5 guesses give a 0.0005% chance. |
| Which code | Only the **newest** code for the email is checked. Requesting a new code makes older ones useless. | No pile-up of valid codes. |
| Single use | A successful login sets `consumed_at`. | Codes cannot be replayed. |
| Request limits | 1 per **60 s** per email, **5 per hour** per email, **20 per hour** per IP. | Stops email bombing and code-spraying. |
| Verify limit | **10 attempts per 15 minutes** per IP (all attempts, right or wrong). | Slows brute force across many emails. |
| Account enumeration | `POST /auth/otp/request` answers `204` for any valid email, whether or not an account exists. | Attackers cannot learn who is a customer. |
| Disabled accounts | Checked after the code is verified; returns `403 account_disabled`. | |
| Clean-up | Codes older than 1 day are deleted every 30 minutes. | Less data kept. |

The email is plain text and bilingual (English and Myanmar). It is sent by SMTP whenever `SMTP_HOST` is set;
with `SMTP_HOST` empty it is only written to the log (`LogMailer`). In development the code is
always logged as well (`DevMailer`), and an SMTP failure is logged instead of blocking sign-in.
SMTP uses STARTTLS on port 587 (required unless the host is `localhost`) or implicit TLS on 465,
with TLS 1.2 minimum.

## 2. Sessions (`auth/auth.go`, `auth/sessions.go`)

| Property | Value |
|---|---|
| Token | 32 random bytes, base64url (43 characters), sent only in the cookie. |
| Stored as | `sessions.id = SHA-256(token)`. Someone who reads the database cannot use the sessions. |
| Cookie | Name `sy_session`, `HttpOnly` (JavaScript cannot read it), `SameSite=Lax`, `Path=/`, `Secure` when `COOKIE_SECURE=true`, `Max-Age` = absolute lifetime. |
| Absolute expiry | `SESSION_TTL_HOURS` (default 720 h = 30 days) from sign-in. |
| Idle expiry | `SESSION_IDLE_HOURS` (default 168 h = 7 days) since last use. `last_seen_at` is refreshed at most once a minute. |
| Max per user | `SESSION_MAX_PER_USER` (default 5). A new sign-in deletes the oldest extra sessions. |
| Invalid cookie | If the session is unknown, expired, idle or the user is disabled, the cookie is cleared and the request continues as anonymous. |
| Revoke | `POST /auth/logout` deletes the current session. The Security page can delete any one session (`DELETE /me/sessions/{id}`, using a 16-character public prefix of the hash, never the token) or all others (`POST /me/sessions/revoke-others`). |
| Admin action | Disabling a user deletes **all** their sessions immediately. |
| Clean-up | Expired and idle sessions are deleted every 30 minutes. |

**Why not JWT?** A database session can be revoked instantly and listed for the user. The cost is
one primary-key lookup per request, which is negligible at this scale.

## 3. Rate limits

All limits are token buckets kept **in memory** (`ratelimit` package). A blocked request gets
`429 rate_limited` with a `Retry-After` header (seconds); the frontend shows "Try again in N
seconds".

| Scope | Key | Limit | Where |
|---|---|---|---|
| All `/api/*` requests | client IP | 300 / minute | global middleware |
| OTP request | IP | 20 / hour | `auth.RequestOTP` |
| OTP request | email | 1 / 60 s and 5 / hour | `auth.RequestOTP` |
| OTP verify | IP | 10 / 15 minutes | `auth.VerifyOTP` |
| Place order (`POST /orders`) | user (IP if anonymous) | 10 / hour | route middleware |
| Uploads (slips, product images, category images, event banners) | user | 20 / hour | route middleware |

Limits reset when the process restarts and are **not shared** between several API instances. The
client IP comes from the TCP connection, or from the first `X-Forwarded-For` entry when
`TRUST_PROXY=true` — so `TRUST_PROXY` must only be enabled behind a proxy that **overwrites** that
header, otherwise anyone can fake their IP and escape the limits.

> Admin tip: the upload limit counts product images too. An admin uploading more than 20 images in
> an hour will be asked to wait.

## 4. CSRF protection

Every `/api/*` request that is not `GET`, `HEAD` or `OPTIONS` must carry the header
`X-Requested-With: sawyun`, otherwise it gets `403 csrf`.

Why this works: a malicious site can make the browser *send* a form POST with our cookie, but it
cannot add a custom header without a CORS pre-flight, and our server never approves cross-origin
pre-flights in production (CORS is only enabled in development, for one configured origin). The
`SameSite=Lax` cookie adds a second layer: the cookie is not sent on cross-site POSTs at all.
GET endpoints never change data.

## 5. CORS

- **Production:** no CORS headers at all. The site and API share one origin.
- **Development:** if `CORS_ORIGIN` is set, that single origin may make credentialed requests (methods GET/POST/PUT/PATCH/DELETE/OPTIONS; headers `Content-Type`, `X-Requested-With`; pre-flight cached 10 minutes). With the Vite proxy you do not need it.

## 6. Security headers (`middleware.SecurityHeaders`)

Sent on every response:

| Header | Value |
|---|---|
| `Content-Security-Policy` | `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; img-src 'self' data: blob:; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'` |
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` (no clickjacking) |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` — **only when `COOKIE_SECURE=true`** |

The CSP allows no inline scripts and no third-party scripts. `img-src data: blob:` is needed for the
QR code image (a data URL) and image previews before upload. ESLint forbids `v-html`
(`vue/no-v-html: error`), so user text is always escaped by Vue.

JSON responses also send `Cache-Control: no-store`.

## 7. Input validation and SQL

- JSON bodies are limited to **1 MB**; every handler validates fields and returns `422 validation` with per-field messages (`httpx.Fields`).
- All SQL uses **parameters** (`$1`, `$2`…). Search terms are escaped for `ILIKE` with `db.LikePattern` so `%` and `_` are treated literally.
- Prices, discounts and totals are **never** taken from the browser; the server recalculates them.
- Quantities must be 1–999 per line, at most 50 different variants per order.
- Addresses: 5-digit postcode, phone pattern, province must be one of the 77 known provinces.
- Tracking URLs must be `http(s)` with a host. Slugs must be lowercase `a-z0-9-`.

## 8. File uploads (`httpx/upload.go`)

| Check | Detail |
|---|---|
| Size | Max **5 MB** per file (request body capped slightly above). Bigger → `413 too_large`. |
| Type | Detected from the **file bytes** (`http.DetectContentType` + WebP signature). Only JPEG, PNG and WebP are accepted; the client's file name and MIME type are ignored. |
| Name | Random 32-hex-character name + extension from the detected type. Users cannot choose paths. |
| Location | Public images under `UPLOAD_DIR/public/…`, served at `/uploads/…`. Payment slips under `UPLOAD_DIR/private/slips/`, which is **never** served statically. |
| Slip access | Only staff/admin via `GET /api/v1/admin/slips/{id}/file`, with `Cache-Control: private, no-store`. Customers cannot see slip images through the API, even their own. |
| Path traversal | Static handlers clean paths; stored relative paths containing `..` are refused. |
| Clean-up | If saving the database row fails, the new file is deleted. Replaced/deleted images delete the old file. |

## 9. Authorization (role checks)

- `RequireAuth` → `401 unauthorized` for anonymous users.
- `RequireRole("staff","admin")` on all `/admin/*` routes, and `RequireRole("admin")` on shipping zones, settings, users and audit log → `403 forbidden` for the wrong role.
- Customer data is always filtered by the signed-in user (`… WHERE code = $1 AND user_id = $2`), so guessing another order code returns `404`.
- Admins cannot demote or disable themselves; `ADMIN_EMAIL` is re-promoted on every start.
- The frontend router also hides pages (`requiresAuth`, `requiresStaff`, `requiresAdmin`), but this is only for convenience — **the backend checks are what protect data**.

## 10. Audit log

Every staff/admin write calls `audit.Log` **inside the same database transaction** as the change, so
a change without an audit row cannot happen. Each row stores the user, action (for example
`order.status`, `order.tracking`, `slip.approve`, `slip.reject`, `product.create`,
`variant.update`, `inventory.adjust`, `coupon.create`, `coupon.deactivate`,
`shipping_zone.update`, `settings.update`, `user.grant_role`, `user.update`, `review.approve`),
the entity and id, a JSON `details` object (e.g. `{"from":"packing","to":"shipped"}`) and the IP.
Admins read it on **Admin → Audit log**. There is no UI to edit or delete audit rows.

## 11. Logging and privacy

- One log line per request (method, path, status, size, duration, IP, user id). Request bodies, cookies and tokens are **not** logged.
- In development the OTP code is logged on purpose. In production with SMTP configured it is not.
- Panics are logged with a stack trace and return a generic `500` to the client.

## 12. Production checklist

Before the shop goes live, the senior and supervisor should tick every line:

- [ ] `APP_ENV=production`.
- [ ] `OTP_SECRET` is a new random value of 32+ characters (e.g. `openssl rand -hex 32`), stored only on the server. (The API refuses to start with the default or anything shorter than 16 characters.)
- [ ] The site is served over **HTTPS** by a reverse proxy, and `COOKIE_SECURE=true` (this also turns on HSTS).
- [ ] `TRUST_PROXY=true` **only** if the proxy overwrites `X-Forwarded-For` with the real client IP; otherwise leave `false`.
- [ ] SMTP configured and tested with a real inbox; SPF/DKIM set up for the sender domain so codes do not land in spam.
- [ ] `ADMIN_EMAIL` is a real mailbox the owner controls.
- [ ] The PostgreSQL password is changed from `sawyun`, and port 5433 is **not** reachable from the internet (bind to `127.0.0.1` or remove the port mapping).
- [ ] Only ports 80/443 are open on the server firewall; port 8080 is reachable only from the proxy.
- [ ] Daily database backups (`pg_dump`) **and** uploads-volume backups, stored off the server, with a restore tested at least once (see [DEPLOYMENT.md](DEPLOYMENT.md#7-backups-and-restore)).
- [ ] The real PromptPay ID and account name are entered in **Settings**, and a 1-baht test payment was scanned successfully with a banking app.
- [ ] Demo data (seed products, coupons `WELCOME10`/`SAVE50`, demo PromptPay ID) removed or replaced — do **not** run the seed on production.
- [ ] Staff accounts reviewed on **Users & staff**; ex-staff disabled.
- [ ] Server OS and Docker images updated regularly; rebuild the app image to pick up Go/Alpine security fixes.
- [ ] Everyone knows the incident steps: rotate leaked secrets immediately (`OTP_SECRET` only affects sign-in codes; also change the SMTP password). Session tokens do not depend on `OTP_SECRET`, so to sign everybody out run `DELETE FROM sessions;` in `psql`; to sign out one user, disable them on **Users & staff**.
