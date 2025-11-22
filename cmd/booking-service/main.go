package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	bookinghttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/http"
	bookingnotification "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/notification"
	bookingpayment "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/payment"
	bookingrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/repository"
	bookingworker "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/worker"
	hotelrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/repository"
	bookinguc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	"github.com/ftryyln/hotel-booking-microservices/pkg/database"
	"github.com/ftryyln/hotel-booking-microservices/pkg/logger"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	db, err := database.NewGormPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	if err := bookingrepo.AutoMigrate(db); err != nil {
		log.Fatal("failed to run booking migrations", zap.Error(err))
	}
	if err := hotelrepo.AutoMigrate(db); err != nil {
		log.Fatal("failed to run hotel migrations", zap.Error(err))
	}

	repoFactory := bookingrepo.NewGormFactory(db)
	repo, err := repoFactory.CreateBookingRepository(bookingrepo.TypeGorm)
	if err != nil {
		log.Fatal("failed to create booking repository", zap.Error(err))
	}

	hRepo := hotelrepo.NewGormRepository(db)
	paymentClient := bookingpayment.NewHTTPGateway(cfg.PaymentServiceURL)
	notifier := bookingnotification.NewHTTPGateway(cfg.NotificationURL)
	service := bookinguc.NewService(repo, hRepo, paymentClient, notifier)
	handler := bookinghttp.NewHandler(service)

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWT(cfg.JWTSecret))
		r.Mount("/", handler.Routes())
	})

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	// Initialize and start auto-checkout scheduler
	scheduler := bookingworker.NewAutoCheckoutScheduler(service, log)
	if err := scheduler.Start(); err != nil {
		log.Fatal("failed to start auto-checkout scheduler", zap.Error(err))
	}
	defer scheduler.Stop()

	<-ctx.Done()
	log.Info("Shutting down gracefully...")
	scheduler.Stop()
	_ = srv.Stop(context.Background())
}
