package discover

import (
	"strings"
)

type Discover struct{}

func New() *Discover {
	return &Discover{}
}

var (
	prefixedCommands = []string{"git", "go", "npm", "npx", "cargo", "docker", "kubectl", "pytest", "jest", "make", "cmake", "gradle", "mvn", "dotnet", "mix", "ruby", "bundle", "rake", "psql", "aws", "terraform", "helm", "ansible"}
)

func RewriteCommand(cmd string, args []string) (string, bool) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return cmd, false
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return cmd, false
	}

	base := parts[0]
	for _, p := range prefixedCommands {
		if base == p {
			return "tokman " + cmd, true
		}
	}

	return cmd, false
}
