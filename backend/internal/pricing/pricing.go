// Package pricing contains the pure, authoritative checkout price calculation.
package pricing

import "time"

// Line is one cart line.
type Line struct {
	UnitPrice int64
	Qty       int64
}

// Coupon is the subset of coupon data needed for validation and discount.
type Coupon struct {
	Code         string
	Type         string // percent | fixed
	Value        int64
	MinOrder     int64
	MaxDiscount  *int64
	StartsAt     *time.Time
	EndsAt       *time.Time
	UsageLimit   *int64
	PerUserLimit int64
	UsedCount    int64
	IsActive     bool
}

// Zone is the shipping zone for the destination province.
type Zone struct {
	Fee      int64
	FreeOver *int64
}

// Coupon message codes.
const (
	CouponNotFound  = "coupon_not_found"
	CouponInactive  = "coupon_inactive"
	CouponNotStart  = "coupon_not_started"
	CouponExpired   = "coupon_expired"
	CouponMinOrder  = "coupon_min_order"
	CouponUsedUp    = "coupon_used_up"
	CouponUserLimit = "coupon_user_limit"
)

// ValidateCoupon checks whether c can be applied. It returns "" when valid or a message code.
func ValidateCoupon(c *Coupon, subtotal, userRedemptions int64, now time.Time) string {
	switch {
	case c == nil:
		return CouponNotFound
	case !c.IsActive:
		return CouponInactive
	case c.StartsAt != nil && now.Before(*c.StartsAt):
		return CouponNotStart
	case c.EndsAt != nil && !now.Before(*c.EndsAt):
		return CouponExpired
	case c.UsageLimit != nil && c.UsedCount >= *c.UsageLimit:
		return CouponUsedUp
	case c.PerUserLimit > 0 && userRedemptions >= c.PerUserLimit:
		return CouponUserLimit
	case subtotal < c.MinOrder:
		return CouponMinOrder
	}
	return ""
}

// Discount computes the coupon discount for subtotal.
func Discount(c *Coupon, subtotal int64) int64 {
	if c == nil || subtotal <= 0 {
		return 0
	}
	var d int64
	switch c.Type {
	case "percent":
		d = subtotal * c.Value / 100
		if c.MaxDiscount != nil && d > *c.MaxDiscount {
			d = *c.MaxDiscount
		}
	case "fixed":
		d = min(c.Value, subtotal)
	}
	return max(0, min(d, subtotal))
}

// Input is everything needed to price an order.
type Input struct {
	Lines     []Line
	Coupon    *Coupon // already validated; nil means no discount
	Zone      *Zone   // nil means no shipping fee
	VATRateBP int64   // basis points, 700 = 7%
	CODMax    int64
}

// Result is the priced order.
type Result struct {
	Subtotal   int64
	Discount   int64
	Shipping   int64
	VAT        int64
	Total      int64
	CODAllowed bool
}

// Compute applies the pricing rules from docs/API.md.
func Compute(in Input) Result {
	var r Result
	for _, l := range in.Lines {
		r.Subtotal += l.UnitPrice * l.Qty
	}
	r.Discount = Discount(in.Coupon, r.Subtotal)
	afterDiscount := r.Subtotal - r.Discount
	if in.Zone != nil {
		r.Shipping = in.Zone.Fee
		if in.Zone.FreeOver != nil && afterDiscount >= *in.Zone.FreeOver {
			r.Shipping = 0
		}
	}
	if r.Subtotal == 0 {
		r.Shipping = 0
	}
	r.VAT = VAT(afterDiscount+r.Shipping, in.VATRateBP)
	r.Total = afterDiscount + r.Shipping + r.VAT
	r.CODAllowed = r.Total <= in.CODMax
	return r
}

// VAT returns round-half-up(base × bp / 10000).
func VAT(base, bp int64) int64 {
	if base <= 0 || bp <= 0 {
		return 0
	}
	return (base*bp + 5000) / 10000
}
