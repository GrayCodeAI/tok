package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var dockerJSON bool

func formatAsJSON(output string) string {
	return fmt.Sprintf(`{"output": %s}`, strconv.Quote(output))
}

var dockerCmd = &cobra.Command{
	Use:   "docker [command] [args...]",
	Short: "Docker CLI with filtered output",
	Long: `Docker CLI with token-optimized output.

Specialized filters for common commands:
  - docker ps: Compact container listing
  - docker images: Image size summary
  - docker logs: Log deduplication
  - docker build: Build progress summary
  - docker run: Container output filtering
  - docker inspect: Compact JSON inspection
  - docker stats: Resource usage summary
  - docker compose: Service status and lifecycle
  - docker network/volume/system: Compact listings

Examples:
  tok docker ps
  tok docker images
  tok docker logs <container>
  tok docker build .
  tok docker compose ps
  tok docker inspect <container>`,
	DisableFlagParsing: true,
	RunE:               runDocker,
}

func init() {
	registry.Add(func() { registry.Register(dockerCmd) })
	dockerCmd.Flags().BoolVarP(&dockerJSON, "json", "j", false, "Output as JSON")
}

func runDocker(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runDockerPassthrough(args)
	}

	switch args[0] {
	case "ps":
		return runDockerPs(args[1:])
	case "images":
		return runDockerImages(args[1:])
	case "logs":
		return runDockerLogs(args[1:])
	case "build":
		return runDockerBuild(args[1:])
	case "inspect":
		return runDockerInspect(args[1:])
	case "stats":
		return runDockerStats(args[1:])
	case "top":
		return runDockerTop(args[1:])
	case "port":
		return runDockerPort(args[1:])
	case "network":
		return runDockerNetwork(args[1:])
	case "volume":
		return runDockerVolume(args[1:])
	case "system":
		return runDockerSystem(args[1:])
	case "info":
		return runDockerInfo(args[1:])
	case "version":
		return runDockerVersion(args[1:])
	case "compose":
		return runDockerCompose(args[1:])
	case "stop", "start", "restart", "rm", "rmi", "pull", "push":
		return runDockerLifecycle(args)
	case "exec":
		return runDockerExec(args[1:])
	case "run":
		return runDockerRun(args[1:])
	}
	return runDockerPassthrough(args)
}

func runDockerPassthrough(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", args...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	if dockerJSON {
		out.Global().Println(formatAsJSON(output))
		originalTokens := filter.EstimateTokens(output)
		timer.Track(fmt.Sprintf("docker %s", strings.Join(args, " ")), "tok docker", originalTokens, originalTokens)
		return err
	}

	out.Global().Print(output)

	originalTokens := filter.EstimateTokens(output)
	timer.Track(fmt.Sprintf("docker %s", strings.Join(args, " ")), "tok docker", originalTokens, originalTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runDockerPs(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}\t{{.Ports}}")
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	c.Stdout = stdoutBuf

	err := c.Run()
	output := stdoutBuf.String()

	if strings.TrimSpace(output) == "" {
		out.Global().Println("0 containers")
		timer.Track("docker ps", "tok docker ps", 0, 0)
		return nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := len(lines)

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d containers:", count)
		for i, line := range lines {
			if i >= 5 {
				break
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				filtered += " " + parts[1]
			}
		}
		if count > 5 {
			filtered += fmt.Sprintf(" +%d", count-5)
		}
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker ps", "tok docker ps", originalTokens, filteredTokens)
		return nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d containers:\n", count))

	for i, line := range lines {
		if i >= 15 {
			break
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			id := parts[0]
			if len(id) > 12 {
				id = id[:12]
			}
			name := parts[1]
			shortImage := parts[3]
			if idx := strings.LastIndex(shortImage, "/"); idx != -1 {
				shortImage = shortImage[idx+1:]
			}
			ports := ""
			if len(parts) > 4 && parts[4] != "" {
				ports = fmt.Sprintf(" [%s]", compactPorts(parts[4]))
			}
			result.WriteString(fmt.Sprintf("  %s %s (%s)%s\n", id, name, shortImage, ports))
		}
	}

	if count > 15 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", count-15))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker ps", "tok docker ps", originalTokens, filteredTokens)

	return err
}

