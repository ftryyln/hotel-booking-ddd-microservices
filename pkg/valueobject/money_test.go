package valueobject

import "testing"

func TestNewMoney(t *testing.T) {
	_, err := NewMoney(1000, "IDR")
	if err != nil {
		t.Fatalf("expected valid money, got %v", err)
	}
	if _, err := NewMoney(-1, "IDR"); err == nil {
		t.Fatalf("expected error for negative amount")
	}
	if _, err := NewMoney(1000, ""); err == nil {
		t.Fatalf("expected error for empty currency")
	}
}
