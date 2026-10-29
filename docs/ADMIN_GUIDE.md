# Admin and staff guide

This guide is for the people who run the shop. Sign in with your staff or admin email (see
[USER_GUIDE.md](USER_GUIDE.md#4-signing-in-no-password)) and press **Admin** in the header, or go to
`/admin`. Pages marked ★ are visible to **admins only**; staff do not see them in the menu and the
server refuses their requests.

The menu is grouped as **Sales** (Dashboard, Orders), **Catalog** (Products, Inventory, Categories),
**Marketing** (Events, Coupons, Reviews) and **Administration** (★ Shipping zones, ★ Settings,
★ Users & staff, ★ Audit log). **View store** opens the shop.

All prices in admin forms are typed in **baht** (e.g. `1290` or `1,290.50`); they are saved as satang.
Product prices are **without VAT** — VAT is added at checkout. Times for events and coupons are
**Bangkok time**. Every change you save is recorded in the audit log with your email.

---

## Dashboard

| Card | Meaning |
|---|---|
| Today's sales | Sum of totals of today's orders (Bangkok date) that are paid, confirmed, packing, shipped or delivered. |
| Today's orders | All orders created today, any status. |
| Slips to review | Payment slips waiting for a decision. **Handle these first.** |
| To pack | Orders in *Paid* or *Confirmed* — ready to pack. |
| Low stock | Active variants at or below their low-stock threshold. |

Below: a 7-day sales chart, the 10 latest orders and the 10 lowest-stock items, each linking to the
detail page.

## Orders

**List:** tabs filter by status. Search by order code, customer email, recipient name or phone.
Filter by payment method and a date range (From / To, Bangkok dates). A **Slip to review** badge marks
orders with a pending slip.

**Order detail** shows the customer, items with SKU and prices, shipping address, customer note,
money breakdown, payment slips, tracking, status actions and full history.

### Verifying a PromptPay slip

1. Open the order (status *Verifying payment*). In **Payment slips** press **View slip**.
2. In your bank app or internet banking, find the incoming transfer. Check **all** of:
   - the amount equals the order **Total** exactly;
   - the receiving account is the shop's PromptPay account;
   - the date/time is after the order was placed;
   - the transaction reference has not been used for another order (search old orders if unsure).
3. If everything matches, press **Approve payment** (optional note) → the order becomes **Paid**.
4. If not, press **Reject slip** and write a clear reason — the customer sees it, e.g. "Amount is ฿2,400, order total is ฿2,450.30". The order goes back to *Awaiting payment* with a **fresh** payment window so the customer can upload again.

Never approve a slip only because the image looks real — always confirm the money arrived.

### Moving an order forward

Buttons appear only when the step is allowed:

| From | Button | To | Notes |
|---|---|---|---|
| Paid / Confirmed (COD) | **Start packing** | Packing | |
| Packing | **Mark as shipped** | Shipped | Save tracking first (below), otherwise you get "Add a tracking number before marking as shipped". |
| Shipped | **Mark as delivered** | Delivered | Increases "sold" counts and lets the customer write reviews. |
| Any status before Shipped | **Cancel order** | Cancelled | Optional reason (default "Cancelled by shop"). Stock is restored and the coupon use released. |

*Verifying payment → Paid* only happens through **Approve payment**. Shipped, delivered, cancelled
and expired orders are closed.

If you cancel an order the customer already paid, arrange the refund outside the system (bank
transfer) and write that in the reason.

### Tracking

In **Shipment tracking** enter the **Courier** (e.g. Kerry Express, Flash, Thailand Post), the
**Tracking number** and optionally a **Tracking link** (must start with `http://` or `https://`) and
press **Save tracking**. Tracking can be saved once the order is Paid/Confirmed or later (not while
awaiting or verifying payment, and not on cancelled/expired orders). The customer sees it on their
order page.

## Products

**List:** search by name, slug or SKU; filter by status and category. Columns show variants, total
stock and the lowest price.

**New / edit product:**

1. **Product details:** slug (URL part, lowercase letters, numbers, dashes — must be unique), English name (required) and Myanmar name, descriptions in both languages, category, **status** and **Featured**.
   - *Draft* — hidden from the shop (use while preparing).
   - *Active* — visible and for sale.
   - *Archived* — hidden; kept for order history. **Archive** in the list does this.
2. **Events:** tick the events this product should appear in.
3. Press **Save**. Images and variants can be added only after the first save.
4. **Images:** upload JPG/PNG/WebP up to 5 MB. The first image is the **Cover** shown on product cards. Use **Move earlier / Move later** to reorder, or delete. (Uploads are limited to 20 per hour per person.)
5. **Variants:** at least one is needed for customers to buy. Each has a **SKU** (unique, saved in capitals), option names in both languages (e.g. "Red / M"), **Price**, optional **Compare-at price** (the old price, must be higher; shows a discount badge), **Low-stock alert at** (default 5), **Initial stock** (only when creating), **Available for sale** and sort order.
   - To change stock later use **Inventory**, not the variant form.
   - **Remove variant** deletes it, or only deactivates it if it was ever ordered.

The product card price is the cheapest active variant.

## Inventory

**Stock levels** tab: every variant of non-archived products, low-stock rows first. Search by SKU or
product name; tick *Low stock only*.

**Adjust** a row: enter a **Change** (positive to add, negative to remove), a **Reason** and an
optional note. The dialog shows the new stock before you save; stock can never go below 0.

| Reason | Use for |
|---|---|
| Restock | New goods arrived. |
| Correction | Stock count differs from the system. |
| Damaged / lost | Broken, expired or missing items. |

**Movements** tab: the full ledger — every order, cancel, expiry, restock, correction and damage with
the change, stock after, order code, who did it and when. Open it from a row to see one variant's
history. If a number looks wrong, the movements explain why.

## Categories

Create or edit with slug, English/Myanmar names, sort order (lower first) and Active. Upload an image
after saving. A category that still has products (including archived ones) cannot be deleted — move
the products or make the category inactive.

## Events

An event is a promotion page such as "11.11 Sale". Fields: slug, titles and descriptions in both
languages, **Starts** / **Ends** (Bangkok time), Active, sort order and **Products** (search and
add). Upload a wide **banner** (about 1600 × 600 px) after saving. The list shows *Scheduled*,
*Running* or *Ended*. Only running events appear on the home page; deleting an event does not delete
its products.

Events do **not** change prices by themselves. For a real discount, set compare-at prices on the
variants or create a coupon.

## Coupons

| Field | Meaning |
|---|---|
| Code | What customers type; 3–32 letters, numbers, `-` or `_`; saved in capitals. |
| Type / Discount | *Percent off* (1–100) or *Fixed amount* (baht). |
| Minimum order | Subtotal needed before discount. |
| Maximum discount | Optional cap for percent coupons. |
| Total uses | Leave empty for unlimited. |
| Uses per customer | Default 1. `0` means unlimited. |
| Starts / Ends | Optional, Bangkok time. Leave empty for no limit. |
| Active | Turn off to stop the coupon immediately. |

The **Used** column counts orders using the coupon; cancelled and expired orders give the use back.
Deleting a coupon that was used only deactivates it (so old orders stay correct).

## Reviews

Tabs: *Pending*, *Approved*, *Rejected*. Read the comment and press **Approve** or **Reject**. Only
approved reviews are public and count toward the product's star rating, which is recalculated
immediately. You can change your decision later. Reject spam, offensive text and personal data.

## ★ Shipping zones

Each zone has English/Myanmar names, a **Delivery fee**, an optional **Free delivery over** amount
(compared with the subtotal after discount, before VAT), a sort order and a list of **Provinces**.

- A province can be in only one zone; the picker shows which zone already has it.
- Exactly one zone should be the **Default zone** — it applies to every province not in any zone. The page tells you how many provinces are unassigned.
- The default zone cannot be deleted. Make another zone the default first.

Changes apply to new quotes and orders immediately; existing orders keep their shipping fee.

## ★ Settings

| Setting | Notes |
|---|---|
| Shop name | Shown to customers. |
| PromptPay ID | The shop's mobile number (e.g. `0812345678`) or 13-digit tax ID. **Must be correct** — customers pay to whatever is here. After changing it, scan a test QR with your own banking app. If empty, PromptPay orders cannot be paid. |
| Account name | Shown under the QR so customers can check before paying. |
| VAT (%) | Default 7. Applies to new quotes/orders. |
| Payment window (minutes) | 5–10,080. Unpaid PromptPay orders expire after this; a rejected slip also gets this much new time. |
| Cash on delivery limit | COD is offered only for totals up to this amount. Set 0 to effectively disable COD. |
| Support email / phone | Shown in the shop footer. |

## ★ Users & staff

- Search users by email or name; filter by role. The list shows order counts and join dates.
- **Add staff:** enter an email, optional name and role (*Staff* or *Admin*). If the email already has an account it is promoted. The person then signs in with an email code as usual.
- Change a user's **role** (Customer / Staff / Admin) or **Disable** them. Disabling signs them out everywhere and blocks sign-in; **Enable** reverses it.
- You cannot change your own role or disable yourself. The owner email in the server setting `ADMIN_EMAIL` is always restored as admin when the server restarts.

Give the **Staff** role to people who handle orders and products; keep **Admin** for the owner and the
senior developer. Disable accounts of people who leave on their last day.

## ★ Audit log

A read-only list, newest first: time, user, action (e.g. `order.status`, `slip.approve`,
`inventory.adjust`, `settings.update`), entity and id, IP, and **Show details** for the exact
values. Use it to answer "who changed this, and when?".

---

## Daily workflow checklist

**Morning (and every few hours during the day)**

- [ ] Open the **Dashboard**.
- [ ] **Slips to review:** approve or reject every pending slip after checking the bank account.
- [ ] **To pack:** open Orders → *Paid* and *Confirmed* tabs; press **Start packing** for each order you pack.
- [ ] For parcels handed to the courier: **Save tracking**, then **Mark as shipped**.

**Afternoon**

- [ ] Orders → *Shipped*: mark orders **Delivered** when the courier tracking says delivered (or COD money was received).
- [ ] **Low stock:** check the list; order goods; when they arrive, **Inventory → Adjust → Restock**.
- [ ] **Reviews → Pending:** approve or reject.

**Weekly**

- [ ] Look at *Expired* and *Cancelled* orders for patterns (e.g. customers struggling to pay).
- [ ] Check events and coupons that start or end soon.
- [ ] (Admin) Glance through the **Audit log** for unexpected changes, and review staff accounts.
- [ ] (Admin/developer) Confirm that backups ran (see [DEPLOYMENT.md](DEPLOYMENT.md#7-backups-and-restore)).

## Common questions

| Problem | Answer |
|---|---|
| "Too many attempts. Try again in N seconds" while uploading images | Uploads are limited to 20 per hour per person. Wait, or ask another staff member. |
| A customer says they paid but the order expired | Check the bank. If the money arrived, ask the customer to place a new order and approve it with a note, or handle a refund. Expired orders cannot be revived. |
| A product does not appear in the shop | Its status must be *Active*. An active product without any *Available for sale* variant still appears but cannot be bought (it shows as out of stock). |
| "That conflicts with existing data" | The slug, SKU or coupon code already exists. Choose another. |
| "It's still in use and can't be deleted" | A category still has products. |
