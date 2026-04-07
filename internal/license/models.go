package license

import "time"

// Tier represents a subscription tier.
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

// License represents a user/team license.
type License struct {
	ID            string
	TeamID        string
	Tier          Tier
	Status        LicenseStatus
	MonthlyQuota  int  // tokens per month
	ApiLimit      int  // requests per day
	CanCustomize  bool // custom filters
	Analytics     bool // full analytics
	TeamUsers     int  // max team members
	SsoEnabled    bool
	PrioritySupport bool
	CreatedAt     time.Time
	ExpiresAt     time.Time
	RenewsAt      time.Time
}

// LicenseStatus represents the license status.
type LicenseStatus string

const (
	StatusActive    LicenseStatus = "active"
	StatusExpired   LicenseStatus = "expired"
	StatusSuspended LicenseStatus = "suspended"
	StatusCanceled  LicenseStatus = "canceled"
)

// TierFeatures defines features available in each tier.
var TierFeatures = map[Tier]*Features{
	TierFree: {
		Name:                "Free",
		Price:               0,
		MonthlyTokenQuota:   1000000,   // 1M tokens
		RequestsPerDay:      100,
		MaxTeamSize:         1,
		Filters:             10,        // Basic 10 layers
		Analytics:           false,
		CustomFilters:       false,
		ApiAccess:           false,
		IdePlugins:          false,
		CloudSync:           false,
		SsoEnabled:          false,
		PrioritySupport:     false,
		Sla:                 "",
	},
	TierPro: {
		Name:                "Professional",
		Price:               99,
		MonthlyTokenQuota:   50000000,  // 50M tokens
		RequestsPerDay:      10000,
		MaxTeamSize:         5,
		Filters:             31,        // All filters
		Analytics:           true,
		CustomFilters:       true,
		ApiAccess:           true,
		IdePlugins:          true,
		CloudSync:           true,
		SsoEnabled:          false,
		PrioritySupport:     false,
		Sla:                 "",
	},
	TierEnterprise: {
		Name:                "Enterprise",
		Price:               -1,        // Custom pricing
		MonthlyTokenQuota:   -1,        // Unlimited
		RequestsPerDay:      -1,        // Unlimited
		MaxTeamSize:         -1,        // Unlimited
		Filters:             31,
		Analytics:           true,
		CustomFilters:       true,
		ApiAccess:           true,
		IdePlugins:          true,
		CloudSync:           true,
		SsoEnabled:          true,
		PrioritySupport:     true,
		Sla:                 "99.9%",
	},
}

// Features represents available features for a tier.
type Features struct {
	Name              string
	Price             int    // USD per month
	MonthlyTokenQuota int
	RequestsPerDay    int
	MaxTeamSize       int
	Filters           int
	Analytics         bool
	CustomFilters     bool
	ApiAccess         bool
	IdePlugins        bool
	CloudSync         bool
	SsoEnabled        bool
	PrioritySupport   bool
	Sla               string
}

// UsageQuota represents monthly usage quota tracking.
type UsageQuota struct {
	TeamID              string
	Month               string // YYYY-MM
	TokensUsed          int
	TokensQuota         int
	ApiCallsUsed        int
	ApiCallsQuota       int
	LastUpdated         time.Time
}

// IsUnlimited checks if usage is unlimited.
func (uq *UsageQuota) IsUnlimited() bool {
	return uq.TokensQuota <= 0
}

// PercentUsed returns the percentage of quota used (0-100).
func (uq *UsageQuota) PercentUsed() int {
	if uq.TokensQuota <= 0 {
		return 0
	}
	return (uq.TokensUsed * 100) / uq.TokensQuota
}

// ExceededQuota checks if quota is exceeded.
func (uq *UsageQuota) ExceededQuota() bool {
	return uq.TokensQuota > 0 && uq.TokensUsed > uq.TokensQuota
}

// FeatureFlag represents a feature flag for a team.
type FeatureFlag struct {
	TeamID       string
	FeatureName  string
	Enabled      bool
	Value        string // JSON encoded config
	ExpiresAt    *time.Time
	CreatedAt    time.Time
}

// BillingEvent represents a billing event (payment, upgrade, etc.)
type BillingEvent struct {
	ID              string
	TeamID          string
	EventType       string // "payment", "upgrade", "downgrade", "overage"
	Amount          float64
	Currency        string
	Description     string
	Status          string // "pending", "completed", "failed"
	CreatedAt       time.Time
}

// PaymentMethod represents a stored payment method.
type PaymentMethod struct {
	ID              string
	TeamID          string
	Type            string // "credit_card", "bank_transfer"
	LastFour        string
	ExpiresAt       time.Time
	IsDefault       bool
	CreatedAt       time.Time
}
