package valueobject

import "testing"

func TestValidateHotel(t *testing.T) {
	name, addr, err := ValidateHotel("Hotel", "Address")
	if err != nil {
		t.Fatalf("expected valid hotel, got %v", err)
	}
	if name == "" || addr == "" {
		t.Fatalf("name/address should not be empty")
	}
	if _, _, err := ValidateHotel("", ""); err == nil {
		t.Fatalf("expected error for empty fields")
	}
}

func TestRoomTypeSpec(t *testing.T) {
	if err := RoomTypeSpec(2, 1000); err != nil {
		t.Fatalf("expected valid room type, got %v", err)
	}
	if err := RoomTypeSpec(0, 1000); err == nil {
		t.Fatalf("expected error for capacity 0")
	}
	if err := RoomTypeSpec(2, -1); err == nil {
		t.Fatalf("expected error for negative price")
	}
}

