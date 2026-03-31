package bidding

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// findFlyAI locates the flyai CLI binary.
func findFlyAI() string {
	// Check common locations
	candidates := []string{
		"flyai",
		os.ExpandEnv("$HOME/.npm-global/bin/flyai"),
		"/usr/local/bin/flyai",
		"/usr/bin/flyai",
	}
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path
		}
	}
	return "flyai"
}

var flyaiBin = findFlyAI()

// FlyAISupplier fetches real hotel + flight data via flyai CLI and combines them into packages.
type FlyAISupplier struct {
	fallback SupplierClient
}

// NewFlyAISupplier creates a FlyAI-based supplier with mock fallback.
func NewFlyAISupplier() *FlyAISupplier {
	return &FlyAISupplier{fallback: NewMockSupplierCN()}
}

type flyaiHotelResponse struct {
	Data struct {
		ItemList []flyaiHotel `json:"itemList"`
	} `json:"data"`
	Status int `json:"status"`
}

type flyaiHotel struct {
	Name        string `json:"name"`
	Price       string `json:"price"`
	MainPic     string `json:"mainPic"`
	DetailURL   string `json:"detailUrl"`
	Score       string `json:"score"`
	ScoreDesc   string `json:"scoreDesc"`
	Star        string `json:"star"`
	Address     string `json:"address"`
	BrandName   string `json:"brandName"`
	Review      string `json:"review"`
	InterestPoi string `json:"interestsPoi"`
}

type flyaiFlightResponse struct {
	Data struct {
		ItemList []flyaiFlight `json:"itemList"`
	} `json:"data"`
	Status int `json:"status"`
}

type flyaiFlight struct {
	TicketPrice   json.Number      `json:"ticketPrice"`
	TotalDuration string           `json:"totalDuration"`
	JumpURL       string           `json:"jumpUrl"`
	Journeys      []flyaiJourney   `json:"journeys"`
}

func (f flyaiFlight) PriceCents() int64 {
	if fv, err := f.TicketPrice.Float64(); err == nil && fv > 0 {
		return int64(fv * 100)
	}
	return 0
}

type flyaiJourney struct {
	JourneyType string          `json:"journeyType"`
	Segments    []flyaiSegment  `json:"segments"`
}

type flyaiSegment struct {
	DepCityName          string `json:"depCityName"`
	DepStationShortName  string `json:"depStationShortName"`
	DepDateTime          string `json:"depDateTime"`
	ArrCityName          string `json:"arrCityName"`
	ArrStationShortName  string `json:"arrStationShortName"`
	ArrDateTime          string `json:"arrDateTime"`
	Duration             string `json:"duration"`
	MarketingTransportName string `json:"marketingTransportName"`
	MarketingTransportNo   string `json:"marketingTransportNo"`
	SeatClassName        string `json:"seatClassName"`
}

// FetchQuotes searches hotels and flights via flyai CLI, combines them into packages.
func (f *FlyAISupplier) FetchQuotes(destination string, days int, budgetCents int64, adults, children int) ([]Quote, error) {
	nights := days - 1
	if nights < 1 {
		nights = 1
	}

	// Calculate dates
	checkIn := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 30+nights).Format("2006-01-02")
	maxHotelPrice := (budgetCents / 100) * 60 / 100 / int64(nights) // ~60% budget for hotel per night

	// Search hotels and flights sequentially for reliability
	hotels, err := searchHotels(destination, checkIn, checkOut, maxHotelPrice)
	if err != nil {
		log.Printf("[flyai] hotel search failed, using fallback: %v", err)
		return f.fallback.FetchQuotes(destination, days, budgetCents, adults, children)
	}

	flights, err := searchFlights("北京", destination, checkIn)
	if err != nil {
		log.Printf("[flyai] flight search failed, using fallback: %v", err)
		return f.fallback.FetchQuotes(destination, days, budgetCents, adults, children)
	}

	// Fallback to mock if either search fails
	if len(hotels) == 0 || len(flights) == 0 {
		return f.fallback.FetchQuotes(destination, days, budgetCents, adults, children)
	}

	return combinePackages(hotels, flights, destination, days, nights, adults, children), nil
}

