package aiparser

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// WinterPeakRange represents the solar date range for a lunar winter peak period.
// Winter peak: La Yue 23 (Little New Year) through Zheng Yue 15 (Lantern Festival).
type WinterPeakRange struct {
	Year      int
	StartDate time.Time // Solar date of La Yue 23
	EndDate   time.Time // Solar date of Zheng Yue 15
}

// SummerPeakRange represents the fixed summer peak dates.
// Summer peak: July 15 - August 15 (fixed Gregorian).
type SummerPeakRange struct {
	Year      int
	StartDate time.Time
	EndDate   time.Time
}

// LunarService provides lunar calendar calculations for peak season detection.
type LunarService struct {
	winterPeaks map[int]WinterPeakRange
	summerPeaks map[int]SummerPeakRange
}

// NewLunarService pre-computes peak date ranges for the given year span.
func NewLunarService(startYear, endYear int) (*LunarService, error) {
	svc := &LunarService{
		winterPeaks: make(map[int]WinterPeakRange),
		summerPeaks: make(map[int]SummerPeakRange),
	}

	for year := startYear; year <= endYear; year++ {
		wp, err := computeWinterPeak(year)
		if err != nil {
			return nil, fmt.Errorf("winter peak %d: %w", year, err)
		}
		svc.winterPeaks[year] = wp

		svc.summerPeaks[year] = SummerPeakRange{
			Year:      year,
			StartDate: time.Date(year, 7, 15, 0, 0, 0, 0, time.Local),
			EndDate:   time.Date(year, 8, 15, 0, 0, 0, 0, time.Local),
		}
	}

	return svc, nil
}

// computeWinterPeak calculates the solar dates for La Yue 23 ~ Zheng Yue 15.
//
// The winter peak spans two lunar years:
// - La Yue 23 of lunar year Y (around Jan/Feb of solar year Y+1)
// - Zheng Yue 15 of lunar year Y+1 (around Feb/Mar of solar year Y+1)
//
// We index by the solar year that contains most of the peak period.
func computeWinterPeak(solarYear int) (WinterPeakRange, error) {
	// La Yue (month 12) day 23 of the lunar year before this solar year.
	// Lunar year N's La Yue falls in solar year N+1 (Jan/Feb).
	lunarYearForLaYue := solarYear - 1

	startSolar, err := lunarToSolar(lunarYearForLaYue, 12, 23)
	if err != nil {
		return WinterPeakRange{}, fmt.Errorf("la yue 23 of lunar %d: %w", lunarYearForLaYue, err)
	}

	// Zheng Yue (month 1) day 15 of the current lunar year.
	lunarYearForZhengYue := solarYear
	endSolar, err := lunarToSolar(lunarYearForZhengYue, 1, 15)
	if err != nil {
		return WinterPeakRange{}, fmt.Errorf("zheng yue 15 of lunar %d: %w", lunarYearForZhengYue, err)
	}

	return WinterPeakRange{
		Year:      solarYear,
		StartDate: startSolar,
		EndDate:   endSolar,
	}, nil
}

// lunarToSolar converts a lunar date to a solar date using 6tail/lunar-go.
func lunarToSolar(lunarYear, lunarMonth, lunarDay int) (time.Time, error) {
	defer func() {
		if r := recover(); r != nil {
			// lunar-go panics on invalid dates
		}
	}()

	lunar := calendar.NewLunarFromYmd(lunarYear, lunarMonth, lunarDay)
	solar := lunar.GetSolar()

	return time.Date(
		solar.GetYear(),
		time.Month(solar.GetMonth()),
		solar.GetDay(),
		0, 0, 0, 0, time.Local,
	), nil
}

// IsInPeakSeason checks if a given date falls within any peak season.
func (s *LunarService) IsInPeakSeason(date time.Time) (bool, string) {
	year := date.Year()
	d := time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	// Check summer peak
	if sp, ok := s.summerPeaks[year]; ok {
		if !d.Before(sp.StartDate) && !d.After(sp.EndDate) {
			return true, "summer"
		}
	}

	// Check winter peak for this year and next year (peaks span year boundary)
	for _, checkYear := range []int{year, year + 1} {
		if wp, ok := s.winterPeaks[checkYear]; ok {
			if !d.Before(wp.StartDate) && !d.After(wp.EndDate) {
				return true, "winter"
			}
		}
	}

	return false, ""
}

// GetWinterPeak returns the winter peak range for a given solar year.
func (s *LunarService) GetWinterPeak(solarYear int) (WinterPeakRange, bool) {
	wp, ok := s.winterPeaks[solarYear]
	return wp, ok
}

// GetSummerPeak returns the summer peak range for a given year.
func (s *LunarService) GetSummerPeak(year int) (SummerPeakRange, bool) {
	sp, ok := s.summerPeaks[year]
	return sp, ok
}

// AllWinterPeaks returns all pre-computed winter peak ranges.
func (s *LunarService) AllWinterPeaks() map[int]WinterPeakRange {
	return s.winterPeaks
}