func runDockerImages(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.ID}}")
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	c.Stdout = stdoutBuf

	err := c.Run()
	output := stdoutBuf.String()

	if strings.TrimSpace(output) == "" {
		out.Global().Println("0 images")
		timer.Track("docker images", "tok docker images", 0, 0)
		return nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := len(lines)

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d images:", count)
		for i, line := range lines {
			if i >= 5 {
				break
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 1 {
				filtered += " " + parts[0]
			}
		}
		if count > 5 {
			filtered += fmt.Sprintf(" +%d", count-5)
		}
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker images", "tok docker images", originalTokens, filteredTokens)
		return err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d images\n", count))

	for i, line := range lines {
		if i >= 15 {
			break
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 1 {
			image := parts[0]
			size := ""
			if len(parts) > 1 {
				size = fmt.Sprintf(" [%s]", parts[1])
			}
			if len(image) > 40 {
				image = "..." + image[len(image)-37:]
			}
			result.WriteString(fmt.Sprintf("  %s%s\n", image, size))
		}
	}

	if count > 15 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", count-15))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker images", "tok docker images", originalTokens, filteredTokens)

	return err
}

func runDockerLogs(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		out.Global().Println("Usage: tok docker logs <container>")
		return nil
	}

	container := args[0]
	c := exec.Command("docker", "logs", "--tail", "100", container)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterLogOutput(output)
	out.Global().Printf("Logs for %s:\n%s", container, filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("docker logs %s", container), "tok docker logs", originalTokens, filteredTokens)

	return err
}

