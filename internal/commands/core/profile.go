package core

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
)

var (
	profileCPU     bool
	profileMem     bool
	profileTrace   bool
	profileDuration time.Duration
	profileOutput  string
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Run performance profiling",
	Long: `Run CPU, memory, or execution trace profiling on tok.

This command is used for performance analysis and optimization.

Examples:
  tok profile --cpu --duration=30s --output=cpu.prof
  tok profile --mem --duration=10s --output=mem.prof
  tok profile --trace --duration=5s --output=trace.out`,
	RunE: runProfile,
}

func init() {
	registry.Add(func() { registry.Register(profileCmd) })

	profileCmd.Flags().BoolVar(&profileCPU, "cpu", false, "Profile CPU usage")
	profileCmd.Flags().BoolVar(&profileMem, "mem", false, "Profile memory usage")
	profileCmd.Flags().BoolVar(&profileTrace, "trace", false, "Profile execution trace")
	profileCmd.Flags().DurationVarP(&profileDuration, "duration", "d", 30*time.Second, "Profiling duration")
	profileCmd.Flags().StringVarP(&profileOutput, "output", "o", "profile.prof", "Output file path")
}

func runProfile(cmd *cobra.Command, args []string) error {
	// Check that at least one profile type is selected
	if !profileCPU && !profileMem && !profileTrace {
		return fmt.Errorf("must specify at least one profile type (--cpu, --mem, or --trace)")
	}

	// Check that only one profile type is selected
	profileCount := 0
	if profileCPU {
		profileCount++
	}
	if profileMem {
		profileCount++
	}
	if profileTrace {
		profileCount++
	}
	if profileCount > 1 {
		return fmt.Errorf("cannot profile multiple types simultaneously; run separate commands")
	}

	switch {
	case profileCPU:
		return runCPUProfile()
	case profileMem:
		return runMemProfile()
	case profileTrace:
		return runTraceProfile()
	default:
		return fmt.Errorf("unknown profile type")
	}
}

func runCPUProfile() error {
	fmt.Printf("🔥 Starting CPU profiling for %v...\n", profileDuration)
	fmt.Printf("Output: %s\n\n", profileOutput)

	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %w", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("could not start CPU profile: %w", err)
	}
	defer pprof.StopCPUProfile()

	// Run workload
	runWorkload(profileDuration)

	fmt.Println("✅ CPU profiling complete!")
	fmt.Printf("View with: go tool pprof -http=:8080 %s\n", profileOutput)
	return nil
}

func runMemProfile() error {
	fmt.Printf("🧠 Starting memory profiling for %v...\n", profileDuration)
	fmt.Printf("Output: %s\n\n", profileOutput)

	// Run workload first
	runWorkload(profileDuration)

	// Force GC before profiling
	runtime.GC()

	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	fmt.Println("✅ Memory profiling complete!")
	fmt.Printf("View with: go tool pprof -http=:8080 %s\n", profileOutput)
	return nil
}

func runTraceProfile() error {
	fmt.Printf("🎯 Starting execution trace for %v...\n", profileDuration)
	fmt.Printf("Output: %s\n\n", profileOutput)

	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("could not create trace file: %w", err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		return fmt.Errorf("could not start trace: %w", err)
	}
	defer trace.Stop()

	// Run workload
	runWorkload(profileDuration)

	fmt.Println("✅ Execution trace complete!")
	fmt.Printf("View with: go tool trace %s\n", profileOutput)
	return nil
}

// runWorkload simulates tok workload for profiling
func runWorkload(duration time.Duration) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := time.After(duration)
	iterations := 0

	for {
		select {
		case <-done:
			fmt.Printf("Completed %d iterations\n", iterations)
			return
		case <-ticker.C:
			// Simulate command processing
			simulateCommandProcessing()
			iterations++
			if iterations%100 == 0 {
				fmt.Printf("Progress: %d iterations...\r", iterations)
			}
		}
	}
}

// simulateCommandProcessing simulates typical tok operations
func simulateCommandProcessing() {
	// Simulate discover operations
	for i := 0; i < 100; i++ {
		_ = i * i
	}

	// Simulate some memory allocations
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Simulate string operations
	testCmds := []string{
		"git status",
		"cargo test",
		"npm test",
		"docker ps",
		"ls -la",
	}

	for _, cmd := range testCmds {
		_ = len(cmd)
	}
}
