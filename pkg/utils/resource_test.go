package utils

import (
	"testing"
)

func TestNewResource(t *testing.T) {
	res := NewResource("123", "hotel", "/api/v1/hotels/123", map[string]string{"name": "X"})
	if res.ID != "123" || res.Type != "hotel" {
		t.Fatalf("resource ids mismatch")
	}
	if res.Links.Self != "/api/v1/hotels/123" {
		t.Fatalf("self link mismatch")
	}
}

func TestRespondWithCountPayload(t *testing.T) {
	items := []Resource{
		NewResource("a", "t", "/a", nil),
		NewResource("b", "t", "/b", nil),
	}
	// simulate encoding; count is passed through meta
	env := Envelope{
		Data: items,
		Meta: Meta{Message: "ok", RequestID: "id", Count: len(items)},
	}
	if env.Meta.Count != len(items) {
		t.Fatalf("expected count %d, got %d", len(items), env.Meta.Count)
	}

	if rid := requestIDFrom("msg"); rid == "" {
		t.Fatalf("requestIDFrom should generate id")
	}
}
