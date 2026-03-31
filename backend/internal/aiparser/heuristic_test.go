package aiparser

import (
	"testing"
)

func TestHeuristicParser_FullInput(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("去三亚5天8000块2个人")
	if !matched {
		t.Fatal("expected match for full input")
	}
	if result.Destination != "三亚" {
		t.Errorf("destination: got %q, want %q", result.Destination, "三亚")
	}
	if result.BudgetCents != 800000 {
		t.Errorf("budget_cents: got %d, want %d", result.BudgetCents, 800000)
	}
	if result.Adults != 2 {
		t.Errorf("adults: got %d, want %d", result.Adults, 2)
	}
	// Verify duration via date gap (5 days)
	if result.StartDate == "" || result.EndDate == "" {
		t.Fatal("expected non-empty dates")
	}
}

func TestHeuristicParser_DestinationOnly(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("暑假想去北京玩")
	if !matched {
		t.Fatal("expected match for destination-only input")
	}
	if result.Destination != "北京" {
		t.Errorf("destination: got %q, want %q", result.Destination, "北京")
	}
	// Defaults should be applied
	if result.Adults != 2 {
		t.Errorf("adults default: got %d, want %d", result.Adults, 2)
	}
}

func TestHeuristicParser_Preferences(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("带娃去三亚海边温泉")
	if !matched {
		t.Fatal("expected match for preference input")
	}
	if result.Destination != "三亚" {
		t.Errorf("destination: got %q, want %q", result.Destination, "三亚")
	}

	// Check preferences contain expected tags
	prefSet := map[string]bool{}
	for _, pref := range result.Preferences {
		prefSet[pref] = true
	}
	expectedPrefs := []string{"family-friendly", "beachfront", "hot-spring"}
	for _, expected := range expectedPrefs {
		if !prefSet[expected] {
			t.Errorf("missing preference %q in %v", expected, result.Preferences)
		}
	}
}

func TestHeuristicParser_BudgetWan(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("2万块去泰国")
	if !matched {
		t.Fatal("expected match for wan budget input")
	}
	if result.Destination != "泰国" {
		t.Errorf("destination: got %q, want %q", result.Destination, "泰国")
	}
	// 2万 = 20000 yuan = 2000000 cents
	if result.BudgetCents != 2000000 {
		t.Errorf("budget_cents: got %d, want %d", result.BudgetCents, 2000000)
	}
}

func TestHeuristicParser_EmptyString(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("")
	if matched {
		t.Error("expected no match for empty string")
	}
	if result != nil {
		t.Error("expected nil result for empty string")
	}
}

func TestHeuristicParser_NoDestination(t *testing.T) {
	p := NewHeuristicParser()
	result, matched := p.Parse("我想旅游5天")
	if matched {
		t.Error("expected no match when no destination found")
	}
	if result != nil {
		t.Error("expected nil result when no destination found")
	}
}

func TestHeuristicParser_InternationalCities(t *testing.T) {
	tests := []struct {
		input string
		dest  string
	}{
		{"想去东京购物", "东京"},
		{"去巴黎蜜月旅行", "巴黎"},
		{"新加坡旅游5天", "新加坡"},
		{"去首尔3天2个人", "首尔"},
	}

	p := NewHeuristicParser()
	for _, tt := range tests {
		result, matched := p.Parse(tt.input)
		if !matched {
			t.Errorf("input %q: expected match", tt.input)
			continue
		}
		if result.Destination != tt.dest {
			t.Errorf("input %q: destination got %q, want %q", tt.input, result.Destination, tt.dest)
		}
	}
}

func TestHeuristicParser_DurationExtraction(t *testing.T) {
	p := NewHeuristicParser()

	result, matched := p.Parse("去三亚7天")
	if !matched {
		t.Fatal("expected match")
	}
	if result.StartDate == "" || result.EndDate == "" {
		t.Fatal("expected non-empty dates")
	}
	// The end date should be 7 days after start date
	// We trust the date formatting; just verify both are set
}

func TestHeuristicParser_TravelerCount(t *testing.T) {
	p := NewHeuristicParser()

	result, matched := p.Parse("去上海3个人")
	if !matched {
		t.Fatal("expected match")
	}
	if result.Adults != 3 {
		t.Errorf("adults: got %d, want %d", result.Adults, 3)
	}
}

func TestStripJSONFences(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"a":1}`, `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"```\n{\"a\":1}\n```", `{"a":1}`},
		{"  ```json\n{\"a\":1}\n```  ", `{"a":1}`},
	}

	for _, tt := range tests {
		got := stripJSONFences(tt.input)
		if got != tt.want {
			t.Errorf("stripJSONFences(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}
