package bidding

import (
	"fmt"
	"math/rand"
)

// Quote represents a supplier's travel package quote.
type Quote struct {
	Supplier                string   `json:"supplier"`
	PackageTitle            string   `json:"package_title"`
	Destination             string   `json:"destination"`
	DurationDays            int      `json:"duration_days"`
	DurationNights          int      `json:"duration_nights"`
	BasePriceCents          int64    `json:"base_price_cents"`
	RefundGuaranteeFeeCents int64    `json:"refund_guarantee_fee_cents"`
	CommissionCents         int64    `json:"commission_cents"`
	TotalPriceCents         int64    `json:"total_price_cents"`
	StarRating              float64  `json:"star_rating"`
	ReviewCount             int      `json:"review_count"`
	HotelName               string   `json:"hotel_name"`
	Highlights              []string `json:"highlights"`
	Inclusions              []string `json:"inclusions"`
	ImageURL                string   `json:"image_url"`
}

// SupplierClient is the interface for fetching travel package quotes.
type SupplierClient interface {
	FetchQuotes(destination string, days int, budgetCents int64, adults, children int) ([]Quote, error)
}

// MockSupplier provides preset quotes for development.
type MockSupplier struct{}

// NewMockSupplier creates a mock supplier client.
func NewMockSupplier() *MockSupplier {
	return &MockSupplier{}
}

var mockPackages = []struct {
	titleTemplate string
	hotel         string
	highlights    []string
	inclusions    []string
	baseStarRating float64
}{
	{
		titleTemplate: "%s %dD%dN Beach Paradise",
		hotel:         "Hilton Resort & Spa",
		highlights:    []string{"Ocean view room", "Airport transfer", "Private butler"},
		inclusions:    []string{"Breakfast", "Spa voucher", "Kids club"},
		baseStarRating: 4.8,
	},
	{
		titleTemplate: "%s %dD%dN Family Fun",
		hotel:         "Marriott Family Resort",
		highlights:    []string{"Water park access", "Family suite", "Guided tour"},
		inclusions:    []string{"All meals", "Theme park tickets", "Photography"},
		baseStarRating: 4.6,
	},
	{
		titleTemplate: "%s %dD%dN Luxury Escape",
		hotel:         "Ritz-Carlton",
		highlights:    []string{"Presidential suite", "Private pool", "Fine dining"},
		inclusions:    []string{"All inclusive", "Limousine service", "Personal chef"},
		baseStarRating: 4.9,
	},
	{
		titleTemplate: "%s %dD%dN Budget Smart",
		hotel:         "Holiday Inn Express",
		highlights:    []string{"Central location", "Free WiFi", "Shuttle bus"},
		inclusions:    []string{"Breakfast", "City map", "Welcome drink"},
		baseStarRating: 4.2,
	},
	{
		titleTemplate: "%s %dD%dN Cultural Tour",
		hotel:         "Hyatt Regency",
		highlights:    []string{"Heritage tour", "Local cuisine", "Temple visit"},
		inclusions:    []string{"Half board", "Guide service", "Souvenir kit"},
		baseStarRating: 4.5,
	},
	{
		titleTemplate: "%s %dD%dN Adventure Trip",
		hotel:         "Sheraton Adventure Lodge",
		highlights:    []string{"Snorkeling", "Hiking trail", "Sunset cruise"},
		inclusions:    []string{"Equipment rental", "Lunch pack", "Refund guarantee included"},
		baseStarRating: 4.4,
	},
	{
		titleTemplate: "%s %dD%dN Romantic Getaway",
		hotel:         "Four Seasons",
		highlights:    []string{"Couples spa", "Candlelight dinner", "Beach villa"},
		inclusions:    []string{"Full board", "Champagne", "Photo session"},
		baseStarRating: 4.7,
	},
	{
		titleTemplate: "%s %dD%dN Value Pack",
		hotel:         "Courtyard by Marriott",
		highlights:    []string{"Pool access", "Gym", "Business center"},
		inclusions:    []string{"Breakfast", "Late checkout", "Parking"},
		baseStarRating: 4.3,
	},
}

// FetchQuotes returns 8 randomized mock quotes.
func (m *MockSupplier) FetchQuotes(destination string, days int, budgetCents int64, _, _ int) ([]Quote, error) {
	nights := days - 1
	if nights < 1 {
		nights = 1
	}

	const refundGuaranteeFee int64 = 10000 // 100 yuan in cents

	// Default budget when user doesn't specify one (500 yuan/day as baseline)
	effectiveBudget := budgetCents
	if effectiveBudget <= 0 {
		effectiveBudget = int64(days) * 50000 // 500 yuan per day in cents
	}

	quotes := make([]Quote, len(mockPackages))
	for i, pkg := range mockPackages {
		// Price varies around budget: 80%-120% of budget
		priceMultiplier := 0.8 + rand.Float64()*0.4
		basePrice := int64(float64(effectiveBudget) * priceMultiplier)
		commission := basePrice * 5 / 100
		totalPrice := basePrice + refundGuaranteeFee

		quotes[i] = Quote{
			Supplier:                "Fliggy",
			PackageTitle:            fmt.Sprintf(pkg.titleTemplate, destination, days, nights),
			Destination:             destination,
			DurationDays:            days,
			DurationNights:          nights,
			BasePriceCents:          basePrice,
			RefundGuaranteeFeeCents: refundGuaranteeFee,
			CommissionCents:         commission,
			TotalPriceCents:         totalPrice,
			StarRating:              pkg.baseStarRating,
			ReviewCount:             50 + rand.Intn(200),
			HotelName:               pkg.hotel,
			Highlights:              pkg.highlights,
			Inclusions:              pkg.inclusions,
			ImageURL:                fmt.Sprintf("https://images.crayfish.travel/%s-%d.jpg", destination, i+1),
		}
	}

	return quotes, nil
}
