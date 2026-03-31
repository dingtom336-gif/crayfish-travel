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

// RankTop5 scores quotes using multi-factor ranking (price 50%, rating 30%, reviews 20%)
// and returns the top 5. The highest-scored quote is marked as best value.
// Every quote MUST have the price split: base_price + refund_guarantee_fee (compliance).
func RankTop5(quotes []Quote) []RankedQuote {
	if len(quotes) == 0 {
		return nil
	}

	// Find max values for normalization
	var maxPrice int64
	var maxReviews int
	for _, q := range quotes {
		if q.TotalPriceCents > maxPrice {
			maxPrice = q.TotalPriceCents
		}
		if q.ReviewCount > maxReviews {
			maxReviews = q.ReviewCount
		}
	}
	if maxPrice == 0 {
		maxPrice = 1
	}
	if maxReviews == 0 {
		maxReviews = 1
	}

	// Score each quote
	type scoredQuote struct {
		quote Quote
		score float64
	}
	scored := make([]scoredQuote, len(quotes))
	for i, q := range quotes {
		priceScore := 1.0 - float64(q.TotalPriceCents)/float64(maxPrice)
		ratingScore := float64(q.StarRating) / 5.0
		reviewScore := float64(q.ReviewCount) / float64(maxReviews)
		if reviewScore > 1.0 {
			reviewScore = 1.0
		}
		scored[i] = scoredQuote{
			quote: q,
			score: 0.5*priceScore + 0.3*ratingScore + 0.2*reviewScore,
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Take top 5
	limit := 5
	if len(scored) < limit {
		limit = len(scored)
	}

	result := make([]RankedQuote, limit)
	for i := 0; i < limit; i++ {
		result[i] = RankedQuote{
			Quote:       scored[i].quote,
			Rank:        i + 1,
			IsBestValue: i == 0,
		}
	}
	return result
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
