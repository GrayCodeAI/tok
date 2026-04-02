package retention

type Config struct {
	MaxDays int `json:"max_days"`
}

type Manager struct {
	config Config
}

func NewManager(maxDays int) *Manager {
	if maxDays == 0 {
		maxDays = 90
	}
	return &Manager{config: Config{MaxDays: maxDays}}
}

func (m *Manager) GetMaxDays() int {
	return m.config.MaxDays
}

func (m *Manager) SetMaxDays(days int) {
	m.config.MaxDays = days
}
