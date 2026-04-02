package cursorsync

import (
	"database/sql"
	"time"
)

type CursorAccount struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Active   bool      `json:"active"`
	LastSync time.Time `json:"last_sync"`
}

type CursorSync struct {
	db       *sql.DB
	accounts map[string]*CursorAccount
}

func NewCursorSync(db *sql.DB) *CursorSync {
	return &CursorSync{
		db:       db,
		accounts: make(map[string]*CursorAccount),
	}
}

func (cs *CursorSync) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS cursor_accounts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		active BOOLEAN NOT NULL DEFAULT 1,
		last_sync DATETIME
	);
	`
	_, err := cs.db.Exec(query)
	return err
}

func (cs *CursorSync) AddAccount(account *CursorAccount) {
	cs.accounts[account.ID] = account
}

func (cs *CursorSync) SwitchAccount(id string) error {
	for _, acc := range cs.accounts {
		acc.Active = false
	}
	if acc, ok := cs.accounts[id]; ok {
		acc.Active = true
		acc.LastSync = time.Now()
		return nil
	}
	return nil
}

func (cs *CursorSync) GetActiveAccount() *CursorAccount {
	for _, acc := range cs.accounts {
		if acc.Active {
			return acc
		}
	}
	return nil
}

func (cs *CursorSync) ListAccounts() []*CursorAccount {
	var accounts []*CursorAccount
	for _, acc := range cs.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

func (cs *CursorSync) Status() map[string]interface{} {
	active := cs.GetActiveAccount()
	return map[string]interface{}{
		"total_accounts": len(cs.accounts),
		"active_account": active,
	}
}