func runDockerBuild(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"build"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterDockerBuild(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker build", "tok docker build", originalTokens, filteredTokens)

	return err
}

func runDockerInspect(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		out.Global().Println("Usage: tok docker inspect <container|image>")
		return nil
	}

	c := exec.Command("docker", append([]string{"inspect"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	if shared.UltraCompact {
		filtered := compactDockerInspect(raw)
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker inspect", "tok docker inspect", originalTokens, filteredTokens)
		return err
	}

	var data interface{}
	if jsonErr := json.Unmarshal([]byte(raw), &data); jsonErr == nil {
		buf := &strings.Builder{}
		compactJSON(data, "", buf, 2)
		filtered := buf.String()
		out.Global().Print(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker inspect", "tok docker inspect", originalTokens, filteredTokens)
		return err
	}

	out.Global().Print(raw)
	originalTokens := filter.EstimateTokens(raw)
	timer.Track("docker inspect", "tok docker inspect", originalTokens, originalTokens)
	return err
}

func runDockerStats(args []string) error {
	timer := tracking.Start()

	statsArgs := []string{"stats", "--no-stream"}
	statsArgs = append(statsArgs, args...)

	c := exec.Command("docker", statsArgs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	if shared.UltraCompact {
		lines := strings.Split(strings.TrimSpace(raw), "\n")
		if len(lines) <= 1 {
			out.Global().Print(raw)
		} else {
			var result strings.Builder
			for i, line := range lines {
				if i == 0 || i >= len(lines)-1 {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 4 {
					result.WriteString(fmt.Sprintf("%s: cpu=%s mem=%s\n", fields[1], fields[2], fields[3]))
				}
			}
			out.Global().Print(result.String())
		}
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(strings.Split(strings.TrimSpace(raw), "\n")[0])
		timer.Track("docker stats", "tok docker stats", originalTokens, filteredTokens)
		return err
	}

	filtered := filterDockerStats(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker stats", "tok docker stats", originalTokens, filteredTokens)

	return err
}

func runDockerTop(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		out.Global().Println("Usage: tok docker top <container>")
		return nil
	}

	c := exec.Command("docker", append([]string{"top"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterDockerTop(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker top", "tok docker top", originalTokens, filteredTokens)

	return err
}

func runDockerPort(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		out.Global().Println("Usage: tok docker port <container>")
		return nil
	}

	c := exec.Command("docker", append([]string{"port"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	out.Global().Print(raw)

	originalTokens := filter.EstimateTokens(raw)
	timer.Track("docker port", "tok docker port", originalTokens, originalTokens)

	return err
}

func runDockerNetwork(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"ls"}
	}

	if args[0] == "ls" || args[0] == "list" {
		c := exec.Command("docker", append([]string{"network", "ls", "--format", "{{.Name}}\t{{.Driver}}\t{{.Scope}}"}, args[1:]...)...)
		c.Env = os.Environ()

		stdoutBuf := shared.GetBuffer()
		defer shared.PutBuffer(stdoutBuf)
		c.Stdout = stdoutBuf

		err := c.Run()
		output := stdoutBuf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		count := len(lines)
		if strings.TrimSpace(output) == "" {
			count = 0
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("%d networks:\n", count))
		for i, line := range lines {
			if i >= 20 {
				break
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				result.WriteString(fmt.Sprintf("  %s (%s)\n", parts[0], parts[1]))
			}
		}
		if count > 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", count-20))
		}

		filtered := result.String()
		out.Global().Print(filtered)

		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker network ls", "tok docker network ls", originalTokens, filteredTokens)

		return err
	}

	return runDockerPassthrough(append([]string{"network"}, args...))
}

func runDockerVolume(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"ls"}
	}

	if args[0] == "ls" || args[0] == "list" {
		c := exec.Command("docker", append([]string{"volume", "ls", "--format", "{{.Name}}\t{{.Driver}}"}, args[1:]...)...)
		c.Env = os.Environ()

		stdoutBuf := shared.GetBuffer()
		defer shared.PutBuffer(stdoutBuf)
		c.Stdout = stdoutBuf

		err := c.Run()
		output := stdoutBuf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		count := len(lines)
		if strings.TrimSpace(output) == "" {
			count = 0
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("%d volumes:\n", count))
		for i, line := range lines {
			if i >= 20 {
				break
			}
			parts := strings.Split(line, "\t")
			name := parts[0]
			if len(name) > 50 {
				name = "..." + name[len(name)-47:]
			}
			driver := ""
			if len(parts) > 1 {
				driver = fmt.Sprintf(" (%s)", parts[1])
			}
			result.WriteString(fmt.Sprintf("  %s%s\n", name, driver))
		}
		if count > 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", count-20))
		}

		filtered := result.String()
		out.Global().Print(filtered)

		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker volume ls", "tok docker volume ls", originalTokens, filteredTokens)

		return err
	}

	return runDockerPassthrough(append([]string{"volume"}, args...))
}

func runDockerSystem(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"df"}
	}

	if args[0] == "df" {
		c := exec.Command("docker", "system", "df")
		c.Env = os.Environ()

		output, err := c.CombinedOutput()
		raw := string(output)

		filtered := filterDockerSystemDf(raw)
		out.Global().Print(filtered)

		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker system df", "tok docker system df", originalTokens, filteredTokens)

		return err
	}

	if args[0] == "info" {
		return runDockerInfo(args[1:])
	}

	return runDockerPassthrough(append([]string{"system"}, args...))
}

func runDockerInfo(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"info"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterDockerInfo(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker info", "tok docker info", originalTokens, filteredTokens)

	return err
}

func runDockerVersion(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"version"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterDockerVersion(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker version", "tok docker version", originalTokens, filteredTokens)

	return err
}

func runDockerLifecycle(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", args...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterDockerLifecycle(output, args)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("docker %s", args[0]), "tok docker", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runDockerExec(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"exec"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterDockerExecOutput(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker exec", "tok docker exec", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runDockerRun(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"run"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	if shared.UltraCompact {
		filtered := filterDockerRunUltraCompact(output)
		out.Global().Print(filtered)
		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("docker run", "tok docker run", originalTokens, filteredTokens)
		return err
	}

	filtered := filterDockerRunOutput(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker run", "tok docker run", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runDockerCompose(args []string) error {
	if len(args) == 0 {
		return runDockerPassthrough(append([]string{"compose"}, args...))
	}

	switch args[0] {
	case "ps":
		return runComposePs(args[1:])
	case "logs":
		return runComposeLogs(args[1:])
	case "build":
		return runComposeBuild(args[1:])
	case "up":
		return runComposeUp(args[1:])
	case "down":
		return runComposeDown(args[1:])
	case "config":
		return runComposeConfig(args[1:])
	default:
		return runDockerPassthrough(append([]string{"compose"}, args...))
	}
}

func runComposeUp(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"compose", "up"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterComposeUp(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose up", "tok docker compose up", originalTokens, filteredTokens)

	return err
}

func runComposeDown(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"compose", "down"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterComposeDown(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose down", "tok docker compose down", originalTokens, filteredTokens)

	return err
}

func runComposeConfig(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"compose", "config"}, args...)...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var result strings.Builder
	services := 0
	networks := 0
	volumes := 0

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "services:") || strings.HasPrefix(trimmed, "service:") {
			services++
		}
		if strings.HasPrefix(trimmed, "networks:") || strings.HasPrefix(trimmed, "network:") {
			networks++
		}
		if strings.HasPrefix(trimmed, "volumes:") || strings.HasPrefix(trimmed, "volume:") {
			volumes++
		}
	}

	result.WriteString("Compose config valid\n")
	if services > 0 || networks > 0 || volumes > 0 {
		result.WriteString(fmt.Sprintf("  Services defined\n"))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose config", "tok docker compose config", originalTokens, filteredTokens)

	return err
}

func runComposePs(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", "compose", "ps", "--format", "{{.Name}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}")
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	c.Stdout = stdoutBuf

	err := c.Run()
	output := stdoutBuf.String()

	if strings.TrimSpace(output) == "" {
		out.Global().Println("0 compose services")
		timer.Track("docker compose ps", "tok docker compose ps", 0, 0)
		return nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := len(lines)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d compose services:\n", count))

	for i, line := range lines {
		if i >= 20 {
			break
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			name := parts[0]
			image := parts[1]
			status := parts[2]
			ports := ""
			if len(parts) > 3 && parts[3] != "" {
				ports = fmt.Sprintf(" [%s]", compactPorts(parts[3]))
			}
			shortImage := image
			if idx := strings.LastIndex(image, "/"); idx != -1 {
				shortImage = image[idx+1:]
			}
			result.WriteString(fmt.Sprintf("  %s (%s) %s%s\n", name, shortImage, status, ports))
		}
	}

	if count > 20 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", count-20))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose ps", "tok docker compose ps", originalTokens, filteredTokens)

	return err
}

func runComposeLogs(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"compose", "logs", "--tail", "100"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterLogOutput(output)
	out.Global().Printf("Compose logs:\n%s", filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose logs", "tok docker compose logs", originalTokens, filteredTokens)

	return err
}

func runComposeBuild(args []string) error {
	timer := tracking.Start()

	c := exec.Command("docker", append([]string{"compose", "build"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterComposeBuild(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("docker compose build", "tok docker compose build", originalTokens, filteredTokens)

	return err
}

// --- Filter functions ---

func compactPorts(ports string) string {
	if ports == "" {
		return "-"
	}

	portNums := []string{}
	for _, p := range strings.Split(ports, ",") {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "->"); idx != -1 {
			p = p[:idx]
		}
		if idx := strings.LastIndex(p, ":"); idx != -1 {
			p = p[idx+1:]
		}
		if p != "" {
			portNums = append(portNums, p)
		}
	}

	if len(portNums) <= 3 {
		return strings.Join(portNums, ", ")
	}
	return fmt.Sprintf("%s, ... +%d", strings.Join(portNums[:2], ", "), len(portNums)-2)
}

func filterLogOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	prevLine := ""
	repeatCount := 0

	for _, line := range lines {
		if line == prevLine {
			repeatCount++
			continue
		}

		if repeatCount > 1 {
			result.WriteString(fmt.Sprintf("  [repeated %dx]\n", repeatCount))
		}
		repeatCount = 0

		if strings.TrimSpace(line) != "" {
			result.WriteString(line + "\n")
		}
		prevLine = line
	}

	if repeatCount > 1 {
		result.WriteString(fmt.Sprintf("  [repeated %dx]\n", repeatCount))
	}

	return result.String()
}

func filterComposeBuild(output string) string {
	var result strings.Builder

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Building") && strings.Contains(line, "FINISHED") {
			result.WriteString(fmt.Sprintf("Build: %s\n", strings.TrimSpace(line)))
			break
		}
	}

	if result.Len() == 0 {
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Building") {
				result.WriteString(fmt.Sprintf("Build: %s\n", strings.TrimSpace(line)))
				break
			}
		}
	}

	if result.Len() == 0 {
		result.WriteString("Build:\n")
	}

	services := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		if idx := strings.Index(line, "["); idx != -1 {
			if idx2 := strings.Index(line[idx:], "]"); idx2 != -1 {
				bracket := line[idx+1 : idx+idx2]
				parts := strings.Fields(bracket)
				if len(parts) > 0 {
					svc := parts[0]
					if svc != "" && svc != "+" {
						services[svc] = true
					}
				}
			}
		}
	}

	if len(services) > 0 {
		var names []string
		for name := range services {
			names = append(names, name)
		}
		result.WriteString(fmt.Sprintf("  Services: %s\n", strings.Join(names, ", ")))
	}

	return result.String()
}

func filterDockerBuild(output string) string {
	var result strings.Builder
	var steps []string
	var errors []string
	var warnings []string
	var imageID string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Step") {
			parts := strings.SplitN(trimmed, " : ", 2)
			if len(parts) == 2 {
				stepDesc := parts[1]
				if len(stepDesc) > 60 {
					stepDesc = stepDesc[:57] + "..."
				}
				steps = append(steps, stepDesc)
			}
			continue
		}

		if strings.Contains(trimmed, "Successfully built") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				imageID = parts[len(parts)-1]
			}
			continue
		}

		if strings.Contains(trimmed, "Successfully tagged") {
			continue
		}

		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "ERROR") {
			errors = append(errors, trimmed)
			continue
		}

		if strings.Contains(trimmed, "WARNING") {
			warnings = append(warnings, trimmed)
			continue
		}
	}

	result.WriteString(fmt.Sprintf("Build: %d steps\n", len(steps)))

	if len(steps) > 0 && len(steps) <= 5 {
		for _, step := range steps {
			result.WriteString(fmt.Sprintf("  %s\n", step))
		}
	}

	if imageID != "" {
		result.WriteString(fmt.Sprintf("  Image: %s\n", imageID))
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("  Errors: %d\n", len(errors)))
		for i, e := range errors {
			if i >= 3 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-3))
				break
			}
			result.WriteString(fmt.Sprintf("    %s\n", shared.TruncateLine(e, 80)))
		}
	}

	if len(warnings) > 0 {
		result.WriteString(fmt.Sprintf("  Warnings: %d\n", len(warnings)))
	}

	if result.Len() == 0 {
		return filterLogOutput(output)
	}

	return result.String()
}

