package valueobject

import "testing"

func TestValidateRole(t *testing.T) {
	role, err := ParseRole("admin")
	if err != nil || role != RoleAdmin {
		t.Fatalf("expected admin role, got %v err=%v", role, err)
	}
	if _, err := ParseRole("bad"); err == nil {
		t.Fatalf("expected error for invalid role")
	}
}
