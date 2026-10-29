package pricing

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestCompute(t *testing.T) {
	welcome := &Coupon{Code: "WELCOME10", Type: "percent", Value: 10, MaxDiscount: ptr[int64](20000), MinOrder: 30000, IsActive: true}
	save50 := &Coupon{Code: "SAVE50", Type: "fixed", Value: 5000, MinOrder: 50000, IsActive: true}
	bkk := &Zone{Fee: 4000, FreeOver: ptr[int64](100000)}
	flat := &Zone{Fee: 8000}

	cases := []struct {
		name string
		in   Input
		want Result
	}{
		{"no coupon, shipping charged", Input{Lines: []Line{{29000, 2}}, Zone: bkk, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 58000, Shipping: 4000, VAT: 4340, Total: 66340, CODAllowed: true}},
		{"percent coupon under cap", Input{Lines: []Line{{50000, 1}}, Coupon: welcome, Zone: bkk, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 50000, Discount: 5000, Shipping: 4000, VAT: 3430, Total: 52430, CODAllowed: true}},
		{"percent coupon capped", Input{Lines: []Line{{300000, 1}}, Coupon: welcome, Zone: bkk, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 300000, Discount: 20000, Shipping: 0, VAT: 19600, Total: 299600, CODAllowed: true}},
		{"percent floors", Input{Lines: []Line{{33333, 1}}, Coupon: welcome, VATRateBP: 0, CODMax: 1000000},
			Result{Subtotal: 33333, Discount: 3333, Total: 30000, CODAllowed: true}},
		{"fixed coupon", Input{Lines: []Line{{60000, 1}}, Coupon: save50, Zone: flat, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 60000, Discount: 5000, Shipping: 8000, VAT: 4410, Total: 67410, CODAllowed: true}},
		{"fixed coupon larger than subtotal", Input{Lines: []Line{{3000, 1}}, Coupon: &Coupon{Type: "fixed", Value: 5000}, VATRateBP: 700, CODMax: 1},
			Result{Subtotal: 3000, Discount: 3000, Total: 0, CODAllowed: true}},
		{"free shipping evaluated after discount", Input{Lines: []Line{{105000, 1}}, Coupon: save50, Zone: bkk, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 105000, Discount: 5000, Shipping: 0, VAT: 7000, Total: 107000, CODAllowed: true}},
		{"discount drops below free threshold", Input{Lines: []Line{{101000, 1}}, Coupon: save50, Zone: bkk, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 101000, Discount: 5000, Shipping: 4000, VAT: 7000, Total: 107000, CODAllowed: true}},
		{"vat rounds half up", Input{Lines: []Line{{50, 1}}, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 50, VAT: 4, Total: 54, CODAllowed: true}},
		{"vat rounds down below half", Input{Lines: []Line{{64, 1}}, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 64, VAT: 4, Total: 68, CODAllowed: true}},
		{"cod limit exceeded", Input{Lines: []Line{{1000000, 1}}, Zone: flat, VATRateBP: 700, CODMax: 1000000},
			Result{Subtotal: 1000000, Shipping: 8000, VAT: 70560, Total: 1078560, CODAllowed: false}},
		{"cod limit equal allowed", Input{Lines: []Line{{1000000, 1}}, VATRateBP: 0, CODMax: 1000000},
			Result{Subtotal: 1000000, Total: 1000000, CODAllowed: true}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Compute(c.in); got != c.want {
				t.Errorf("got %+v\nwant %+v", got, c.want)
			}
		})
	}
}

func TestValidateCoupon(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	base := func() *Coupon {
		return &Coupon{Code: "X", Type: "percent", Value: 10, MinOrder: 30000, PerUserLimit: 1, IsActive: true}
	}
	cases := []struct {
		name    string
		mod     func(c *Coupon)
		sub     int64
		used    int64
		want    string
		nilCoup bool
	}{
		{name: "valid", sub: 30000},
		{name: "missing", nilCoup: true, want: CouponNotFound},
		{name: "inactive", mod: func(c *Coupon) { c.IsActive = false }, sub: 50000, want: CouponInactive},
		{name: "not started", mod: func(c *Coupon) { c.StartsAt = ptr(now.Add(time.Hour)) }, sub: 50000, want: CouponNotStart},
		{name: "expired", mod: func(c *Coupon) { c.EndsAt = ptr(now) }, sub: 50000, want: CouponExpired},
		{name: "inside window", mod: func(c *Coupon) { c.StartsAt = ptr(now.Add(-time.Hour)); c.EndsAt = ptr(now.Add(time.Hour)) }, sub: 50000},
		{name: "min order", sub: 29999, want: CouponMinOrder},
		{name: "used up", mod: func(c *Coupon) { c.UsageLimit = ptr[int64](3); c.UsedCount = 3 }, sub: 50000, want: CouponUsedUp},
		{name: "user limit", sub: 50000, used: 1, want: CouponUserLimit},
		{name: "user limit 2 ok", mod: func(c *Coupon) { c.PerUserLimit = 2 }, sub: 50000, used: 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var cp *Coupon
			if !c.nilCoup {
				cp = base()
				if c.mod != nil {
					c.mod(cp)
				}
			}
			if got := ValidateCoupon(cp, c.sub, c.used, now); got != c.want {
				t.Errorf("got %q want %q", got, c.want)
			}
		})
	}
}
