# Functional specification

## Customer flows

1. **Discover:** land on a focused editorial storefront, understand the brand promise, and jump to the collection.
2. **Browse:** view the complete product catalog, live-search by name/category, filter by category, and sort by featured, price, or rating.
3. **Save:** toggle wishlist items. Saved IDs persist on the current device.
4. **Cart:** quick-add products, change quantities, remove lines, and see live subtotal, shipping threshold, and total. The cart persists between visits.
5. **Checkout:** enter contact and shipping details, submit a validated order to the Go API, and receive a confirmation state. No payment is charged in demo mode.
6. **Trust:** review delivery, sourcing, warranty, returns, ratings, and newsletter content without leaving the purchase flow.

## Operator flow

The dashboard icon opens an inventory overview backed by TanStack Table. It reports product count, total stock, low-stock count, and product-level price/rating/stock. `GET /api/orders` is protected by `X-Admin-Key` for integration with a future secured admin route.

## Application states

- Catalog loading uses the cached representative collection immediately while TanStack Query refreshes from the API.
- An unavailable API falls back to the embedded catalog so the storefront remains usable.
- Search/filter with zero results shows an explicit reset action.
- The cart has useful empty and populated states.
- Checkout uses native required/email validation, then a confirmation state.
- Orders are recalculated server-side; submitted client prices are never trusted.

## Acceptance checklist

- Responsive from 320px through large desktop.
- Keyboard-focusable native controls and labeled icon buttons.
- Search, categories, sorting, wishlist, cart quantity/removal, totals, checkout form, newsletter reset, and inventory table are interactive.
- Vue/TypeScript production build succeeds.
- API exposes health, catalog, product detail, order creation, and protected order listing.
