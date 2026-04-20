package tui

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/session"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

type Options struct {
	RefreshInterval time.Duration
	Days            int
	ProjectPath     string
	AgentName       string
	Provider        string
	ModelName       string
	SessionID       string
	Theme           ThemeName
}

func (o Options) normalized() Options {
	if o.RefreshInterval <= 0 {
		o.RefreshInterval = 20 * time.Second
	}
	if o.Days <= 0 {
		o.Days = 30
	}
	if o.Theme == "" {
		o.Theme = ThemeDark
	}
	return o
}

// snapshotLoader fetches the canonical workspace snapshot the TUI renders.
// Implementations must be safe for concurrent Load calls. Close releases any
// held resources (DB handles, file descriptors) and MUST be called on TUI
// teardown. Calling Load after Close returns ErrLoaderClosed.
type snapshotLoader interface {
	Load(ctx context.Context, opts Options) (*tracking.WorkspaceDashboardSnapshot, error)
	Close() error
}

// ErrLoaderClosed is returned by Load after Close has been invoked.
var ErrLoaderClosed = errors.New("tui loader: closed")

// workspaceLoader is the production snapshotLoader backed by a long-lived
// Tracker + SessionManager. It lazily opens the DBs on first Load so that
// unit tests can swap in stubLoader without touching the filesystem.
type workspaceLoader struct {
	mu      sync.Mutex
	tracker *tracking.Tracker
	manager *session.SessionManager
	closed  bool
}

// newWorkspaceLoader returns a fresh loader. DB handles open lazily.
func newWorkspaceLoader() *workspaceLoader {
	return &workspaceLoader{}
}

func (l *workspaceLoader) ensure() error {
	if l.closed {
		return ErrLoaderClosed
	}
	if l.tracker == nil {
		t, err := shared.OpenTracker()
		if err != nil {
			return err
		}
		l.tracker = t
	}
	if l.manager == nil {
		m, err := session.NewSessionManager()
		if err != nil {
			return err
		}
		l.manager = m
	}
	return nil
}

func (l *workspaceLoader) Load(ctx context.Context, opts Options) (*tracking.WorkspaceDashboardSnapshot, error) {
	opts = opts.normalized()

	l.mu.Lock()
	if err := l.ensure(); err != nil {
		l.mu.Unlock()
		return nil, err
	}
	tracker := l.tracker
	manager := l.manager
	l.mu.Unlock()

	// Fast-path the common ctx-cancelled case so callers don't wait on SQLite.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// SQLite queries here are synchronous; we don't get per-query cancellation
	// but returning early keeps the tea.Program Quit path from blocking on a
	// stale in-flight tick.
	type result struct {
		snap *tracking.WorkspaceDashboardSnapshot
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		snap, err := tracking.GetWorkspaceDashboardSnapshot(
			tracker,
			manager,
			tracking.DashboardQueryOptions{
				Days:                 opts.Days,
				ProjectPath:          opts.ProjectPath,
				AgentName:            opts.AgentName,
				Provider:             opts.Provider,
				ModelName:            opts.ModelName,
				SessionID:            opts.SessionID,
				Limit:                8,
				ReductionGoalPct:     40,
				DailyTokenBudget:     100_000,
				WeeklyTokenBudget:    500_000,
				MonthlyTokenBudget:   2_000_000,
				DailyCostBudgetUSD:   5,
				WeeklyCostBudgetUSD:  25,
				MonthlyCostBudgetUSD: 100,
			},
			session.SessionListOptions{
				Agent:       opts.AgentName,
				ProjectPath: opts.ProjectPath,
				Limit:       10,
			},
			10,
		)
		ch <- result{snap, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.snap, r.err
	}
}

// Close releases the tracker + session manager handles. Safe to call
// multiple times; subsequent calls are no-ops.
func (l *workspaceLoader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true

	var errs []error
	if l.tracker != nil {
		if err := l.tracker.Close(); err != nil {
			errs = append(errs, err)
		}
		l.tracker = nil
	}
	if l.manager != nil {
		if err := l.manager.Close(); err != nil {
			errs = append(errs, err)
		}
		l.manager = nil
	}
	return errors.Join(errs...)
}