func filterDockerStats(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return raw
	}

	var result strings.Builder
	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			name := fields[1]
			cpu := fields[2]
			mem := fields[3]
			net := ""
			if len(fields) >= 7 {
				net = fields[5]
			}
			block := ""
			if len(fields) >= 9 {
				block = fields[7]
			}
			pid := ""
			if len(fields) >= 10 {
				pid = fields[8]
			}
			if shared.UltraCompact {
				result.WriteString(fmt.Sprintf("%s cpu=%s mem=%s\n", name, cpu, mem))
			} else {
				result.WriteString(fmt.Sprintf("  %-20s cpu=%-8s mem=%-10s net=%s block=%s pids=%s\n", name, cpu, mem, net, block, pid))
			}
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterDockerTop(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return raw
	}

	var result strings.Builder
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			result.WriteString(fmt.Sprintf("%s %s %s %s\n", fields[0], fields[1], fields[4], fields[5]))
		} else {
			result.WriteString(line + "\n")
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func compactDockerInspect(raw string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}

	arr, ok := data.([]interface{})
	if !ok || len(arr) == 0 {
		return raw
	}

	var result strings.Builder
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name := ""
		if n, ok := obj["Name"].(string); ok {
			name = strings.TrimPrefix(n, "/")
		}
		state, _ := obj["State"].(map[string]interface{})
		status := ""
		if s, ok := state["Status"].(string); ok {
			status = s
		}
		image, _ := obj["Config"].(map[string]interface{})
		imageName := ""
		if img, ok := image["Image"].(string); ok {
			imageName = img
		}

		result.WriteString(fmt.Sprintf("%s: %s (%s)\n", name, status, imageName))
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func compactJSON(data interface{}, prefix string, buf *strings.Builder, depth int) {
	if depth <= 0 {
		buf.WriteString(prefix + "...\n")
		return
	}

	switch v := data.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		for _, key := range keys {
			val := v[key]
			switch val := val.(type) {
			case string:
				if len(val) > 100 {
					val = val[:97] + "..."
				}
				buf.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, key, val))
			case float64:
				buf.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, val))
			case bool:
				buf.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, val))
			case nil:
				buf.WriteString(fmt.Sprintf("%s%s: null\n", prefix, key))
			case []interface{}:
				if len(val) > 5 {
					buf.WriteString(fmt.Sprintf("%s%s: [%d items]\n", prefix, key, len(val)))
				} else {
					buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
					compactJSON(val, prefix+"  ", buf, depth-1)
				}
			case map[string]interface{}:
				if len(val) > 10 {
					buf.WriteString(fmt.Sprintf("%s%s: {%d keys}\n", prefix, key, len(val)))
				} else {
					buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
					compactJSON(val, prefix+"  ", buf, depth-1)
				}
			default:
				buf.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, val))
			}
		}
	case []interface{}:
		for i, item := range v {
			if i >= 10 {
				buf.WriteString(fmt.Sprintf("%s... +%d more\n", prefix, len(v)-10))
				break
			}
			compactJSON(item, prefix+"  ", buf, depth-1)
		}
	default:
		buf.WriteString(fmt.Sprintf("%s%v\n", prefix, v))
	}
}

