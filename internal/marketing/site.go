package marketing

import "time"

type Testimonial struct {
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	Company string    `json:"company"`
	Quote   string    `json:"quote"`
	Savings string    `json:"savings"`
	Date    time.Time `json:"date"`
}

type CaseStudy struct {
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Industry    string   `json:"industry"`
	Challenge   string   `json:"challenge"`
	Solution    string   `json:"solution"`
	Results     []string `json:"results"`
	Testimonial string   `json:"testimonial"`
}

type PricingTier struct {
	Name        string         `json:"name"`
	Price       float64        `json:"price"`
	Period      string         `json:"period"`
	Features    []string       `json:"features"`
	Limits      map[string]int `json:"limits"`
	Highlighted bool           `json:"highlighted"`
}

type ComparisonEntry struct {
	Feature    string `json:"feature"`
	TokMan     string `json:"tokman"`
	Competitor string `json:"competitor"`
}

type MarketingSite struct {
	testimonials []Testimonial
	caseStudies  []CaseStudy
	pricing      []PricingTier
	comparisons  []ComparisonEntry
}

func NewMarketingSite() *MarketingSite {
	return &MarketingSite{
		pricing: []PricingTier{
			{
				Name:     "Free",
				Price:    0,
				Period:   "forever",
				Features: []string{"31-layer compression", "Basic filters", "CLI access", "Community support"},
				Limits:   map[string]int{"commands_per_day": 1000},
			},
			{
				Name:        "Pro",
				Price:       19,
				Period:      "month",
				Features:    []string{"Everything in Free", "Advanced filters", "Dashboard access", "Priority support", "API access"},
				Limits:      map[string]int{"commands_per_day": 100000},
				Highlighted: true,
			},
			{
				Name:     "Enterprise",
				Price:    99,
				Period:   "month",
				Features: []string{"Everything in Pro", "Multi-tenancy", "SSO", "Custom filters", "SLA guarantee", "Dedicated support"},
				Limits:   map[string]int{"commands_per_day": -1},
			},
		},
		comparisons: []ComparisonEntry{
			{Feature: "Compression Layers", TokMan: "31", Competitor: "3-5"},
			{Feature: "Filter System", TokMan: "TOML + Lua DSL", Competitor: "Basic regex"},
			{Feature: "AI Gateway", TokMan: "Built-in", Competitor: "Separate tool"},
			{Feature: "Security", TokMan: "3-layer defense", Competitor: "Basic scanning"},
			{Feature: "Open Source", TokMan: "MIT License", Competitor: "Proprietary"},
		},
	}
}

func (m *MarketingSite) AddTestimonial(t Testimonial) {
	m.testimonials = append(m.testimonials, t)
}

func (m *MarketingSite) AddCaseStudy(cs CaseStudy) {
	m.caseStudies = append(m.caseStudies, cs)
}

func (m *MarketingSite) GetPricing() []PricingTier {
	return m.pricing
}

func (m *MarketingSite) GetComparisons() []ComparisonEntry {
	return m.comparisons
}

func (m *MarketingSite) GetTestimonials() []Testimonial {
	return m.testimonials
}

func (m *MarketingSite) GetCaseStudies() []CaseStudy {
	return m.caseStudies
}

type Newsletter struct {
	subscribers []string
}

func NewNewsletter() *Newsletter {
	return &Newsletter{}
}

func (n *Newsletter) Subscribe(email string) {
	n.subscribers = append(n.subscribers, email)
}

func (n *Newsletter) Unsubscribe(email string) {
	for i, s := range n.subscribers {
		if s == email {
			n.subscribers = append(n.subscribers[:i], n.subscribers[i+1:]...)
			return
		}
	}
}

func (n *Newsletter) Count() int {
	return len(n.subscribers)
}

type ReferralProgram struct {
	referrals map[string]int
	rewards   map[string]int
}

func NewReferralProgram() *ReferralProgram {
	return &ReferralProgram{
		referrals: make(map[string]int),
		rewards:   make(map[string]int),
	}
}

func (r *ReferralProgram) Refer(referrer, referee string) {
	r.referrals[referrer]++
	r.rewards[referrer] += 10
}

func (r *ReferralProgram) GetRewards(userID string) int {
	return r.rewards[userID]
}
