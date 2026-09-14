# Database model and relationships

The checked-in `database/schema.sql` is the production PostgreSQL model. The zero-dependency demo API uses JSON order persistence so it starts with only Go installed; migrate its repository layer to PostgreSQL using this schema for deployment at scale.

## Relationship map

```text
users 1 ── * addresses
users 1 ── * carts 1 ── * cart_items * ── 1 product_variants
users 1 ── * orders 1 ── * order_items * ── 0..1 product_variants
orders 1 ── * payments
categories 1 ── * products 1 ── * product_variants
categories 1 ── * categories (parent/child)
products 1 ── * product_images
products 1 ── * reviews * ── 0..1 users
users * ── * products (through wishlists)
```

## Table responsibilities

| Table | Responsibility | Important rules |
|---|---|---|
| `users` | Customer/admin identity | Unique normalized email; hashed passwords only |
| `addresses` | Reusable user delivery addresses | Cascade with user; one default enforced in service/partial index |
| `categories` | Navigation taxonomy | Self-reference supports nested categories |
| `products` | Shared merchandising content | Unique slug; lifecycle via `status` |
| `product_variants` | Sellable SKU, price and stock | Non-negative money/stock; unique SKU |
| `product_images` | Ordered product media | Product cascade; useful alt text required |
| `carts` / `cart_items` | Guest or customer basket | One variant per cart; positive quantity |
| `orders` / `order_items` | Immutable purchase snapshot | Copy name, SKU, and unit price at purchase time |
| `payments` | Provider transaction attempts | Never store raw card data |
| `reviews` | Moderated product feedback | Rating restricted to 1–5 |
| `wishlists` | User-to-product saved list | Composite key prevents duplicates |

## Money, inventory, and deletion

Store money as integer minor units (`price_cents`), never floating point. Reserve and decrement inventory in a database transaction when an order is accepted. Orders retain historical line snapshots if products/variants are removed. User deletion removes personal address/cart data but preserves financial orders with a nullable user link and a separately governed retention policy.
