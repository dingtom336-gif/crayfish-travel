package bidding

import (
	"testing"
)

func TestRankTop5_SortsByPrice(t *testing.T) {
	quotes := []Quote{
		{PackageTitle: "Expensive", BasePriceCents: 100000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 110000},
		{PackageTitle: "Cheapest", BasePriceCents: 50000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 60000},
		{PackageTitle: "Mid", BasePriceCents: 70000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 80000},
		{PackageTitle: "High", BasePriceCents: 90000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 100000},
		{PackageTitle: "Low", BasePriceCents: 55000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 65000},
		{PackageTitle: "Extra", BasePriceCents: 120000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 130000},
	}

	ranked := RankTop5(quotes)

	if len(ranked) != 5 {
		t.Fatalf("expected 5 ranked quotes, got %d", len(ranked))
	}

	// First should be cheapest
	if ranked[0].PackageTitle != "Cheapest" {
		t.Errorf("rank 1 should be Cheapest, got %s", ranked[0].PackageTitle)
	}

	// Should be sorted ascending
	for i := 1; i < len(ranked); i++ {
		if ranked[i].TotalPriceCents < ranked[i-1].TotalPriceCents {
			t.Errorf("rank %d (%d) is cheaper than rank %d (%d)",
				i+1, ranked[i].TotalPriceCents, i, ranked[i-1].TotalPriceCents)
		}
	}

	// First should be best value
	if !ranked[0].IsBestValue {
		t.Error("rank 1 should be marked as best value")
	}
	for i := 1; i < len(ranked); i++ {
		if ranked[i].IsBestValue {
			t.Errorf("rank %d should not be best value", i+1)
		}
	}

	// Ranks should be 1-5
	for i, q := range ranked {
		if q.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, q.Rank)
		}
	}
}

func TestRankTop5_LessThan5(t *testing.T) {
	quotes := []Quote{
		{PackageTitle: "A", BasePriceCents: 50000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 60000},
		{PackageTitle: "B", BasePriceCents: 60000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 70000},
	}

	ranked := RankTop5(quotes)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 ranked quotes, got %d", len(ranked))
	}
}

func TestValidatePriceSplit_Valid(t *testing.T) {
	ranked := []RankedQuote{
		{Quote: Quote{BasePriceCents: 50000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 60000}},
		{Quote: Quote{BasePriceCents: 60000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 70000}},
	}

	if idx := ValidatePriceSplit(ranked); idx != -1 {
		t.Errorf("expected all valid, got failure at index %d", idx)
	}
}

func TestValidatePriceSplit_MismatchTotal(t *testing.T) {
	ranked := []RankedQuote{
		{Quote: Quote{BasePriceCents: 50000, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 99999}},
	}

	if idx := ValidatePriceSplit(ranked); idx != 0 {
		t.Errorf("expected failure at 0, got %d", idx)
	}
}

func TestValidatePriceSplit_ZeroBase(t *testing.T) {
	ranked := []RankedQuote{
		{Quote: Quote{BasePriceCents: 0, RefundGuaranteeFeeCents: 10000, TotalPriceCents: 10000}},
	}

	if idx := ValidatePriceSplit(ranked); idx != 0 {
		t.Errorf("expected failure at 0, got %d", idx)
	}
}

func TestMockSupplier_Returns8Quotes(t *testing.T) {
	mock := NewMockSupplier()
	quotes, err := mock.FetchQuotes("Sanya", 5, 800000, 2, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(quotes) != 8 {
		t.Errorf("expected 8 quotes, got %d", len(quotes))
	}

	for i, q := range quotes {
		if q.RefundGuaranteeFeeCents != 10000 {
			t.Errorf("quote %d: refund fee should be 10000, got %d", i, q.RefundGuaranteeFeeCents)
		}
		if q.TotalPriceCents != q.BasePriceCents+q.RefundGuaranteeFeeCents {
			t.Errorf("quote %d: price split mismatch", i)
		}
		if q.Supplier != "Fliggy" {
			t.Errorf("quote %d: expected Fliggy supplier", i)
		}
	}
}
