package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Service coordinates registration/login use cases.
type Service struct {
	repo   domain.UserRepository
	issuer domain.TokenIssuer
}

func NewService(repo domain.UserRepository, issuer domain.TokenIssuer) *Service {
	return &Service{repo: repo, issuer: issuer}
}

var allowedRoles = map[string]struct{}{
	"customer": {},
	"admin":    {},
}

// Register creates new user and issues tokens.
func (s *Service) Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !strings.Contains(email, "@") {
		return dto.AuthResponse{}, errors.New("bad_request", "invalid email")
	}
	req.Email = email
	if req.Password == "" {
		return dto.AuthResponse{}, errors.New("bad_request", "password required")
	}
	if req.Role == "" {
		req.Role = "customer"
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if _, ok := allowedRoles[role]; !ok {
		return dto.AuthResponse{}, errors.New("bad_request", "invalid role")
	}
	req.Role = role

	if _, err := s.repo.FindByEmail(ctx, req.Email); err == nil {
		return dto.AuthResponse{}, errors.New("conflict", "email already used")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	user := domain.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hash),
		Role:      req.Role,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return dto.AuthResponse{}, err
	}

	return s.issueTokens(ctx, user)
}

// Login verifies credentials.
func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return dto.AuthResponse{}, errors.New("unauthorized", "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return dto.AuthResponse{}, errors.New("unauthorized", "invalid credentials")
	}

	return s.issueTokens(ctx, user)
}

// Me returns profile info.
func (s *Service) Me(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProfileResponse{}, errors.New("not_found", "user not found")
	}
	return dto.ProfileResponse{ID: user.ID.String(), Email: user.Email, Role: user.Role}, nil
}

func (s *Service) issueTokens(ctx context.Context, user domain.User) (dto.AuthResponse, error) {
	access, refresh, err := s.issuer.Generate(ctx, user)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	return dto.AuthResponse{AccessToken: access, RefreshToken: refresh}, nil
}
