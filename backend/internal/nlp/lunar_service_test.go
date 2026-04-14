package nlp

import (
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// TestNewLunarService_10YearRange verifies that the lunar service can pre-compute
// peak dates for 2026-2035 without errors.
func TestNewLunarService_10YearRange(t *testing.T) {
	svc, err := NewLunarService(2026, 2035)
	if err != nil {
		t.Fatalf("failed to create lunar service: %v", err)
	}

	for year := 2026; year <= 2035; year++ {
		wp, ok := svc.GetWinterPeak(year)
		if !ok {
			t.Errorf("missing winter peak for %d", year)
			continue
		}

		// Winter peak start must be before end
		if !wp.StartDate.Before(wp.EndDate) {
			t.Errorf("year %d: winter peak start %v is not before end %v",
				year, wp.StartDate.Format("2006-01-02"), wp.EndDate.Format("2006-01-02"))
		}

		// Start should be in January or February (La Yue 23 -> solar)
		if wp.StartDate.Month() < 1 || wp.StartDate.Month() > 2 {
			t.Errorf("year %d: unexpected start month %d (expected Jan or Feb)",
				year, wp.StartDate.Month())
		}

		// End should be in February or March (Zheng Yue 15 -> solar)
		if wp.EndDate.Month() < 2 || wp.EndDate.Month() > 3 {
			t.Errorf("year %d: unexpected end month %d (expected Feb or Mar)",
				year, wp.EndDate.Month())
		}

		// Peak duration should be 23-25 days (La Yue 23 to Zheng Yue 15)
		duration := wp.EndDate.Sub(wp.StartDate).Hours() / 24
		if duration < 20 || duration > 30 {
			t.Errorf("year %d: unusual peak duration %.0f days (expected 20-30)",
				year, duration)
		}

		t.Logf("Year %d: winter peak %s ~ %s (%.0f days)",
			year,
			wp.StartDate.Format("2006-01-02"),
			wp.EndDate.Format("2006-01-02"),
			duration,
		)
	}
}

// TestWinterPeakDates_KnownValues verifies specific known lunar-solar conversions.
func TestWinterPeakDates_KnownValues(t *testing.T) {
	// Known: Lunar 2025-12-23 = Solar 2026-02-10 (La Yue 23, Year of the Snake)
	// Known: Lunar 2026-01-15 = Solar 2026-03-03 (Zheng Yue 15, Lantern Festival)
	svc, err := NewLunarService(2026, 2026)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	wp, ok := svc.GetWinterPeak(2026)
	if !ok {
		t.Fatal("missing 2026 winter peak")
	}

	// Cross-validate with direct lunar-go call
	laYue23 := calendar.NewLunarFromYmd(2025, 12, 23)
	expectedStart := laYue23.GetSolar()
	if wp.StartDate.Day() != expectedStart.GetDay() ||
		int(wp.StartDate.Month()) != expectedStart.GetMonth() ||
		wp.StartDate.Year() != expectedStart.GetYear() {
		t.Errorf("2026 start mismatch: got %s, expected %d-%02d-%02d",
			wp.StartDate.Format("2006-01-02"),
			expectedStart.GetYear(), expectedStart.GetMonth(), expectedStart.GetDay())
	}

	zhengYue15 := calendar.NewLunarFromYmd(2026, 1, 15)
	expectedEnd := zhengYue15.GetSolar()
	if wp.EndDate.Day() != expectedEnd.GetDay() ||
		int(wp.EndDate.Month()) != expectedEnd.GetMonth() ||
		wp.EndDate.Year() != expectedEnd.GetYear() {
		t.Errorf("2026 end mismatch: got %s, expected %d-%02d-%02d",
			wp.EndDate.Format("2006-01-02"),
			expectedEnd.GetYear(), expectedEnd.GetMonth(), expectedEnd.GetDay())
	}

	t.Logf("2026 winter peak verified: %s ~ %s",
		wp.StartDate.Format("2006-01-02"),
		wp.EndDate.Format("2006-01-02"))
}

// TestSummerPeak verifies summer peak is fixed Gregorian range.
func TestSummerPeak(t *testing.T) {
	svc, err := NewLunarService(2026, 2035)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	for year := 2026; year <= 2035; year++ {
		sp, ok := svc.GetSummerPeak(year)
		if !ok {
			t.Errorf("missing summer peak for %d", year)
			continue
		}
		expectedStart := time.Date(year, 7, 15, 0, 0, 0, 0, time.Local)
		expectedEnd := time.Date(year, 8, 15, 0, 0, 0, 0, time.Local)
		if !sp.StartDate.Equal(expectedStart) || !sp.EndDate.Equal(expectedEnd) {
			t.Errorf("year %d: summer peak mismatch", year)
		}
	}
}

// TestIsInPeakSeason_Summer tests summer peak detection.
func TestIsInPeakSeason_Summer(t *testing.T) {
	svc, err := NewLunarService(2026, 2026)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		date   time.Time
		isPeak bool
		season string
	}{
		{time.Date(2026, 7, 14, 0, 0, 0, 0, time.Local), false, ""},
		{time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local), true, "summer"},
		{time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local), true, "summer"},
		{time.Date(2026, 8, 15, 0, 0, 0, 0, time.Local), true, "summer"},
		{time.Date(2026, 8, 16, 0, 0, 0, 0, time.Local), false, ""},
	}

	for _, tt := range tests {
		isPeak, season := svc.IsInPeakSeason(tt.date)
		if isPeak != tt.isPeak || season != tt.season {
			t.Errorf("date %s: got (%v, %q), want (%v, %q)",
				tt.date.Format("2006-01-02"), isPeak, season, tt.isPeak, tt.season)
		}
	}
}

// TestLunarCrossValidation performs cross-validation of lunar-go by independently
// verifying that the La Yue 23 -> Solar conversion is consistent across all 10 years.
// This is the test invoked by scripts/lunar-verify.sh.
func TestLunarCrossValidation(t *testing.T) {
	for lunarYear := 2025; lunarYear <= 2034; lunarYear++ {
		solarYear := lunarYear + 1

		// Method 1: Direct lunar-go conversion
		laYue23 := calendar.NewLunarFromYmd(lunarYear, 12, 23)
		solar1 := laYue23.GetSolar()
		date1 := time.Date(solar1.GetYear(), time.Month(solar1.GetMonth()), solar1.GetDay(), 0, 0, 0, 0, time.Local)

		// Method 2: Our service computation
		svc, err := NewLunarService(solarYear, solarYear)
		if err != nil {
			t.Fatalf("year %d: %v", solarYear, err)
		}
		wp, _ := svc.GetWinterPeak(solarYear)

		if !date1.Equal(wp.StartDate) {
			t.Errorf("cross-validation failed for solar year %d: direct=%s service=%s",
				solarYear,
				date1.Format("2006-01-02"),
				wp.StartDate.Format("2006-01-02"))
		} else {
			t.Logf("solar year %d: La Yue 23 = %s (verified)", solarYear, date1.Format("2006-01-02"))
		}

		// Also verify Zheng Yue 15
		zhengYue15 := calendar.NewLunarFromYmd(solarYear, 1, 15)
		solar2 := zhengYue15.GetSolar()
		date2 := time.Date(solar2.GetYear(), time.Month(solar2.GetMonth()), solar2.GetDay(), 0, 0, 0, 0, time.Local)

		if !date2.Equal(wp.EndDate) {
			t.Errorf("cross-validation failed for solar year %d end: direct=%s service=%s",
				solarYear,
				date2.Format("2006-01-02"),
				wp.EndDate.Format("2006-01-02"))
		}
	}
}
