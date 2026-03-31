package bidding

import "testing"

// TestRankTop5_MultiFactorScoringWithRatingData verifies that RankTop5 correctly
// ranks quotes using multi-factor scoring: price (50%), rating (30%), reviews (20%).
func TestRankTop5_MultiFactorScoringWithRatingData(t *testing.T) {
	quotes := []Quote{
		{
			TotalPriceCents:         500000,
			BasePriceCents:          490000,
			RefundGuaranteeFeeCents: 10000,
			StarRating:              3.0,
			ReviewCount:             50,
			PackageTitle:            "Budget",
		},
		{
			TotalPriceCents:         800000,
			BasePriceCents:          790000,
			RefundGuaranteeFeeCents: 10000,
			StarRating:              5.0,
			ReviewCount:             300,
			PackageTitle:            "Premium",
		},
		{
			TotalPriceCents:         600000,
			BasePriceCents:          590000,
			RefundGuaranteeFeeCents: 10000,
			StarRating:              4.0,
			ReviewCount:             200,
			PackageTitle:            "MidRange",
		},
	}

	ranked := RankTop5(quotes)
	if len(ranked) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ranked))
	}

	// Budget: price=0.5*(1-500000/800000)=0.1875, rating=0.3*(3.0/5.0)=0.18, review=0.2*(50/300)=0.033 => 0.401
	// MidRange: price=0.5*(1-600000/800000)=0.125, rating=0.3*(4.0/5.0)=0.24, review=0.2*(200/300)=0.133 => 0.498
	// Premium: price=0.5*(1-800000/800000)=0.0, rating=0.3*(5.0/5.0)=0.30, review=0.2*(300/300)=0.20 => 0.500
	// Ranking: Premium > MidRange > Budget
	if ranked[0].PackageTitle != "Premium" {
		t.Errorf("rank 1 should be Premium (highest multi-factor score), got %s", ranked[0].PackageTitle)
	}
	if ranked[1].PackageTitle != "MidRange" {
		t.Errorf("rank 2 should be MidRange, got %s", ranked[1].PackageTitle)
	}
	if ranked[2].PackageTitle != "Budget" {
		t.Errorf("rank 3 should be Budget, got %s", ranked[2].PackageTitle)
	}

	// First ranked item should be marked as best value
	if !ranked[0].IsBestValue {
		t.Error("first ranked item should be marked as best value")
	}
	for i := 1; i < len(ranked); i++ {
		if ranked[i].IsBestValue {
			t.Errorf("rank %d should not be marked as best value", i+1)
		}
	}

	// Verify StarRating and ReviewCount are preserved through ranking
	if ranked[0].StarRating != 5.0 {
		t.Errorf("Premium star rating not preserved: got %f", ranked[0].StarRating)
	}
	if ranked[0].ReviewCount != 300 {
		t.Errorf("Premium review count not preserved: got %d", ranked[0].ReviewCount)
	}
}

// TestRankTop5_EmptyInput verifies nil/empty input returns nil.
func TestRankTop5_EmptyInput(t *testing.T) {
	ranked := RankTop5(nil)
	if len(ranked) != 0 {
		t.Errorf("expected empty result for nil input, got %d items", len(ranked))
	}

	ranked = RankTop5([]Quote{})
	if len(ranked) != 0 {
		t.Errorf("expected empty result for empty input, got %d items", len(ranked))
	}
}

// TestRankTop5_SingleQuote verifies single-element input.
func TestRankTop5_SingleQuote(t *testing.T) {
	quotes := []Quote{
		{
			TotalPriceCents:         700000,
			BasePriceCents:          690000,
			RefundGuaranteeFeeCents: 10000,
			PackageTitle:            "Solo",
			StarRating:              4.5,
			ReviewCount:             100,
		},
	}

	ranked := RankTop5(quotes)
	if len(ranked) != 1 {
		t.Fatalf("expected 1 result, got %d", len(ranked))
	}
	if ranked[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", ranked[0].Rank)
	}
	if !ranked[0].IsBestValue {
		t.Error("single quote should be marked as best value")
	}
}

// TestRankTop5_ExactlyFive verifies the cap at 5 results.
func TestRankTop5_ExactlyFive(t *testing.T) {
	quotes := make([]Quote, 10)
	for i := range quotes {
		price := int64((i + 1) * 100000)
		quotes[i] = Quote{
			TotalPriceCents:         price,
			BasePriceCents:          price - 10000,
			RefundGuaranteeFeeCents: 10000,
			StarRating:              4.0,
			ReviewCount:             100,
		}
	}

	ranked := RankTop5(quotes)
	if len(ranked) != 5 {
		t.Fatalf("expected 5 results from 10 quotes, got %d", len(ranked))
	}

	// With equal rating and reviews, cheaper quotes score higher (price is 50% weight).
	// The 5 cheapest should be selected, ordered by price ascending (cheapest = highest score).
	for i, q := range ranked {
		expectedPrice := int64((i + 1) * 100000)
		if q.TotalPriceCents != expectedPrice {
			t.Errorf("rank %d: expected price %d, got %d", i+1, expectedPrice, q.TotalPriceCents)
		}
	}
}

// TestRankTop5_EqualPrices verifies behavior with ties.
func TestRankTop5_EqualPrices(t *testing.T) {
	quotes := []Quote{
		{TotalPriceCents: 500000, BasePriceCents: 490000, RefundGuaranteeFeeCents: 10000, PackageTitle: "A", StarRating: 5.0, ReviewCount: 100},
		{TotalPriceCents: 500000, BasePriceCents: 490000, RefundGuaranteeFeeCents: 10000, PackageTitle: "B", StarRating: 3.0, ReviewCount: 100},
	}

	ranked := RankTop5(quotes)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 results, got %d", len(ranked))
	}

	// With equal prices and reviews, higher rating wins
	if ranked[0].PackageTitle != "A" {
		t.Errorf("rank 1 should be A (higher rating), got %s", ranked[0].PackageTitle)
	}

	// First should be best value
	if !ranked[0].IsBestValue {
		t.Error("first ranked should be best value even with equal prices")
	}
	if ranked[1].IsBestValue {
		t.Error("second ranked should not be best value")
	}
}

// TestValidatePriceSplit_NegativeRefundFee verifies negative refund fee is caught.
func TestValidatePriceSplit_NegativeRefundFee(t *testing.T) {
	ranked := []RankedQuote{
		{Quote: Quote{BasePriceCents: 50000, RefundGuaranteeFeeCents: -100, TotalPriceCents: 49900}},
	}
	if idx := ValidatePriceSplit(ranked); idx != 0 {
		t.Errorf("expected failure at 0 for negative refund fee, got %d", idx)
	}
}

// TestValidatePriceSplit_EmptySlice verifies empty input passes validation.
func TestValidatePriceSplit_EmptySlice(t *testing.T) {
	if idx := ValidatePriceSplit(nil); idx != -1 {
		t.Errorf("expected -1 for nil input, got %d", idx)
	}
	if idx := ValidatePriceSplit([]RankedQuote{}); idx != -1 {
		t.Errorf("expected -1 for empty input, got %d", idx)
	}
}
