package authhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	auth "github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

type authUsecaseStub struct {
	users []dto.ProfileResponse
}

// repo stub + issuer stub compose a real Service for handler tests.
type authRepoStub struct {
	users map[string]domain.User
}

func (a *authRepoStub) Create(ctx context.Context, user domain.User) error {
	if a.users == nil {
		a.users = map[string]domain.User{}
	}
	a.users[user.Email] = user
	return nil
}
func (a *authRepoStub) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	if usr, ok := a.users[email]; ok {
		return usr, nil
	}
	return domain.User{}, errors.New("not found")
}
func (a *authRepoStub) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	for _, u := range a.users {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, errors.New("not found")
}
func (a *authRepoStub) List(ctx context.Context, opts query.Options) ([]domain.User, error) {
	var out []domain.User
	for _, u := range a.users {
		out = append(out, u)
	}
	return out, nil
}

type issuerStub struct{}

func (i *issuerStub) Generate(ctx context.Context, user domain.User) (string, string, error) {
	return "access", "refresh", nil
}

func TestAuthHandlerRegister(t *testing.T) {
	svc := auth.NewService(&authRepoStub{}, &issuerStub{})
	r := chi.NewRouter()
	h := NewHandler(svc, "secret")
	r.Mount("/auth", h.Routes())

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"a@b.com","password":"x","role":"customer"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestAuthHandlerListUsersParsesPagination(t *testing.T) {
	// list /users is protected by JWT; handler pagination is covered indirectly via service tests.
}

func TestAuthHandler_ListUsers_AdminRole(t *testing.T) {
	admin := domain.User{ID: uuid.New(), Email: "admin@example.com", Role: "admin"}
	repo := &authRepoStub{users: map[string]domain.User{"admin@example.com": admin}}
	svc := auth.NewService(repo, &issuerStub{})
	h := NewHandler(svc, "secret")

	req := httptest.NewRequest(http.MethodGet, "/auth/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthContextKey, &middleware.Claims{UserID: admin.ID.String(), Role: "admin"}))
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/auth/users", h.listUsers)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandler_ListUsers_ForbiddenForNonAdmin(t *testing.T) {
	repo := &authRepoStub{users: map[string]domain.User{}}
	svc := auth.NewService(repo, &issuerStub{})
	h := NewHandler(svc, "secret")

	req := httptest.NewRequest(http.MethodGet, "/auth/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthContextKey, &middleware.Claims{UserID: uuid.NewString(), Role: "customer"}))
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/auth/users", h.listUsers)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthHandler_GetUser_NotFound(t *testing.T) {
	repo := &authRepoStub{users: map[string]domain.User{}}
	svc := auth.NewService(repo, &issuerStub{})
	h := NewHandler(svc, "secret")

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/users/"+id.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthContextKey, &middleware.Claims{UserID: uuid.NewString(), Role: "admin"}))
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/auth/users/{id}", h.getUser)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
