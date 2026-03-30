package nlp

import (
	"testing"
)

func TestDateValidator_SummerPeak(t *testing.T) {
	lunar, err := NewLunarService(2026, 2026)
	if err != nil {
		t.Fatal(err)
	}
	v := NewDateValidator(lunar)

	tests := []struct {
		start, end string
		isPeak     bool
		peakType   string
	}{
		{"2026-07-01", "2026-07-10", false, ""},
		{"2026-07-15", "2026-07-20", true, "summer"},
		{"2026-08-10", "2026-08-20", true, "summer"},
		{"2026-09-01", "2026-09-05", false, ""},
	}

	for _, tt := range tests {
		result, err := v.Validate(tt.start, tt.end)
		if err != nil {
			t.Errorf("validate %s~%s: %v", tt.start, tt.end, err)
			continue
		}
		if result.IsPeakSeason != tt.isPeak {
			t.Errorf("%s~%s: isPeak got %v, want %v", tt.start, tt.end, result.IsPeakSeason, tt.isPeak)
		}
		if result.PeakType != tt.peakType {
			t.Errorf("%s~%s: peakType got %q, want %q", tt.start, tt.end, result.PeakType, tt.peakType)
		}
	}
}

func TestDateValidator_InvalidDates(t *testing.T) {
	lunar, _ := NewLunarService(2026, 2026)
	v := NewDateValidator(lunar)

	_, err := v.Validate("bad-date", "2026-07-20")
	if err == nil {
		t.Error("expected error for invalid start date")
	}

	result, err := v.Validate("2026-07-20", "2026-07-15")
	if err != nil {
		t.Fatal(err)
	}
	if result.IsValid {
		t.Error("expected invalid for end before start")
	}
}

func TestDateValidator_PeakWarning(t *testing.T) {
	lunar, _ := NewLunarService(2026, 2026)
	v := NewDateValidator(lunar)

	result, _ := v.Validate("2026-07-16", "2026-07-20")
	if result.Warning == "" {
		t.Error("expected warning for summer peak")
	}
}
