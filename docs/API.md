# HTTP API

Base path: `/api`. JSON responses use `Content-Type: application/json`.

| Method | Route | Behavior |
|---|---|---|
| `GET` | `/health` | Returns `{ "status": "ok" }` |
| `GET` | `/products?q=&category=` | Returns catalog; optional case-insensitive filters |
| `GET` | `/products/{id}` | Returns one product or 404 |
| `POST` | `/orders` | Validates customer, address and quantities; recalculates totals; persists order |
| `GET` | `/orders` | Lists orders; requires `X-Admin-Key` |

## Create order

```json
{
  "email": "buyer@example.com",
  "customer_name": "Taylor North",
  "shipping_address": "14 Market Street, Portland 04101",
  "items": [{ "product_id": 1, "quantity": 2 }]
}
```

The response adds an order ID, canonical unit prices and totals, status, and UTC creation time. Invalid products, out-of-range quantities, missing customer data, and malformed JSON return 4xx JSON errors.
