package security

import (
	"sync"
	"testing"
)

func TestPluginSandbox(t *testing.T) {
	ps := NewPluginSandbox(1024, 2)
	out, err := ps.Execute("plugin", "input")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out != "input" {
		t.Errorf("expected %q, got %q", "input", out)
	}
}

func TestAuditLogger(t *testing.T) {
	al := NewAuditLogger()
	al.Log(AuditEntry{Timestamp: "2026-01-01", User: "alice", Action: "read", Resource: "file1"})
	al.Log(AuditEntry{Timestamp: "2026-01-02", User: "bob", Action: "write", Resource: "file2"})

	entries := al.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].User != "alice" {
		t.Errorf("expected user alice, got %q", entries[0].User)
	}
	if entries[1].Action != "write" {
		t.Errorf("expected action write, got %q", entries[1].Action)
	}
}

func TestAuditLogger_Concurrent(t *testing.T) {
	al := NewAuditLogger()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			al.Log(AuditEntry{User: "user", Action: "test"})
		}(i)
		go func() {
			defer wg.Done()
			al.Entries()
		}()
	}
	wg.Wait()
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no null", "hello", "hello"},
		{"with null", "hel\x00lo", "hello"},
		{"multiple nulls", "a\x00b\x00c", "abc"},
		{"only nulls", "\x00\x00\x00", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeInput(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3)

	if !rl.Allow("user1") {
		t.Error("first request should be allowed")
	}
	if !rl.Allow("user1") {
		t.Error("second request should be allowed")
	}
	if !rl.Allow("user1") {
		t.Error("third request should be allowed")
	}
	if rl.Allow("user1") {
		t.Error("fourth request should be denied")
	}

	// Different key should have its own limit
	if !rl.Allow("user2") {
		t.Error("user2 first request should be allowed")
	}
}

func TestRateLimiter_DefaultLimit(t *testing.T) {
	rl := NewRateLimiter(0)
	if rl.limit != 100 {
		t.Errorf("expected default limit 100, got %d", rl.limit)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(1000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(10)
		for j := 0; j < 10; j++ {
			go func() {
				defer wg.Done()
				rl.Allow("user")
			}()
		}
	}
	wg.Wait()
}

func TestRBAC(t *testing.T) {
	r := NewRBAC()

	// Deny by default
	if r.HasPermission("alice", "read") {
		t.Error("should deny by default")
	}

	r.RegisterRole("admin", []string{"read", "write", "delete"})
	r.AssignRole("alice", "admin")

	if !r.HasPermission("alice", "read") {
		t.Error("alice should have read permission")
	}
	if !r.HasPermission("alice", "delete") {
		t.Error("alice should have delete permission")
	}
	if r.HasPermission("alice", "execute") {
		t.Error("alice should not have execute permission")
	}

	// Unknown user
	if r.HasPermission("bob", "read") {
		t.Error("bob should not have any permission")
	}
}

func TestRBAC_Concurrent(t *testing.T) {
	r := NewRBAC()
	r.RegisterRole("admin", []string{"read", "write"})
	r.AssignRole("alice", "admin")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			r.HasPermission("alice", "read")
		}()
		go func() {
			defer wg.Done()
			r.AssignRole("bob", "admin")
		}()
	}
	wg.Wait()
}

func TestSecretsManager(t *testing.T) {
	sm := NewSecretsManager()

	sm.Set("api_key", "secret123")
	if sm.Get("api_key") != "secret123" {
		t.Errorf("expected secret123, got %q", sm.Get("api_key"))
	}

	if sm.Get("nonexistent") != "" {
		t.Error("expected empty string for nonexistent key")
	}
}

func TestSecretsManager_Concurrent(t *testing.T) {
	sm := NewSecretsManager()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			sm.Set("key", "value")
		}(i)
		go func() {
			defer wg.Done()
			sm.Get("key")
		}()
	}
	wg.Wait()
}

func TestTLSConfig(t *testing.T) {
	tc := &TLSConfig{CertFile: "cert.pem", KeyFile: "key.pem"}
	if err := tc.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}

	tc2 := &TLSConfig{CertFile: "", KeyFile: "key.pem"}
	if err := tc2.Validate(); err == nil {
		t.Error("expected error for empty CertFile")
	}

	tc3 := &TLSConfig{CertFile: "cert.pem", KeyFile: ""}
	if err := tc3.Validate(); err == nil {
		t.Error("expected error for empty KeyFile")
	}
}

func TestSecurityScanner(t *testing.T) {
	ss := &SecurityScanner{}

	findings := ss.Scan("some code")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}

	if ss.Scanned() != 1 {
		t.Errorf("expected 1 scanned, got %d", ss.Scanned())
	}

	if ss.Found() != 0 {
		t.Errorf("expected 0 found, got %d", ss.Found())
	}

	// Second scan
	ss.Scan("more code")
	if ss.Scanned() != 2 {
		t.Errorf("expected 2 scanned, got %d", ss.Scanned())
	}
}
