package archive

import (
	"context"
	"time"
)

type Archive struct{}

func New() *Archive { return &Archive{} }

type Entry struct {
	ID        string
	Command   string
	Input     string
	Output    string
	Tokens    int
	Timestamp time.Time
}

type Stats struct {
	TotalCommands int
	TotalTokens   int
	Savings       int
}

func (a *Archive) Search(ctx context.Context, query string) ([]Entry, error) {
	return nil, nil
}

func (a *Archive) Store(ctx context.Context, entry Entry) error {
	return nil
}

func (a *Archive) Get(ctx context.Context, id string) (*Entry, error) {
	return nil, nil
}

func (a *Archive) Delete(ctx context.Context, id string) error {
	return nil
}

func (a *Archive) Stats(ctx context.Context) (*Stats, error) {
	return &Stats{}, nil
}

func (a *Archive) Export(ctx context.Context, format string) ([]byte, error) {
	return nil, nil
}

func (a *Archive) Import(ctx context.Context, data []byte) error {
	return nil
}

func (a *Archive) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	return 0, nil
}
