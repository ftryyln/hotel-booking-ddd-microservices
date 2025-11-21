package valueobject

import "testing"

func TestValidateBookingStatus(t *testing.T) {
	s, err := ValidateBookingStatus("pending_payment")
	if err != nil || s != StatusPendingPayment {
		t.Fatalf("expected pending_payment, got %v err=%v", s, err)
	}
	if _, err := ValidateBookingStatus("bad"); err == nil {
		t.Fatalf("expected error for invalid booking status")
	}
}

func TestBookingStatusTransition(t *testing.T) {
	curr, _ := ValidateBookingStatus("pending_payment")
	next, _ := ValidateBookingStatus("confirmed")
	if err := curr.CanTransition(next); err != nil {
		t.Fatalf("expected transition ok, got %v", err)
	}

	// confirmed -> cancelled is allowed, so test a disallowed jump confirmed -> completed
	completed, _ := ValidateBookingStatus("completed")
	if err := next.CanTransition(completed); err == nil {
		t.Fatalf("expected invalid transition to fail")
	}
}

func TestValidatePaymentStatus(t *testing.T) {
	if _, err := ValidatePaymentStatus("pending"); err != nil {
		t.Fatalf("expected pending ok: %v", err)
	}
	if _, err := ValidatePaymentStatus("bad"); err == nil {
		t.Fatalf("expected error for invalid payment status")
	}
}

func TestNormalizeRoomStatus(t *testing.T) {
	if _, err := NormalizeRoomStatus(""); err != nil {
		t.Fatalf("empty defaults to available, got err %v", err)
	}
	if _, err := NormalizeRoomStatus("available"); err != nil {
		t.Fatalf("available should be valid")
	}
	if _, err := NormalizeRoomStatus("bad"); err == nil {
		t.Fatalf("expected error for invalid room status")
	}
}
