package booking

import (
	"github.com/ftryyln/hotel-booking-microservices/pkg/domain"
)

// IsConfirmedSpec checks if booking is confirmed.
type IsConfirmedSpec struct{}

func (s IsConfirmedSpec) IsSatisfiedBy(b Booking) bool {
	return b.Status == StatusConfirmed
}

// IsHighValueSpec checks if booking is high value (> 10,000,000 IDR).
type IsHighValueSpec struct{}

func (s IsHighValueSpec) IsSatisfiedBy(b Booking) bool {
	return b.TotalPrice > 10000000
}

// IsLongStaySpec checks if booking is long stay (> 7 nights).
type IsLongStaySpec struct{}

func (s IsLongStaySpec) IsSatisfiedBy(b Booking) bool {
	return b.TotalNights > 7
}

// NewVIPBookingSpec combines high value OR long stay.
func NewVIPBookingSpec() domain.Specification[Booking] {
	highValue := IsHighValueSpec{}
	longStay := IsLongStaySpec{}
	return domain.Or[Booking](highValue, longStay)
}
