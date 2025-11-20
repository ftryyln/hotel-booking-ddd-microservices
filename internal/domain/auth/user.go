package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents auth entity.
type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
}

// TokenIssuer issues JWT tokens.
type TokenIssuer interface {
	Generate(ctx context.Context, user User) (access, refresh string, err error)
}
