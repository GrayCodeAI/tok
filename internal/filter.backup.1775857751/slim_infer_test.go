package filter

import (
	"strings"
	"testing"
)

func TestSlimInferFilter_Name(t *testing.T) {
	f := NewSlimInferFilter()
	if f.Name() != "30_slim_infer" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestSlimInferFilter_ModeNone(t *testing.T) {
	f := NewSlimInferFilter()
	input := strings.Repeat("an orphan line with unique vocabulary\n", 10)
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return unchanged")
	}
}

func TestSlimInferFilter_ShortInput(t *testing.T) {
	f := NewSlimInferFilter()
	input := "line one\nline two"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short input should pass through")
	}
}

func TestSlimInferFilter_DropsOrphanLines(t *testing.T) {
	f := NewSlimInferFilter()
	// Most lines share terms about "connection" and "database"
	// The orphan line has unique vocabulary that appears nowhere else
	lines := []string{
		"The connection pool manages database connections efficiently.",
		"Database connections are reused from the connection pool cache.",
		"Pool size controls maximum concurrent database connection count.",
		"Connection timeout disconnects idle database pool connections.",
		"The database connection pool monitors health and availability.",
		"ORPHAN: zephyr quokka byzantine xylophone flibbertigibbet.", // unique vocab
		"Connection pool exhaustion causes database request failures.",
		"ERROR: database connection pool limit reached maximum capacity.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings by dropping orphan line")
	}
	if strings.Contains(out, "ORPHAN:") {
		t.Error("orphan line with zero reference score should be dropped")
	}
	if !strings.Contains(out, "ERROR:") {
		t.Error("error line must always be preserved")
	}
}

func TestSlimInferFilter_PreservesConnectedLines(t *testing.T) {
	f := NewSlimInferFilter()
	// All lines share terms — none should be dropped
	lines := []string{
		"Authentication service validates user credentials on login.",
		"User credentials include username and hashed password fields.",
		"Password hashing uses bcrypt algorithm with cost factor twelve.",
		"The authentication token includes user role and expiration claims.",
		"Token validation checks signature and expiration on every request.",
		"Invalid tokens return 401 Unauthorized response to the caller.",
		"Authentication failures are logged for security audit purposes.",
		"Security audit logs include timestamp user IP and failure reason.",
	}
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeMinimal)

	outLines := strings.Split(strings.TrimSpace(out), "\n")
	// All lines are well-connected; most should survive
	if len(outLines) < len(lines)-2 {
		t.Errorf("connected lines should not be aggressively dropped; got %d of %d", len(outLines), len(lines))
	}
}

func TestSlimInferFilter_AggressiveDropsMore(t *testing.T) {
	f := NewSlimInferFilter()
	lines := []string{
		"Core concept: distributed cache invalidation strategy.",
		"Cache invalidation uses event sourcing pattern for consistency.",
		"Events are published to message queue on data modification.",
		"Isolated metric: cache_miss_p99_latency_us_mean_stddev.", // isolated stat
		"Consumers update local cache replicas from event stream.",
		"Stale entries expire based on configured TTL policy.",
		"Random stat footnote: throughput_baseline_ref_3847291a.", // isolated
		"Cache coherence is maintained across all service instances.",
		"ERROR: cache cluster unreachable network timeout exceeded.",
	}
	input := strings.Join(lines, "\n")

	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive should save >= minimal")
	}
}

func TestSlimInferFilter_AlwaysKeepsFirstAndLast(t *testing.T) {
	f := NewSlimInferFilter()
	lines := []string{
		"FIRST: unique_opening_xylophone_zephyr_flibbertigibbet_start.",
	}
	for i := 0; i < 10; i++ {
		lines = append(lines, "connection pool database authentication timeout service.")
	}
	lines = append(lines, "LAST: unique_closing_quokka_byzantine_terminus_end_marker.")
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeAggressive)

	if !strings.Contains(out, "FIRST:") {
		t.Error("first line must always be preserved")
	}
	if !strings.Contains(out, "LAST:") {
		t.Error("last line must always be preserved")
	}
}