func filterDockerSystemDf(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return raw
	}

	var result strings.Builder
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			result.WriteString(fmt.Sprintf("%-20s %s total, %s active, %s reclaimable\n",
				fields[0], fields[1], fields[2], fields[3]))
		} else if len(fields) >= 2 {
			result.WriteString(fmt.Sprintf("%-20s %s\n", fields[0], fields[1]))
		}
	}

	filtered := result.String()
	if filtered == "" {
		return raw
	}
	return filtered
}

func filterDockerInfo(raw string) string {
	var result strings.Builder
	keyFields := map[string]bool{
		"Server Version":   true,
		"Storage Driver":   true,
		"Operating System": true,
		"Architecture":     true,
		"CPUs":             true,
		"Total Memory":     true,
		"Containers":       true,
		"Running":          true,
		"Paused":           true,
		"Stopped":          true,
		"Images":           true,
		"Docker Root Dir":  true,
		"Kernel Version":   true,
	}

	if shared.UltraCompact {
		parts := []string{}
		for _, line := range strings.Split(raw, "\n") {
			if strings.Contains(line, ":") {
				partsSplit := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(partsSplit[0])
				if keyFields[key] {
					val := strings.TrimSpace(partsSplit[1])
					parts = append(parts, fmt.Sprintf("%s=%s", key, val))
				}
			}
		}
		return strings.Join(parts, " ") + "\n"
	}

	for _, line := range strings.Split(raw, "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			if keyFields[key] {
				result.WriteString(line + "\n")
			}
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterDockerVersion(raw string) string {
	var result strings.Builder

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Client:") {
			result.WriteString("Client:\n")
			continue
		}
		if strings.Contains(trimmed, "Server:") {
			result.WriteString("Server:\n")
			continue
		}
		if strings.HasPrefix(trimmed, "Version:") || strings.HasPrefix(trimmed, "API version:") || strings.HasPrefix(trimmed, "Go version:") {
			result.WriteString("  " + trimmed + "\n")
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterDockerLifecycle(output string, args []string) string {
	cmd := args[0]

	if shared.UltraCompact {
		if strings.Contains(output, "Error") || strings.Contains(output, "error") {
			return fmt.Sprintf("%s: failed\n", cmd)
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		nonEmpty := 0
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				nonEmpty++
			}
		}
		if nonEmpty == 0 || nonEmpty == 1 {
			return fmt.Sprintf("%s: ok\n", cmd)
		}
		return fmt.Sprintf("%s: %d lines output\n", cmd, nonEmpty)
	}

	return output
}

func filterDockerExecOutput(output string) string {
	return filterLogOutput(output)
}

func filterDockerRunOutput(output string) string {
	if shared.UltraCompact {
		return filterDockerRunUltraCompact(output)
	}

	var result strings.Builder
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "error") {
			errors = append(errors, trimmed)
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", shared.TruncateLine(e, 80)))
		}
	}

	other := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" && !strings.Contains(line, "Error") && !strings.Contains(line, "error") {
			other++
		}
	}

	if other > 0 && len(errors) == 0 {
		return filterLogOutput(output)
	}

	if result.Len() == 0 {
		return filterLogOutput(output)
	}

	return result.String()
}

