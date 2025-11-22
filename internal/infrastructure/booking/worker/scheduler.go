package worker

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	bookinguc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
)

// AutoCheckoutScheduler handles automated checkout processing.
type AutoCheckoutScheduler struct {
	cron    *cron.Cron
	service *bookinguc.Service
	logger  *zap.Logger
}

// NewAutoCheckoutScheduler creates a new scheduler instance.
func NewAutoCheckoutScheduler(service *bookinguc.Service, logger *zap.Logger) *AutoCheckoutScheduler {
	return &AutoCheckoutScheduler{
		cron:    cron.New(),
		service: service,
		logger:  logger,
	}
}

// Start initializes and starts the cron scheduler.
// Default schedule: daily at 10:00 AM (0 10 * * *)
func (s *AutoCheckoutScheduler) Start() error {
	// Schedule auto-checkout to run daily at 10:00 AM
	_, err := s.cron.AddFunc("0 10 * * *", func() {
		s.logger.Info("🏃 Running auto-checkout scheduler...")
		if err := s.runAutoCheckout(); err != nil {
			s.logger.Error("❌ Auto-checkout failed", zap.Error(err))
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info("✅ Auto-checkout scheduler started (runs daily at 10:00 AM)")
	return nil
}

// Stop gracefully stops the cron scheduler.
func (s *AutoCheckoutScheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
		s.logger.Info("🛑 Auto-checkout scheduler stopped")
	}
}

// runAutoCheckout executes the auto-checkout logic.
func (s *AutoCheckoutScheduler) runAutoCheckout() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	count, err := s.service.AutoCheckout(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		s.logger.Info("✅ No bookings to auto-checkout today")
	} else {
		s.logger.Info("✅ Auto-checkout completed", zap.Int("processed_bookings", count))
	}

	return nil
}
