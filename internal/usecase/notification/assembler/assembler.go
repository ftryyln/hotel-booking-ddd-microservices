package assembler

import (
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Command represents inbound send request.
type Command struct {
	Type    string
	Target  string
	Message string
}

// FromRequest validates and builds command.
func FromRequest(req dto.NotificationRequest) (Command, error) {
	if req.Target == "" {
		return Command{}, errors.New("bad_request", "target is required")
	}
	if req.Message == "" {
		return Command{}, errors.New("bad_request", "message is required")
	}
	return Command{Type: req.Type, Target: req.Target, Message: req.Message}, nil
}

// ToDomain builds domain Notification with generated metadata.
func ToDomain(cmd Command, id string, createdAt time.Time) domain.Notification {
	return domain.Notification{
		ID:        id,
		Type:      cmd.Type,
		Target:    cmd.Target,
		Message:   cmd.Message,
		CreatedAt: createdAt,
	}
}

// ToResponse maps domain Notification to DTO.
func ToResponse(n domain.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Target:    n.Target,
		Message:   n.Message,
		CreatedAt: n.CreatedAt,
	}
}

// ToResponses maps slice to DTOs.
func ToResponses(list []domain.Notification) []dto.NotificationResponse {
	out := make([]dto.NotificationResponse, 0, len(list))
	for _, n := range list {
		out = append(out, ToResponse(n))
	}
	return out
}
