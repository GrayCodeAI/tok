package shared

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/core"
	"github.com/GrayCodeAI/tok/internal/filter" // NEW: for progress callback
	"github.com/GrayCodeAI/tok/internal/tee"
	"github.com/GrayCodeAI/tok/internal/tracking"
	"github.com/GrayCodeAI/tok/internal/utils"
)

// Command execution, recording, and tee-on-failure.
// Depends on core, tracking, tee, and config packages.

// TeeOnFailure writes output to tee file on error.
const maxTrackingOutput = 65536 // 64KB cap per field to avoid DB bloat

func truncateForTracking(s string) string {
	if len(s) <= maxTrackingOutput {
		return s
	}
	return s[:maxTrackingOutput] + "\n[truncated]"
}

// TeeOnFailure writes output to tee file on error.
func TeeOnFailure(raw string, commandSlug string, err error) string {
	if err == nil {
		return ""
	}
	exitCode := 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}
	return tee.WriteAndHint(raw, commandSlug, exitCode)
}

// RecordCommand records command execution metrics to the tracking database.
func RecordCommand(command, originalOutput, filteredOutput string, execTimeMs int64, success bool) error {
	cfg, _ := GetCachedConfig() // Error already logged in GetCachedConfig; returns defaults on failure

	if !cfg.Tracking.Enabled {
		return nil
	}

	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		return fmt.Errorf("tracking not available")
	}

	originalTokens := tracking.EstimateTokens(originalOutput)
	filteredTokens := tracking.EstimateTokens(filteredOutput)
	savedTokens := 0
	if originalTokens > filteredTokens {
		savedTokens = originalTokens - filteredTokens
	}

	record := &tracking.CommandRecord{
		Command:        command,
		OriginalOutput: truncateForTracking(originalOutput),
		FilteredOutput: truncateForTracking(filteredOutput),
		OriginalTokens: originalTokens,
		FilteredTokens: filteredTokens,
		SavedTokens:    savedTokens,
		ProjectPath:    config.ProjectPath(),
		ExecTimeMs:     execTimeMs,
		Timestamp:      time.Now(),
		ParseSuccess:   success,
		// AI Agent attribution from environment
		AgentName:   os.Getenv("TOK_AGENT"),
		ModelName:   os.Getenv("TOK_MODEL"),
		Provider:    os.Getenv("TOK_PROVIDER"),
		ModelFamily: utils.GetModelFamily(os.Getenv("TOK_MODEL")),
	}

	return tracker.Record(record)
}

// ExecuteAndRecord runs a command function, prints output, and records metrics.
// This consolidates the common pattern of: time -> execute -> print -> record.
// Returns an error instead of calling os.Exit so callers control exit behavior.
func ExecuteAndRecord(name string, fn func() (string, string, error)) error {
	startTime := time.Now()

	// Start status line
	status := GetStatusLine()
	statusEnabled := IsEnabled()
	if statusEnabled {
		status.Start(name)

		// Install pipeline progress callback
		originalCb := filter.GetProgressCallback()
		filter.SetProgressCallback(func(layer string, inTokens, outTokens int, progress float64) {
			ev := StatusEvent{
				Command:      name,
				Stage:        "compressing",
				Layer:        layer,
				InputTokens:  inTokens,
				OutputTokens: outTokens,
				ProgressPct:  progress,
				Timestamp:    time.Now(),
			}
			status.Publish(ev)
		})
		defer func() { filter.SetProgressCallback(originalCb) }()
	}

	raw, filtered, err := fn()
	execTime := time.Since(startTime).Milliseconds()

	// Clear status before printing output
	if statusEnabled {
		status.Done()
	}

	if err != nil {
		return err
	}

	out.Global().Print(filtered)

	// Use remote analytics if in remote mode
	if IsRemoteMode() {
		origTokens := core.EstimateTokens(raw)
		filteredTokens := core.EstimateTokens(filtered)
		if rerr := RemoteRecordAnalytics(name, origTokens, filteredTokens, execTime, true); rerr != nil && Verbose > 0 {
			out.Global().Errorf("Warning: failed to record remote analytics: %v\n", rerr)
		}
		return nil
	}

	if rerr := RecordCommand(name, raw, filtered, execTime, true); rerr != nil && Verbose > 0 {
		out.Global().Errorf("Warning: failed to record: %v\n", rerr)
	}
	return nil
}

// RunAndCapture executes a command and captures stdout/stderr.
// This consolidates the common pattern of: exec.Command -> StdoutPipe -> StderrPipe -> Start -> Wait.
// Returns combined stdout and stderr, exit code, and any execution error.
func RunAndCapture(cmd string, args []string) (output string, exitCode int, err error) {
	if cmd == "" {
		return "", 1, fmt.Errorf("command is required")
	}

	execCmd := exec.Command(cmd, args...)

	stdoutPipe, pipeErr := execCmd.StdoutPipe()
	if pipeErr != nil {
		return "", 1, fmt.Errorf("creating stdout pipe: %w", pipeErr)
	}
	defer stdoutPipe.Close()

	stderrPipe, pipeErr := execCmd.StderrPipe()
	if pipeErr != nil {
		return "", 1, fmt.Errorf("creating stderr pipe: %w", pipeErr)
	}
	defer stderrPipe.Close()

	if startErr := execCmd.Start(); startErr != nil {
		return "", 1, fmt.Errorf("starting command: %w", startErr)
	}

	var stdoutBuf, stderrBuf strings.Builder
	errCh := make(chan error, 2)

	go func() {
		_, e := io.Copy(&stdoutBuf, stdoutPipe)
		errCh <- e
	}()

	go func() {
		_, e := io.Copy(&stderrBuf, stderrPipe)
		errCh <- e
	}()

	// Drain pipes BEFORE calling Wait() — per os/exec docs, calling Wait before
	// reads complete can truncate output because Wait closes the pipes.
	var pipeErrs []error
	for i := 0; i < 2; i++ {
		if pErr := <-errCh; pErr != nil && pErr != io.EOF {
			pipeErrs = append(pipeErrs, pErr)
		}
	}

	waitErr := execCmd.Wait()
	exitCode = 0
	if execCmd.ProcessState != nil {
		exitCode = execCmd.ProcessState.ExitCode()
	}
	if len(pipeErrs) > 0 && waitErr == nil {
		waitErr = fmt.Errorf("pipe error: %w", pipeErrs[0])
	}

	output = stdoutBuf.String() + stderrBuf.String()
	return output, exitCode, waitErr
}
