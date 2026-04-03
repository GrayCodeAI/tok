package web

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy -- <command> [args...]",
	Short: "Execute command without filtering but track usage",
	Long: `Execute a command without applying any output filtering.

Unlike other TokMan commands that filter output to reduce tokens,
proxy runs the command as-is while still tracking execution metrics.
Useful for commands where you need full unfiltered output.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if shared.Verbose > 0 {
			fmt.Fprintf(os.Stderr, "Proxy mode: %s\n", strings.Join(args, " "))
		}

		return runProxyContext(cmd.Context(), args)
	},
}

func init() {
	registry.Add(func() { registry.Register(proxyCmd) })
}

func runProxy(args []string) error {
	return runProxyContext(context.Background(), args)
}

func runProxyContext(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("proxy requires a command to run")
	}

	if err := shared.SanitizeArgs(args); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	timer := tracking.Start()

	execCmd := exec.CommandContext(proxyContext(ctx), args[0], args[1:]...)
	execCmd.Stdin = os.Stdin

	stdoutPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderrPipe, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	outputChan := make(chan []byte)
	go func() {
		var output []byte
		buf := make([]byte, 8192)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				output = append(output, buf[:n]...)
				os.Stdout.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		outputChan <- output
	}()

	errChan := make(chan []byte)
	go func() {
		var errOutput []byte
		buf := make([]byte, 8192)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				errOutput = append(errOutput, buf[:n]...)
				os.Stderr.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		errChan <- errOutput
	}()

	stdout := <-outputChan
	stderr := <-errChan

	err = execCmd.Wait()

	fullOutput := string(stdout) + string(stderr)
	cmdStr := strings.Join(args, " ")
	originalTokens := tracking.EstimateTokens(fullOutput)
	timer.Track(cmdStr, fmt.Sprintf("tokman proxy %s", cmdStr), originalTokens, originalTokens)

	if err != nil {
		return err
	}
	return nil
}

func proxyContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
