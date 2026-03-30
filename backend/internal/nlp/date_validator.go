package nlp

import (
	"fmt"
	"time"
)

// DateValidation holds the result of date validation.
type DateValidation struct {
	IsValid      bool   `json:"is_valid"`
	IsPeakSeason bool   `json:"is_peak_season"`
	PeakType     string `json:"peak_type,omitempty"`
	Warning      string `json:"warning,omitempty"`
}

// DateValidator checks travel dates against peak season rules.
type DateValidator struct {
	lunar *LunarService
}

// NewDateValidator creates a date validator with pre-computed peak data.
func NewDateValidator(lunar *LunarService) *DateValidator {
	return &DateValidator{lunar: lunar}
}

// Validate checks if the travel date range is valid and whether it falls in a peak season.
func (v *DateValidator) Validate(startDate, endDate string) (*DateValidation, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	if !end.After(start) {
		return &DateValidation{
			IsValid: false,
			Warning: "End date must be after start date",
		}, nil
	}

	// Check each day of the trip for peak season
	isPeak := false
	peakType := ""
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if peak, pt := v.lunar.IsInPeakSeason(d); peak {
			isPeak = true
			peakType = pt
			break
		}
	}

	result := &DateValidation{
		IsValid:      true,
		IsPeakSeason: isPeak,
		PeakType:     peakType,
	}

	if isPeak {
		switch peakType {
		case "summer":
			result.Warning = "Summer peak season (Jul 15 - Aug 15) - prices may be higher"
		case "winter":
			result.Warning = "Winter/Spring Festival peak season - prices may be higher"
		}
	}

	return result, nil
}
