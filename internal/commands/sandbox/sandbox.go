package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	sandboxCmd = &cobra.Command{
		Use:   "sandbox",
		Short: "Docker sandbox execution",
		Long: `Run commands in isolated Docker containers for security.
		
Commands run in containers - nothing touches your host system.
No approval prompts, no "are you sure?" dialogs.

Examples:
  tokman sandbox run "npm install"
  tokman sandbox run "cargo build"
  tokman sandbox shell
  tokman sandbox status`,
	}

	runCmd = &cobra.Command{
		Use:   "run [command]",
		Short: "Run command in sandbox",
		RunE:  runSandbox,
	}

	shellCmd = &cobra.Command{
		Use:   "shell",
		Short: "Open interactive shell in sandbox",
		RunE:  runSandboxShell,
	}

	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show sandbox status",
		RunE:  runSandboxStatus,
	}

	imagesCmd = &cobra.Command{
		Use:   "images",
		Short: "List available images",
		RunE:  runSandboxImages,
	}

	image    string
	mountDir string
	envVars  []string
)

const (
	defaultImage = "tokman/sandbox:latest"
	sandboxDir   = ".tokman-sandbox"
)

func init() {
	sandboxCmd.AddCommand(runCmd, shellCmd, statusCmd, imagesCmd)

	runCmd.Flags().StringVar(&image, "image", defaultImage, "Docker image to use")
	runCmd.Flags().StringVar(&mountDir, "mount", "", "Directory to mount")
	runCmd.Flags().StringSliceVar(&envVars, "env", []string{}, "Environment variables")

	shellCmd.Flags().StringVar(&image, "image", defaultImage, "Docker image")

	registry.Add(func() {
		registry.Register(sandboxCmd)
	})
}

func runSandbox(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command required")
	}

	if !dockerAvailable() {
		return fmt.Errorf("Docker not available. Install Docker first.")
	}

	command := strings.Join(args, " ")

	fmt.Printf("[Sandbox] Running: %s\n", command)
	fmt.Printf("Image: %s\n", image)

	workDir, _ := os.Getwd()
	mountPath := workDir
	if mountDir != "" {
		mountPath = mountDir
	}

	dockerArgs := []string{"run", "--rm", "-it"}

	projectName := filepath.Base(workDir)
	dockerArgs = append(dockerArgs, "--name", "tokman-"+projectName)

	dockerArgs = append(dockerArgs, "-v", mountPath+":/workspace")
	dockerArgs = append(dockerArgs, "-w", "/workspace")

	for _, e := range envVars {
		dockerArgs = append(dockerArgs, "-e", e)
	}

	dockerArgs = append(dockerArgs, image, "sh", "-c", command)

	execCmd := exec.Command("docker", dockerArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stdin = os.Stdin
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func runSandboxShell(cmd *cobra.Command, args []string) error {
	if !dockerAvailable() {
		return fmt.Errorf("Docker not available")
	}

	workDir, _ := os.Getwd()
	projectName := filepath.Base(workDir)

	dockerArgs := []string{
		"run", "--rm", "-it",
		"--name", "tokman-shell-" + projectName,
		"-v", workDir + ":/workspace",
		"-w", "/workspace",
		image,
	}

	execCmd := exec.Command("docker", dockerArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stdin = os.Stdin
	execCmd.Stderr = os.Stderr

	fmt.Println("Starting interactive shell in sandbox...")
	fmt.Println("Type 'exit' to leave")

	return execCmd.Run()
}

func runSandboxStatus(cmd *cobra.Command, args []string) error {
	if !dockerAvailable() {
		fmt.Println("Docker: not available")
		return nil
	}

	fmt.Println("=== Sandbox Status ===")
	fmt.Println("Docker: available")

	containers, err := listContainers()
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		return nil
	}

	if len(containers) == 0 {
		fmt.Println("Active containers: none")
	} else {
		fmt.Printf("Active containers: %d\n", len(containers))
		for _, c := range containers {
			fmt.Printf("  - %s\n", c)
		}
	}

	return nil
}

func runSandboxImages(cmd *cobra.Command, args []string) error {
	if !dockerAvailable() {
		return fmt.Errorf("Docker not available")
	}

	execCmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	fmt.Println("=== Available Images ===")
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fmt.Println("  ", line)
		}
	}

	return nil
}

func dockerAvailable() bool {
	execCmd := exec.Command("docker", "info")
	return execCmd.Run() == nil
}

func listContainers() ([]string, error) {
	execCmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := execCmd.Output()
	if err != nil {
		return nil, err
	}

	var containers []string
	for _, c := range strings.Split(string(output), "\n") {
		if c := strings.TrimSpace(c); c != "" {
			containers = append(containers, c)
		}
	}

	return containers, nil
}
