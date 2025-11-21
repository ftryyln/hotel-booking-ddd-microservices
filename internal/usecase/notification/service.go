package notification

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// Service wraps dispatcher implementation.
type Service struct {
	dispatcher domain.Dispatcher
	mu         sync.Mutex
	store      []domain.Notification
}

func NewService(dispatcher domain.Dispatcher) *Service {
	return &Service{dispatcher: dispatcher}
}

func (s *Service) Send(ctx context.Context, cmd assembler.Command) (domain.Notification, error) {
	if err := s.dispatcher.Dispatch(ctx, cmd.Target, cmd.Message); err != nil {
		return domain.Notification{}, err
	}
	return s.persist(cmd), nil
}

func (s *Service) persist(cmd assembler.Command) domain.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := domain.Notification{
		ID:        uuid.New().String(),
		Target:    cmd.Target,
		Type:      cmd.Type,
		Message:   cmd.Message,
		CreatedAt: time.Now().UTC(),
	}
	s.store = append(s.store, record)
	return record
}

func (s *Service) List(_ context.Context, opts query.Options) []domain.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	norm := opts.Normalize(50)
	start := norm.Offset
	if start > len(s.store) {
		start = len(s.store)
	}
	end := start + norm.Limit
	if end > len(s.store) {
		end = len(s.store)
	}
	out := make([]domain.Notification, end-start)
	copy(out, s.store[start:end])
	return out
}

func (s *Service) Get(_ context.Context, id string) (domain.Notification, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.store {
		if n.ID == id {
			return n, true
		}
	}
	return domain.Notification{}, false
}
