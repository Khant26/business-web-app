# Production checklist

Before accepting real purchases:

1. Replace the JSON repository with PostgreSQL using `database/schema.sql` and transactional stock reservation.
2. Add secure customer/admin authentication, role authorization, password reset, rate limits, CSRF strategy, and audited admin actions.
3. Integrate a PCI-compliant hosted payment flow (for example, provider checkout/Elements); process signed webhooks idempotently and never handle raw card details.
4. Move product media to object storage/CDN with responsive formats, real alt text, and image moderation.
5. Add tax/shipping services, currency rounding rules, order email, refund/return workflows, privacy/terms pages, and consent controls.
6. Put the API behind TLS, use a secret manager, restrict CORS, rotate the demo admin key, cap request rates, and configure structured logs/metrics/traces.
7. Add unit/integration/e2e tests for pricing, stock races, checkout failure, responsive layouts, accessibility, and payment webhook replay.
8. Run database backups and restore drills; define retention and deletion policies for customer data.
