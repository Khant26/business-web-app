package orders

import "testing"

func TestAdminCanTransition(t *testing.T) {
	all := []string{StatusPendingPayment, StatusPaymentReview, StatusPaid, StatusConfirmed, StatusPacking,
		StatusShipped, StatusDelivered, StatusCancelled, StatusExpired}
	allowed := map[[2]string]bool{
		{StatusPaid, StatusPacking}:             true,
		{StatusConfirmed, StatusPacking}:        true,
		{StatusPacking, StatusShipped}:          true,
		{StatusShipped, StatusDelivered}:        true,
		{StatusPendingPayment, StatusCancelled}: true,
		{StatusPaymentReview, StatusCancelled}:  true,
		{StatusPaid, StatusCancelled}:           true,
		{StatusConfirmed, StatusCancelled}:      true,
		{StatusPacking, StatusCancelled}:        true,
	}
	for _, from := range all {
		for _, to := range all {
			want := allowed[[2]string{from, to}]
			if got := AdminCanTransition(from, to); got != want {
				t.Errorf("AdminCanTransition(%s → %s) = %v, want %v", from, to, got, want)
			}
		}
	}
	if AdminCanTransition("bogus", StatusCancelled) {
		t.Error("unknown status must not transition")
	}
}

func TestCustomerCanCancel(t *testing.T) {
	cases := map[string]bool{
		StatusPendingPayment: true, StatusPaymentReview: true, StatusConfirmed: true,
		StatusPaid: false, StatusPacking: false, StatusShipped: false, StatusDelivered: false,
		StatusCancelled: false, StatusExpired: false,
	}
	for s, want := range cases {
		if CustomerCanCancel(s) != want {
			t.Errorf("CustomerCanCancel(%s) != %v", s, want)
		}
	}
}

func TestCanUploadSlip(t *testing.T) {
	if !CanUploadSlip(StatusPendingPayment, "promptpay") {
		t.Error("pending promptpay should accept slips")
	}
	if CanUploadSlip(StatusPendingPayment, "cod") || CanUploadSlip(StatusPaymentReview, "promptpay") || CanUploadSlip(StatusPaid, "promptpay") {
		t.Error("slip upload allowed in wrong state")
	}
}

func TestOrderCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		c := newOrderCode()
		if len(c) != 9 || c[:3] != "SY-" {
			t.Fatalf("bad code %q", c)
		}
		seen[c] = true
	}
	if len(seen) < 99 {
		t.Fatal("order codes collide too often")
	}
}

func TestNormalizeItems(t *testing.T) {
	items, err := normalizeItems([]ItemReq{{VariantID: 2, Quantity: 1}, {VariantID: 1, Quantity: 2}, {VariantID: 2, Quantity: 3}})
	if err != nil || len(items) != 2 || items[0].VariantID != 1 || items[1].Quantity != 4 {
		t.Fatalf("unexpected %+v %v", items, err)
	}
	if _, err := normalizeItems(nil); err == nil {
		t.Fatal("empty cart must fail")
	}
	if _, err := normalizeItems([]ItemReq{{VariantID: 1, Quantity: 0}}); err == nil {
		t.Fatal("zero quantity must fail")
	}
}
