package valueobject

import (
	"testing"
	"time"
)

func TestNewDateRange(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	dr, err := NewDateRange(start, end)
	if err != nil {
		t.Fatalf("expected valid range, got %v", err)
	}
	if nights := dr.Nights(); nights != 2 {
		t.Fatalf("expected 2 nights, got %d", nights)
	}

	if _, err := NewDateRange(end, start); err == nil {
		t.Fatalf("expected error when start after end")
	}
	if _, err := NewDateRange(time.Time{}, end); err == nil {
		t.Fatalf("expected error when start is zero")
	}
}