func filterDockerRunUltraCompact(output string) string {
	lines := strings.Split(output, "\n")
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	return fmt.Sprintf("%d lines\n", nonEmpty)
}

func filterComposeUp(output string) string {
	var result strings.Builder
	services := make(map[string]bool)
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "error") {
			errors = append(errors, trimmed)
			continue
		}
		if strings.Contains(trimmed, "Creating") || strings.Contains(trimmed, "Starting") || strings.Contains(trimmed, "Started") || strings.Contains(trimmed, "Healthy") {
			fields := strings.Fields(trimmed)
			for _, f := range fields {
				if strings.Contains(f, "-") && !strings.HasPrefix(f, "-") && len(f) > 3 {
					services[f] = true
				}
			}
			result.WriteString(trimmed + "\n")
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors:\n"))
		for _, e := range errors {
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return filterLogOutput(output)
	}
	return result.String()
}

func filterComposeDown(output string) string {
	var result strings.Builder
	removed := 0
	networks := 0

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Removed") || strings.Contains(trimmed, "Removing") {
			removed++
		}
		if strings.Contains(trimmed, "Network") || strings.Contains(trimmed, "Removing network") {
			networks++
		}
		if strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "error") {
			result.WriteString(trimmed + "\n")
		}
	}

	if removed > 0 || networks > 0 {
		result.WriteString(fmt.Sprintf("Down: %d containers, %d networks removed\n", removed, networks))
	} else if result.Len() == 0 {
		return "Down: complete\n"
	}

	return result.String()
}
