package notification

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// Service wraps dispatcher implementation.
type Service struct {
	dispatcher domain.Dispatcher
	mu         sync.Mutex
	store      []dto.NotificationResponse
}

func NewService(dispatcher domain.Dispatcher) *Service {
	return &Service{dispatcher: dispatcher}
}

func (s *Service) Send(ctx context.Context, req dto.NotificationRequest) (dto.NotificationResponse, error) {
	if err := s.dispatcher.Dispatch(ctx, req.Target, req.Message); err != nil {
		return dto.NotificationResponse{}, err
	}
	return s.persist(req), nil
}

func (s *Service) persist(req dto.NotificationRequest) dto.NotificationResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := dto.NotificationResponse{
		ID:        uuid.New().String(),
		Target:    req.Target,
		Type:      req.Type,
		Message:   req.Message,
		CreatedAt: time.Now().UTC(),
	}
	s.store = append(s.store, record)
	return record
}

func (s *Service) List(_ context.Context) []dto.NotificationResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]dto.NotificationResponse, len(s.store))
	copy(out, s.store)
	return out
}

func (s *Service) Get(_ context.Context, id string) (dto.NotificationResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.store {
		if n.ID == id {
			return n, true
		}
	}
	return dto.NotificationResponse{}, false
}
