// Package orders implements checkout, the order lifecycle, payments and the admin order views.
package orders

// Order statuses.
const (
	StatusPendingPayment = "pending_payment"
	StatusPaymentReview  = "payment_review"
	StatusPaid           = "paid"
	StatusConfirmed      = "confirmed"
	StatusPacking        = "packing"
	StatusShipped        = "shipped"
	StatusDelivered      = "delivered"
	StatusCancelled      = "cancelled"
	StatusExpired        = "expired"
)

// ValidStatus reports whether s is a known order status.
func ValidStatus(s string) bool {
	switch s {
	case StatusPendingPayment, StatusPaymentReview, StatusPaid, StatusConfirmed, StatusPacking,
		StatusShipped, StatusDelivered, StatusCancelled, StatusExpired:
		return true
	}
	return false
}

// IsFinalForCancel reports whether an order in status s can no longer be cancelled.
func IsFinalForCancel(s string) bool {
	return s == StatusShipped || s == StatusDelivered || s == StatusCancelled || s == StatusExpired
}

// CustomerCanCancel reports whether the customer may cancel an order in status s.
func CustomerCanCancel(s string) bool {
	return s == StatusPendingPayment || s == StatusPaymentReview || s == StatusConfirmed
}

// CanUploadSlip reports whether a slip may be uploaded.
func CanUploadSlip(status, method string) bool {
	return method == "promptpay" && status == StatusPendingPayment
}

// AdminCanTransition reports whether the admin status endpoint may move from → to.
// payment_review → paid happens only through slip approval, not this endpoint.
func AdminCanTransition(from, to string) bool {
	switch to {
	case StatusPacking:
		return from == StatusPaid || from == StatusConfirmed
	case StatusShipped:
		return from == StatusPacking
	case StatusDelivered:
		return from == StatusShipped
	case StatusCancelled:
		return ValidStatus(from) && !IsFinalForCancel(from)
	}
	return false
}
