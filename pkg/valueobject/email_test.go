package valueobject

import "testing"

func TestNormalizeEmail(t *testing.T) {
	email, err := NormalizeEmail("  USER@example.COM ")
	if err != nil {
		t.Fatalf("expected valid email, got %v", err)
	}
	if email != "user@example.com" {
		t.Fatalf("expected lowercased trimmed email, got %s", email)
	}

	if _, err := NormalizeEmail("bad-email"); err == nil {
		t.Fatalf("expected error for invalid email")
	}
	if _, err := NormalizeEmail("   "); err == nil {
		t.Fatalf("expected error for empty email")
	}
}
