package aiparser

// TravelRequirement is the structured output from AI parsing.
type TravelRequirement struct {
	Destination string   `json:"destination"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	BudgetCents int64    `json:"budget_cents"`
	Adults      int      `json:"adults"`
	Children    int      `json:"children"`
	Preferences []string `json:"preferences"`
}
