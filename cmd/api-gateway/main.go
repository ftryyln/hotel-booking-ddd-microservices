package main

// @title Hotel Booking Microservices API
// @version 1.0
// @description Aggregated API surface for hotel booking platform.
// @host localhost:8088
// @BasePath /gateway
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/gateway"
	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	"github.com/ftryyln/hotel-booking-microservices/pkg/logger"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	handler := gateway.NewHandler(cfg.BookingServiceURL, cfg.PaymentServiceURL, cfg.AggregateTargetURL, cfg.RateLimitPerMinute)

	authTarget, _ := url.Parse(cfg.AuthServiceURL)
	if authTarget.Path == "" || authTarget.Path == "/" {
		authTarget.Path = "/auth"
	}
	authProxy := httputil.NewSingleHostReverseProxy(authTarget)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Mount("/gateway/auth", http.StripPrefix("/gateway/auth", authProxy))
	r.Group(func(router chi.Router) {
		router.Use(middleware.JWT(cfg.JWTSecret))
		router.Mount("/gateway", handler.Routes())
	})

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
