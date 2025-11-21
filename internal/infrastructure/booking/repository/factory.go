package repository

import (
	"errors"

	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
)

// RepositoryType defines the type of repository to create.
type RepositoryType string

const (
	// TypeGorm is the GORM implementation of the repository.
	TypeGorm RepositoryType = "gorm"
	// TypeMemory is an in-memory implementation (for testing/dev).
	TypeMemory RepositoryType = "memory"
)

// Factory creates repositories based on the requested type.
type Factory interface {
	CreateBookingRepository(typ RepositoryType) (domain.Repository, error)
}

// GormFactory is a factory that creates GORM-based repositories.
// It can be extended to support other types if needed.
type GormFactory struct {
	db *gorm.DB
}

// NewGormFactory creates a new GormFactory.
func NewGormFactory(db *gorm.DB) *GormFactory {
	return &GormFactory{db: db}
}

// CreateBookingRepository creates a booking repository.
func (f *GormFactory) CreateBookingRepository(typ RepositoryType) (domain.Repository, error) {
	switch typ {
	case TypeGorm:
		return NewGormRepository(f.db), nil
	default:
		return nil, errors.New("unsupported repository type")
	}
}
