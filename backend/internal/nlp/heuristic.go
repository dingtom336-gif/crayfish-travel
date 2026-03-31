package nlp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HeuristicParser extracts travel requirements from Chinese natural language
// using regex patterns. Stateless, no external dependencies.
type HeuristicParser struct{}

// NewHeuristicParser creates a new heuristic parser.
func NewHeuristicParser() *HeuristicParser {
	return &HeuristicParser{}
}

// Known city names for destination matching.
var knownCities = []string{
	// Domestic
	"三亚", "北京", "上海", "成都", "西安", "丽江", "厦门", "杭州", "大理", "重庆",
	"广州", "深圳", "青岛", "南京", "武汉", "长沙", "昆明", "哈尔滨", "桂林", "拉萨",
	"苏州", "黄山", "张家界", "九寨沟", "敦煌", "乌鲁木齐", "呼伦贝尔", "海口", "珠海", "威海",
	// International
	"日本", "东京", "大阪", "京都", "泰国", "曼谷", "清迈", "普吉岛",
	"韩国", "首尔", "新加坡", "马来西亚", "吉隆坡", "越南",
	"巴厘岛", "马尔代夫", "香港", "澳门", "台北",
	"巴黎", "伦敦", "纽约", "悉尼", "罗马", "迪拜",
}

// Preference keyword map: Chinese keywords -> English tag.
var preferenceKeywords = map[string]string{
	"海边": "beachfront", "海岛": "beachfront", "沙滩": "beachfront",
	"亲子": "family-friendly", "带娃": "family-friendly", "带孩子": "family-friendly",
	"泳池": "pool", "游泳": "pool",
	"温泉": "hot-spring",
	"滑雪": "ski",
	"蜜月": "romantic", "情侣": "romantic",
	"美食": "foodie", "吃": "foodie",
	"购物": "shopping",
	"文化": "cultural", "历史": "cultural", "古迹": "cultural",
	"冒险": "adventure", "刺激": "adventure",
	"放松": "relaxing", "散心": "relaxing", "躺平": "relaxing",
}

// Compiled regex patterns.
var (
	// Budget: digits followed by 万/元/块/钱
	budgetRe = regexp.MustCompile(`(\d+)\s*([万元块钱])`)
	// Duration: digits followed by 天/日/晚
	durationRe = regexp.MustCompile(`(\d+)\s*[天日晚]`)
	// Travelers: digits followed by 个/位/人/口
	travelersRe = regexp.MustCompile(`(\d+)\s*[个位人口]`)
)

// Parse extracts travel requirements from Chinese text using heuristic rules.
// Returns (result, true) if at least a destination was matched, otherwise (nil, false).
func (p *HeuristicParser) Parse(rawInput string) (*TravelRequirement, bool) {
	if rawInput == "" {
		return nil, false
	}

	result := &TravelRequirement{
		Adults:      2,
		Children:    0,
		BudgetCents: 0,
		Preferences: []string{},
	}

	// Extract destination
	dest := extractDestination(rawInput)
	if dest == "" {
		return nil, false
	}
	result.Destination = dest

	// Extract budget
	result.BudgetCents = extractBudget(rawInput)

	// Extract duration and compute dates
	duration := extractDuration(rawInput)
	now := time.Now()
	startDate := now.AddDate(0, 0, 7) // default: one week from now
	result.StartDate = startDate.Format("2006-01-02")
	if duration > 0 {
		result.EndDate = startDate.AddDate(0, 0, duration).Format("2006-01-02")
	} else {
		result.EndDate = startDate.AddDate(0, 0, 5).Format("2006-01-02") // default 5 days
	}

	// Extract travelers
	travelers := extractTravelers(rawInput)
	if travelers > 0 {
		result.Adults = travelers
	}

	// Extract preferences
	result.Preferences = extractPreferences(rawInput)

	return result, true
}

// extractDestination finds the first matching city name in the input.
// Supports patterns: 去(城市), (城市)旅游, 想去(城市), or standalone city name.
func extractDestination(input string) string {
	// Try pattern-based matching first (longer patterns for higher confidence)
	for _, city := range knownCities {
		patterns := []string{
			fmt.Sprintf("去%s", city),
			fmt.Sprintf("%s旅游", city),
			fmt.Sprintf("%s旅行", city),
			fmt.Sprintf("想去%s", city),
			fmt.Sprintf("到%s", city),
		}
		for _, pat := range patterns {
			if strings.Contains(input, pat) {
				return city
			}
		}
	}

	// Fall back to standalone city name match (longest match first to avoid
	// partial matches like "大" matching before "大理" or "大阪")
	// Sort by length descending via manual iteration
	for maxLen := 5; maxLen >= 1; maxLen-- {
		for _, city := range knownCities {
			if len([]rune(city)) == maxLen && strings.Contains(input, city) {
				return city
			}
		}
	}

	return ""
}

// extractBudget parses budget from patterns like "8000块", "2万", "5000元".
// Returns budget in cents.
func extractBudget(input string) int64 {
	matches := budgetRe.FindStringSubmatch(input)
	if len(matches) < 3 {
		return 0
	}

	amount, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0
	}

	unit := matches[2]
	if unit == "万" {
		amount *= 10000
	}

	// Convert yuan to cents
	return amount * 100
}

// extractDuration parses duration from patterns like "5天", "3日", "7晚".
func extractDuration(input string) int {
	matches := durationRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0
	}
	d, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return d
}

// extractTravelers parses number of travelers from patterns like "2个人", "3位".
func extractTravelers(input string) int {
	matches := travelersRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return n
}

// extractPreferences scans for known keywords and returns deduplicated preference tags.
func extractPreferences(input string) []string {
	seen := map[string]bool{}
	var prefs []string

	for keyword, tag := range preferenceKeywords {
		if strings.Contains(input, keyword) && !seen[tag] {
			seen[tag] = true
			prefs = append(prefs, tag)
		}
	}

	return prefs
}
