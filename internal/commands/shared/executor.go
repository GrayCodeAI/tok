package shared

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/tee"
	"github.com/GrayCodeAI/tokman/internal/tracking"
	"github.com/GrayCodeAI/tokman/internal/utils"
)

// Command execution, recording, and tee-on-failure.
// Depends on core, tracking, tee, and config packages.

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

	tracker, err := tracking.NewTracker(cfg.GetDatabasePath())
	if err != nil {
		return err
	}
	defer tracker.Close()

	originalTokens := tracking.EstimateTokens(originalOutput)
	filteredTokens := tracking.EstimateTokens(filteredOutput)
	savedTokens := 0
	if originalTokens > filteredTokens {
		savedTokens = originalTokens - filteredTokens
	}

	record := &tracking.CommandRecord{
		Command:        command,
		OriginalOutput: originalOutput,
		FilteredOutput: filteredOutput,
		OriginalTokens: originalTokens,
		FilteredTokens: filteredTokens,
		SavedTokens:    savedTokens,
		ProjectPath:    config.ProjectPath(),
		ExecTimeMs:     execTimeMs,
		Timestamp:      time.Now(),
		ParseSuccess:   success,
		// AI Agent attribution from environment
		AgentName:   os.Getenv("TOKMAN_AGENT"),
		ModelName:   os.Getenv("TOKMAN_MODEL"),
		Provider:    os.Getenv("TOKMAN_PROVIDER"),
		ModelFamily: utils.GetModelFamily(os.Getenv("TOKMAN_MODEL")),
	}

	return tracker.Record(record)
}

// ExecuteAndRecord runs a command function, prints output, and records metrics.
// This consolidates the common pattern of: time -> execute -> print -> record.
// Returns an error instead of calling os.Exit so callers control exit behavior.
func ExecuteAndRecord(name string, fn func() (string, string, error)) error {
	startTime := time.Now()
	raw, filtered, err := fn()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		return err
	}

	fmt.Print(filtered)

	// Use remote analytics if in remote mode
	if IsRemoteMode() {
		origTokens := core.EstimateTokens(raw)
		filteredTokens := core.EstimateTokens(filtered)
		if rerr := RemoteRecordAnalytics(name, origTokens, filteredTokens, execTime, true); rerr != nil && Verbose > 0 {
			fmt.Fprintf(os.Stderr, "Warning: failed to record remote analytics: %v\n", rerr)
		}
		return nil
	}

	if rerr := RecordCommand(name, raw, filtered, execTime, true); rerr != nil && Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Warning: failed to record: %v\n", rerr)
	}
	return nil
}

// RunAndCapture executes a command and captures stdout/stderr.
// This consolidates the common pattern of: exec.Command -> StdoutPipe -> StderrPipe -> Start -> Wait.
// Returns combined stdout and stderr, exit code, and any execution error.
func RunAndCapture(cmd string, args []string) (output string, exitCode int, err error) {
	execCmd := exec.Command(cmd, args...)

	stdoutPipe, pipeErr := execCmd.StdoutPipe()
	if pipeErr != nil {
		return "", 1, fmt.Errorf("creating stdout pipe: %w", pipeErr)
	}

	stderrPipe, pipeErr := execCmd.StderrPipe()
	if pipeErr != nil {
		return "", 1, fmt.Errorf("creating stderr pipe: %w", pipeErr)
	}

	if startErr := execCmd.Start(); startErr != nil {
		return "", 1, fmt.Errorf("starting command: %w", startErr)
	}

	var stdoutBuf, stderrBuf strings.Builder
	doneOut := make(chan struct{})
	doneErr := make(chan struct{})

	go func() {
		io.Copy(&stdoutBuf, stdoutPipe)
		close(doneOut)
	}()

	go func() {
		io.Copy(&stderrBuf, stderrPipe)
		close(doneErr)
	}()

	<-doneOut
	<-doneErr

	waitErr := execCmd.Wait()
	exitCode = 0
	if execCmd.ProcessState != nil {
		exitCode = execCmd.ProcessState.ExitCode()
	}

	output = stdoutBuf.String() + stderrBuf.String()
	return output, exitCode, waitErr
}
