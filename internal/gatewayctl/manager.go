package gatewayctl

import (
	"sync"
	"time"
)

type KillSwitch struct {
	sources map[string]bool
	mu      sync.RWMutex
}

func NewKillSwitch() *KillSwitch {
	return &KillSwitch{
		sources: make(map[string]bool),
	}
}

func (k *KillSwitch) Block(source string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sources[source] = true
}

func (k *KillSwitch) Unblock(source string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.sources, source)
}

func (k *KillSwitch) IsBlocked(source string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.sources[source]
}

func (k *KillSwitch) List() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	var blocked []string
	for s := range k.sources {
		blocked = append(blocked, s)
	}
	return blocked
}

type QuotaManager struct {
	quotas map[string]*QuotaConfig
	usage  map[string]*QuotaUsage
	mu     sync.RWMutex
}

type QuotaConfig struct {
	Source      string        `json:"source"`
	DailyUSD    float64       `json:"daily_usd"`
	MonthlyUSD  float64       `json:"monthly_usd"`
	DailyTokens int64         `json:"daily_tokens"`
	CallLimit   int           `json:"call_limit"`
	ResetPeriod time.Duration `json:"reset_period"`
}

type QuotaUsage struct {
	Source      string    `json:"source"`
	DailyCost   float64   `json:"daily_cost"`
	MonthlyCost float64   `json:"monthly_cost"`
	DailyTokens int64     `json:"daily_tokens"`
	CallCount   int       `json:"call_count"`
	LastReset   time.Time `json:"last_reset"`
}

func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]*QuotaConfig),
		usage:  make(map[string]*QuotaUsage),
	}
}

func (m *QuotaManager) SetQuota(source string, config *QuotaConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	config.Source = source
	m.quotas[source] = config
	m.usage[source] = &QuotaUsage{Source: source, LastReset: time.Now()}
}

func (m *QuotaManager) RecordUsage(source string, cost float64, tokens int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage, ok := m.usage[source]
	if !ok {
		m.usage[source] = &QuotaUsage{Source: source, LastReset: time.Now()}
		usage = m.usage[source]
	}

	quota, ok := m.quotas[source]
	if !ok {
		usage.DailyCost += cost
		usage.MonthlyCost += cost
		usage.DailyTokens += tokens
		usage.CallCount++
		return true
	}

	if quota.DailyUSD > 0 && usage.DailyCost+cost > quota.DailyUSD {
		return false
	}
	if quota.MonthlyUSD > 0 && usage.MonthlyCost+cost > quota.MonthlyUSD {
		return false
	}
	if quota.DailyTokens > 0 && usage.DailyTokens+tokens > quota.DailyTokens {
		return false
	}
	if quota.CallLimit > 0 && usage.CallCount >= quota.CallLimit {
		return false
	}

	usage.DailyCost += cost
	usage.MonthlyCost += cost
	usage.DailyTokens += tokens
	usage.CallCount++
	return true
}

func (m *QuotaManager) CheckQuota(source string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota, ok := m.quotas[source]
	if !ok {
		return true, ""
	}

	usage := m.usage[source]
	if usage == nil {
		return true, ""
	}

	if quota.DailyUSD > 0 && usage.DailyCost >= quota.DailyUSD {
		return false, "daily USD quota exceeded"
	}
	if quota.MonthlyUSD > 0 && usage.MonthlyCost >= quota.MonthlyUSD {
		return false, "monthly USD quota exceeded"
	}
	if quota.DailyTokens > 0 && usage.DailyTokens >= quota.DailyTokens {
		return false, "daily token quota exceeded"
	}
	if quota.CallLimit > 0 && usage.CallCount >= quota.CallLimit {
		return false, "call limit exceeded"
	}

	return true, ""
}

func (m *QuotaManager) GetUsage(source string) *QuotaUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.usage[source]
}

type ModelAlias struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ModelAliaser struct {
	aliases map[string]string
	mu      sync.RWMutex
}

func NewModelAliaser() *ModelAliaser {
	return &ModelAliaser{
		aliases: make(map[string]string),
	}
}

func (a *ModelAliaser) AddAlias(from, to string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.aliases[from] = to
}

func (a *ModelAliaser) Resolve(model string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if alias, ok := a.aliases[model]; ok {
		return alias
	}
	return model
}

func (a *ModelAliaser) List() []ModelAlias {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var result []ModelAlias
	for from, to := range a.aliases {
		result = append(result, ModelAlias{From: from, To: to})
	}
	return result
}
