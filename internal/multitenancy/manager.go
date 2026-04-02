package multitenancy

import (
	"database/sql"
	"time"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	Quota     int       `json:"quota"`
	Used      int       `json:"used"`
	Features  []string  `json:"features"`
	CreatedAt time.Time `json:"created_at"`
}

type TenantManager struct {
	db      *sql.DB
	tenants map[string]*Tenant
}

func NewTenantManager(db *sql.DB) *TenantManager {
	return &TenantManager{
		db:      db,
		tenants: make(map[string]*Tenant),
	}
}

func (tm *TenantManager) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		plan TEXT NOT NULL DEFAULT 'free',
		quota INTEGER NOT NULL DEFAULT 10000,
		used INTEGER NOT NULL DEFAULT 0,
		features TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := tm.db.Exec(query)
	return err
}

func (tm *TenantManager) Create(tenant *Tenant) error {
	_, err := tm.db.Exec(`
		INSERT INTO tenants (id, name, plan, quota, used, features)
		VALUES (?, ?, ?, ?, 0, '[]')
	`, tenant.ID, tenant.Name, tenant.Plan, tenant.Quota)
	if err == nil {
		tm.tenants[tenant.ID] = tenant
	}
	return err
}

func (tm *TenantManager) Get(id string) *Tenant {
	return tm.tenants[id]
}

func (tm *TenantManager) RecordUsage(id string, tokens int) {
	if tenant, ok := tm.tenants[id]; ok {
		tenant.Used += tokens
	}
}

func (tm *TenantManager) IsWithinQuota(id string) bool {
	if tenant, ok := tm.tenants[id]; ok {
		return tenant.Used < tenant.Quota
	}
	return false
}

func (tm *TenantManager) HasFeature(id, feature string) bool {
	if tenant, ok := tm.tenants[id]; ok {
		for _, f := range tenant.Features {
			if f == feature {
				return true
			}
		}
	}
	return false
}

func (tm *TenantManager) List() []*Tenant {
	var tenants []*Tenant
	for _, t := range tm.tenants {
		tenants = append(tenants, t)
	}
	return tenants
}

type RBACRole string

const (
	RoleAdmin  RBACRole = "admin"
	RoleUser   RBACRole = "user"
	RoleViewer RBACRole = "viewer"
)

type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RBACManager struct {
	roles     map[RBACRole][]Permission
	userRoles map[string]RBACRole
}

func NewRBACManager() *RBACManager {
	m := &RBACManager{
		roles:     make(map[RBACRole][]Permission),
		userRoles: make(map[string]RBACRole),
	}
	m.roles[RoleAdmin] = []Permission{
		{Resource: "*", Action: "*"},
	}
	m.roles[RoleUser] = []Permission{
		{Resource: "compression", Action: "execute"},
		{Resource: "tracking", Action: "read"},
		{Resource: "dashboard", Action: "read"},
	}
	m.roles[RoleViewer] = []Permission{
		{Resource: "dashboard", Action: "read"},
		{Resource: "tracking", Action: "read"},
	}
	return m
}

func (m *RBACManager) AssignRole(userID string, role RBACRole) {
	m.userRoles[userID] = role
}

func (m *RBACManager) CheckPermission(userID, resource, action string) bool {
	role, ok := m.userRoles[userID]
	if !ok {
		return false
	}
	perms := m.roles[role]
	for _, p := range perms {
		if (p.Resource == "*" || p.Resource == resource) && (p.Action == "*" || p.Action == action) {
			return true
		}
	}
	return false
}

type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Tenant  string `json:"tenant,omitempty"`
}

type FeatureFlagManager struct {
	flags map[string]*FeatureFlag
}

func NewFeatureFlagManager() *FeatureFlagManager {
	return &FeatureFlagManager{
		flags: make(map[string]*FeatureFlag),
	}
}

func (m *FeatureFlagManager) Set(name string, enabled bool, tenant string) {
	m.flags[name] = &FeatureFlag{
		Name:    name,
		Enabled: enabled,
		Tenant:  tenant,
	}
}

func (m *FeatureFlagManager) IsEnabled(name, tenant string) bool {
	if flag, ok := m.flags[name]; ok {
		if flag.Tenant == "" || flag.Tenant == tenant {
			return flag.Enabled
		}
	}
	return false
}

func (m *FeatureFlagManager) List() []*FeatureFlag {
	var flags []*FeatureFlag
	for _, f := range m.flags {
		flags = append(flags, f)
	}
	return flags
}
