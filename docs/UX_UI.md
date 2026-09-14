# UX and UI system

## Experience goal

Northstar is a calm, design-led general-goods shop. The interface helps a buyer move from discovery to a small, confident selection without visual noise. Product tools begin directly beneath the first editorial section; checkout stays in context as a drawer and modal rather than forcing unnecessary navigation.

## Visual language

- **Palette:** warm paper `#f5f2eb`, near-black green `#1d211c`, orange accent `#e45b36`, soft sage, clay, ochre, and slate product fields.
- **Typography:** Instrument Serif provides expressive editorial headings; DM Sans keeps controls and details crisp.
- **Shape:** mostly square surfaces and fine borders communicate durable, practical objects. Circular product forms and the orange accent provide warmth.
- **Spacing:** a fluid outer gutter and generous 70–90px section rhythm on desktop; compact stacked layouts on mobile.
- **Motion:** only purposeful drawer fades/slides and quick-add reveal. Reduced visual distraction is intentional.

## Responsive behavior

- Desktop uses a two-column hero and four-column catalog.
- Tablet stacks the hero and uses two product columns.
- Mobile uses one product column, persistent visible quick-add actions, scrollable category pills, single-column principles/forms, and a full-width cart drawer.

## Interaction rationale

- Search is always visible in the catalog and header search focuses it.
- Category chips expose available choices instead of hiding them in a menu.
- Cart totals update immediately and explain the free-shipping threshold.
- Wishlist and cart are device-persistent to protect intent across refreshes.
- Checkout explicitly says demo/no charge, preventing false payment expectations.
- The dashboard is intentionally lightweight but proves the operator information architecture and TanStack Table integration.

## Accessibility notes

Semantic headings, labels, native form validation, descriptive icon-button labels, readable contrast, responsive tap targets, and non-color stock text are included. A production pass should add focus trapping to overlays, Escape handling, automated axe checks, and product photography with finalized alt text.
