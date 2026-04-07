package migration

import (
	"context"
	"fmt"
)

type Migration struct {
	ID      string
	Name    string
	Run     func(ctx context.Context) error
	Version int
}

type MigrationManager struct {
	migrations []Migration
}

func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: make([]Migration, 0),
	}
}

func (m *MigrationManager) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

func (m *MigrationManager) RunAll(ctx context.Context) error {
	for _, migration := range m.migrations {
		if err := migration.Run(ctx); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
	}
	return nil
}

func (m *MigrationManager) RunUpTo(ctx context.Context, version int) error {
	for _, migration := range m.migrations {
		if migration.Version > version {
			break
		}
		if err := migration.Run(ctx); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
	}
	return nil
}

func CreateV1Migration() Migration {
	return Migration{
		ID:      "v1",
		Name:    "initial_schema",
		Version: 1,
		Run: func(ctx context.Context) error {
			return nil
		},
	}
}

func CreateV2Migration() Migration {
	return Migration{
		ID:      "v2",
		Name:    "add_indexes",
		Version: 2,
		Run: func(ctx context.Context) error {
			return nil
		},
	}
}
