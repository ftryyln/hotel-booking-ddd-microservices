package utils

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// Envelope standardizes API responses.
type Envelope struct {
	Data any  `json:"data,omitempty"`
	Meta Meta `json:"meta"`
}

// Meta carries response metadata.
type Meta struct {
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestId"`
	Count     int    `json:"count,omitempty"`
}

// Resource represents a single resource item.
type Resource struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Attributes any         `json:"attributes"`
	Links      ResourceURL `json:"links"`
}

// ResourceURL keeps self reference.
type ResourceURL struct {
	Self string `json:"self"`
}

// NewResource builds a resource wrapper.
func NewResource(id, typ, self string, attrs any) Resource {
	return Resource{
		ID:         id,
		Type:       typ,
		Attributes: attrs,
		Links:      ResourceURL{Self: self},
	}
}

// Respond writes a standardized envelope.
func Respond(w http.ResponseWriter, status int, message string, data any) {
	env := Envelope{
		Data: data,
		Meta: Meta{
			Message:   message,
			RequestID: requestIDFrom(message),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// RespondWithCount writes envelope including item count (for list endpoints).
func RespondWithCount(w http.ResponseWriter, status int, message string, data any, count int) {
	env := Envelope{
		Data: data,
		Meta: Meta{
			Message:   message,
			RequestID: requestIDFrom(message),
			Count:     count,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// requestIDFrom generates a UUID; message used only to vary seed (timestamp added).
func requestIDFrom(_ string) string {
	return uuid.New().String()
}
