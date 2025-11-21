package booking

// PricingService handles pricing calculations (pure domain logic).
type PricingService struct{}

// NewPricingService creates a new PricingService.
func NewPricingService() *PricingService {
	return &PricingService{}
}

// CalculateTotalPrice calculates the base total price including extra guest surcharges.
func (s *PricingService) CalculateTotalPrice(basePrice float64, nights int, guests int) float64 {
	total := basePrice * float64(nights)

	// Extra guest surcharge (domain rule: > 2 guests pay 20% extra per night)
	if guests > 2 {
		extraGuests := guests - 2
		total += float64(extraGuests) * basePrice * 0.2 * float64(nights)
	}

	return total
}

// ApplyDiscount applies discounts based on stay duration.
func (s *PricingService) ApplyDiscount(totalPrice float64, nights int) float64 {
	// Long stay discount (domain rule)
	if nights >= 7 {
		return totalPrice * 0.9 // 10% discount
	}
	if nights >= 3 {
		return totalPrice * 0.95 // 5% discount
	}
	return totalPrice
}
