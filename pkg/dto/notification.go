package dto

import "time"

// NotificationRequest is sent to notification service.
type NotificationRequest struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// NotificationResponse captures stored notification metadata.
type NotificationResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Target    string    `json:"target"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