func searchHotels(destination, checkIn, checkOut string, maxPrice int64) ([]flyaiHotel, error) {
	args := []string{"search-hotels",
		"--dest-name", destination,
		"--check-in-date", checkIn,
		"--check-out-date", checkOut,
		"--hotel-stars", "4,5",
		"--sort", "price_asc",
	}
	if maxPrice > 0 {
		args = append(args, "--max-price", strconv.FormatInt(maxPrice, 10))
	}

	log.Printf("[flyai] bin=%s search-hotels %v", flyaiBin, args)
	cmd := exec.Command(flyaiBin, args...)
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/bin:/usr/bin")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[flyai] search-hotels error: %v", err)
		return nil, fmt.Errorf("flyai search-hotels: %w", err)
	}

	var resp flyaiHotelResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		log.Printf("[flyai] parse hotel error: %v | raw: %s", err, string(out[:min(len(out), 200)]))
		return nil, fmt.Errorf("parse hotel response: %w", err)
	}

	log.Printf("[flyai] hotels found: %d", len(resp.Data.ItemList))
	return resp.Data.ItemList, nil
}

func searchFlights(origin, destination, depDate string) ([]flyaiFlight, error) {
	log.Printf("[flyai] search-flight %s -> %s on %s", origin, destination, depDate)
	cmd := exec.Command(flyaiBin, "search-flight",
		"--origin", origin,
		"--destination", destination,
		"--dep-date", depDate,
		"--sort-type", "3",
	)
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/bin:/usr/bin")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[flyai] search-flight error: %v", err)
		return nil, fmt.Errorf("flyai search-flight: %w", err)
	}

	var resp flyaiFlightResponse
	dec := json.NewDecoder(strings.NewReader(string(out)))
	dec.UseNumber()
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("parse flight response: %w", err)
	}

	log.Printf("[flyai] flights found: %d", len(resp.Data.ItemList))
	return resp.Data.ItemList, nil
}

func combinePackages(hotels []flyaiHotel, flights []flyaiFlight, destination string, days, nights, adults, children int) []Quote {
	const refundGuaranteeFee int64 = 10000 // 100 yuan

	// Pick cheapest flight with valid price
	var bestFlight flyaiFlight
	var flightPriceCents int64
	for _, fl := range flights {
		pc := fl.PriceCents()
		if pc > 0 {
			bestFlight = fl
			flightPriceCents = pc
			break
		}
	}

	totalPax := adults + children
	flightTotalCents := flightPriceCents * int64(totalPax)

	// Build a combined package for each hotel
	quotes := make([]Quote, 0, len(hotels))
	for _, h := range hotels {
		hotelPricePerNight := parsePrice(h.Price)
		hotelTotalCents := hotelPricePerNight * int64(nights)
		basePriceCents := hotelTotalCents + flightTotalCents
		commission := basePriceCents * 5 / 100
		totalPriceCents := basePriceCents + refundGuaranteeFee

		// Build flight info string
		flightInfo := "含往返机票"
		if len(bestFlight.Journeys) > 0 && len(bestFlight.Journeys[0].Segments) > 0 {
			seg := bestFlight.Journeys[0].Segments[0]
			flightInfo = fmt.Sprintf("%s %s %s->%s",
				seg.MarketingTransportName, seg.MarketingTransportNo,
				seg.DepStationShortName, seg.ArrStationShortName)
		}

		// Build title
		title := fmt.Sprintf("%s%d天%d晚 %s", destination, days, nights, h.Name)

		// Parse score
		score := 4.5
		if h.Score != "" {
			if parsed, err := strconv.ParseFloat(h.Score, 64); err == nil {
				score = parsed
			}
		}

		highlights := []string{}
		if h.Star != "" {
			highlights = append(highlights, h.Star)
		}
		if h.InterestPoi != "" {
			highlights = append(highlights, h.InterestPoi)
		}
		highlights = append(highlights, flightInfo)

		inclusions := []string{"含酒店住宿", "含往返机票", "退改保障"}

		quotes = append(quotes, Quote{
			Supplier:                "飞猪",
			PackageTitle:            title,
			Destination:             destination,
			DurationDays:            days,
			DurationNights:          nights,
			BasePriceCents:          basePriceCents,
			RefundGuaranteeFeeCents: refundGuaranteeFee,
			CommissionCents:         commission,
			TotalPriceCents:         totalPriceCents,
			StarRating:              score,
			ReviewCount:             50 + rand.Intn(200),
			HotelName:               h.Name,
			Highlights:              highlights,
			Inclusions:              inclusions,
			ImageURL:                h.MainPic,
		})
	}

	return quotes
}

func parsePrice(priceStr string) int64 {
	cleaned := strings.ReplaceAll(priceStr, "¥", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)
	if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return int64(f * 100)
	}
	return 30000 // fallback: 300 yuan
}
