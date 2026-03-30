package bidding

import (
	"sort"
)

// RankedQuote extends Quote with ranking metadata.
type RankedQuote struct {
	Quote
	ID          string `json:"id,omitempty"`
	Rank        int    `json:"rank"`
	IsBestValue bool   `json:"is_best_value"`
}

// RankTop5 sorts quotes by total price ascending and returns the top 5.
// The cheapest quote is marked as best value.
// Every quote MUST have the price split: base_price + refund_guarantee_fee (compliance).
func RankTop5(quotes []Quote) []RankedQuote {
	// Sort by total price ascending
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].TotalPriceCents < quotes[j].TotalPriceCents
	})

	// Take top 5
	limit := 5
	if len(quotes) < limit {
		limit = len(quotes)
	}

	ranked := make([]RankedQuote, limit)
	for i := 0; i < limit; i++ {
		ranked[i] = RankedQuote{
			Quote:       quotes[i],
			Rank:        i + 1,
			IsBestValue: i == 0,
		}
	}

	return ranked
}

// ValidatePriceSplit ensures every quote has the mandatory price breakdown.
// Returns the first quote index that fails validation, or -1 if all pass.
func ValidatePriceSplit(quotes []RankedQuote) int {
	for i, q := range quotes {
		if q.BasePriceCents <= 0 {
			return i
		}
		if q.RefundGuaranteeFeeCents <= 0 {
			return i
		}
		if q.TotalPriceCents != q.BasePriceCents+q.RefundGuaranteeFeeCents {
			return i
		}
	}
	return -1
}
