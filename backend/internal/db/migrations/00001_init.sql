-- +goose Up
-- Sawyun Ecommerce initial schema. All money values are integer satang (1 THB = 100 satang).

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    email       TEXT NOT NULL UNIQUE CHECK (email = lower(email)),
    name        TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    role        TEXT NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'staff', 'admin')),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE otp_codes (
    id          BIGSERIAL PRIMARY KEY,
    email       TEXT NOT NULL,
    code_hash   TEXT NOT NULL,
    attempts    INT NOT NULL DEFAULT 0,
    ip          TEXT NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_otp_email_created ON otp_codes (email, created_at DESC);

CREATE TABLE sessions (
    id           TEXT PRIMARY KEY,               -- sha256(token), never the raw token
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip           TEXT NOT NULL DEFAULT '',
    user_agent   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_sessions_user ON sessions (user_id, created_at);

CREATE TABLE addresses (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label          TEXT NOT NULL DEFAULT '',
    recipient_name TEXT NOT NULL,
    phone          TEXT NOT NULL,
    line1          TEXT NOT NULL,
    subdistrict    TEXT NOT NULL,
    district       TEXT NOT NULL,
    province       TEXT NOT NULL,                -- English province name (canonical key for shipping zones)
    postcode       TEXT NOT NULL,
    is_default     BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_addresses_user ON addresses (user_id);

CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name_en     TEXT NOT NULL,
    name_my     TEXT NOT NULL,
    image_url   TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id             BIGSERIAL PRIMARY KEY,
    slug           TEXT NOT NULL UNIQUE,
    category_id    BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    name_en        TEXT NOT NULL,
    name_my        TEXT NOT NULL,
    description_en TEXT NOT NULL DEFAULT '',
    description_my TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
    featured       BOOLEAN NOT NULL DEFAULT false,
    sold_count     INT NOT NULL DEFAULT 0,
    rating_avg     NUMERIC(3, 2) NOT NULL DEFAULT 0,
    rating_count   INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_category_status ON products (category_id, status);

CREATE TABLE product_images (
    id          BIGSERIAL PRIMARY KEY,
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_product_images_product ON product_images (product_id, sort_order);

CREATE TABLE product_variants (
    id                  BIGSERIAL PRIMARY KEY,
    product_id          BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku                 TEXT NOT NULL UNIQUE,
    name_en             TEXT NOT NULL,          -- e.g. "Red / M"
    name_my             TEXT NOT NULL,
    price_satang        INT NOT NULL CHECK (price_satang >= 0),
    compare_at_satang   INT CHECK (compare_at_satang IS NULL OR compare_at_satang >= 0),
    stock               INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    low_stock_threshold INT NOT NULL DEFAULT 5,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    sort_order          INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_variants_product ON product_variants (product_id, sort_order);

CREATE TABLE events (
    id             BIGSERIAL PRIMARY KEY,
    slug           TEXT NOT NULL UNIQUE,
    title_en       TEXT NOT NULL,
    title_my       TEXT NOT NULL,
    description_en TEXT NOT NULL DEFAULT '',
    description_my TEXT NOT NULL DEFAULT '',
    banner_url     TEXT NOT NULL DEFAULT '',
    starts_at      TIMESTAMPTZ NOT NULL,
    ends_at        TIMESTAMPTZ NOT NULL,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    sort_order     INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE event_products (
    event_id    BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sort_order  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (event_id, product_id)
);

CREATE TABLE shipping_zones (
    id               BIGSERIAL PRIMARY KEY,
    name_en          TEXT NOT NULL,
    name_my          TEXT NOT NULL,
    fee_satang       INT NOT NULL CHECK (fee_satang >= 0),
    free_over_satang INT CHECK (free_over_satang IS NULL OR free_over_satang >= 0),
    is_default       BOOLEAN NOT NULL DEFAULT false,   -- used for provinces not assigned to any zone
    sort_order       INT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX uq_shipping_zone_default ON shipping_zones (is_default) WHERE is_default;

CREATE TABLE shipping_zone_provinces (
    zone_id   BIGINT NOT NULL REFERENCES shipping_zones(id) ON DELETE CASCADE,
    province  TEXT NOT NULL UNIQUE,                     -- English province name
    PRIMARY KEY (zone_id, province)
);

CREATE TABLE coupons (
    id                  BIGSERIAL PRIMARY KEY,
    code                TEXT NOT NULL UNIQUE CHECK (code = upper(code)),
    type                TEXT NOT NULL CHECK (type IN ('percent', 'fixed')),
    value               INT NOT NULL CHECK (value > 0),   -- percent 1..100, or satang
    min_order_satang    INT NOT NULL DEFAULT 0,
    max_discount_satang INT,
    starts_at           TIMESTAMPTZ,
    ends_at             TIMESTAMPTZ,
    usage_limit         INT,
    per_user_limit      INT NOT NULL DEFAULT 1,
    used_count          INT NOT NULL DEFAULT 0,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (type <> 'percent' OR value <= 100)
);

CREATE TABLE orders (
    id                BIGSERIAL PRIMARY KEY,
    code              TEXT NOT NULL UNIQUE,              -- public order number, e.g. SY-7K2M9Q
    user_id           BIGINT NOT NULL REFERENCES users(id),
    status            TEXT NOT NULL CHECK (status IN ('pending_payment', 'payment_review', 'paid', 'confirmed',
                                                      'packing', 'shipped', 'delivered', 'cancelled', 'expired')),
    payment_method    TEXT NOT NULL CHECK (payment_method IN ('promptpay', 'cod')),
    subtotal_satang   INT NOT NULL,
    discount_satang   INT NOT NULL DEFAULT 0,
    shipping_satang   INT NOT NULL DEFAULT 0,
    vat_rate_bp       INT NOT NULL,                       -- basis points, 700 = 7%
    vat_satang        INT NOT NULL DEFAULT 0,
    total_satang      INT NOT NULL,
    coupon_id         BIGINT REFERENCES coupons(id) ON DELETE SET NULL,
    coupon_code       TEXT NOT NULL DEFAULT '',
    recipient_name    TEXT NOT NULL,
    phone             TEXT NOT NULL,
    email             TEXT NOT NULL,
    line1             TEXT NOT NULL,
    subdistrict       TEXT NOT NULL,
    district          TEXT NOT NULL,
    province          TEXT NOT NULL,
    postcode          TEXT NOT NULL,
    note              TEXT NOT NULL DEFAULT '',
    courier           TEXT NOT NULL DEFAULT '',
    tracking_number   TEXT NOT NULL DEFAULT '',
    tracking_url      TEXT NOT NULL DEFAULT '',
    payment_due_at    TIMESTAMPTZ,
    paid_at           TIMESTAMPTZ,
    shipped_at        TIMESTAMPTZ,
    delivered_at      TIMESTAMPTZ,
    cancelled_at      TIMESTAMPTZ,
    cancel_reason     TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user_created ON orders (user_id, created_at DESC);
CREATE INDEX idx_orders_status_created ON orders (status, created_at DESC);
CREATE INDEX idx_orders_pending_due ON orders (payment_due_at) WHERE status = 'pending_payment';

CREATE TABLE order_items (
    id                BIGSERIAL PRIMARY KEY,
    order_id          BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id        BIGINT REFERENCES products(id) ON DELETE SET NULL,
    variant_id        BIGINT REFERENCES product_variants(id) ON DELETE SET NULL,
    product_name_en   TEXT NOT NULL,
    product_name_my   TEXT NOT NULL,
    variant_name_en   TEXT NOT NULL,
    variant_name_my   TEXT NOT NULL,
    sku               TEXT NOT NULL,
    image_url         TEXT NOT NULL DEFAULT '',
    unit_price_satang INT NOT NULL,
    quantity          INT NOT NULL CHECK (quantity > 0)
);
CREATE INDEX idx_order_items_order ON order_items (order_id);

CREATE TABLE order_status_history (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status      TEXT NOT NULL,
    note        TEXT NOT NULL DEFAULT '',
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_history_order ON order_status_history (order_id, created_at);

CREATE TABLE payment_slips (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    file_path    TEXT NOT NULL,                  -- relative to UPLOAD_DIR/private, never publicly served
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    note         TEXT NOT NULL DEFAULT '',
    reviewed_by  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at  TIMESTAMPTZ,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_slips_order ON payment_slips (order_id, uploaded_at DESC);

CREATE TABLE coupon_redemptions (
    coupon_id   BIGINT NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
    order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (coupon_id, order_id)
);

CREATE TABLE inventory_movements (
    id          BIGSERIAL PRIMARY KEY,
    variant_id  BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    change      INT NOT NULL,
    reason      TEXT NOT NULL CHECK (reason IN ('order', 'cancel', 'expire', 'restock', 'adjust', 'damage')),
    note        TEXT NOT NULL DEFAULT '',
    order_id    BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    stock_after INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_movements_variant ON inventory_movements (variant_id, created_at DESC);

CREATE TABLE reviews (
    id          BIGSERIAL PRIMARY KEY,
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id    BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, user_id)
);
CREATE INDEX idx_reviews_product_status ON reviews (product_id, status);

CREATE TABLE wishlists (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, product_id)
);

CREATE TABLE settings (
    key    TEXT PRIMARY KEY,
    value  JSONB NOT NULL
);
INSERT INTO settings (key, value) VALUES
    ('shop_name', '"Sawyun Ecommerce"'),
    ('promptpay_id', '""'),                 -- phone (0812345678) or 13-digit tax ID
    ('promptpay_name', '""'),               -- account name shown under the QR
    ('vat_rate_bp', '700'),
    ('payment_window_minutes', '30'),
    ('cod_max_satang', '1000000'),          -- 10,000 THB
    ('support_email', '""'),
    ('support_phone', '""');

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,                  -- e.g. order.status, product.update
    entity      TEXT NOT NULL,
    entity_id   TEXT NOT NULL DEFAULT '',
    details     JSONB NOT NULL DEFAULT '{}',
    ip          TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_created ON audit_logs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs, settings, wishlists, reviews, inventory_movements, coupon_redemptions,
    payment_slips, order_status_history, order_items, orders, coupons, shipping_zone_provinces,
    shipping_zones, event_products, events, product_variants, product_images, products, categories,
    addresses, sessions, otp_codes, users CASCADE;
