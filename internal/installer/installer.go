package envfilter

import (
	"os"
	"strings"
)

type EnvFilter struct {
	prefix string
}

func NewEnvFilter(prefix string) *EnvFilter {
	return &EnvFilter{prefix: prefix}
}

func (f *EnvFilter) Filter() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if f.prefix == "" || strings.HasPrefix(parts[0], f.prefix) {
				result[parts[0]] = parts[1]
			}
		}
	}
	return result
}

func (f *EnvFilter) FilterByPrefix(prefixes []string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			for _, prefix := range prefixes {
				if strings.HasPrefix(parts[0], prefix) {
					result[parts[0]] = parts[1]
					break
				}
			}
		}
	}
	return result
}

func (f *EnvFilter) Count() int {
	count := 0
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && (f.prefix == "" || strings.HasPrefix(parts[0], f.prefix)) {
			count++
		}
	}
	return count
}

type TampPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

func NewTampPlugin(name, version string) *TampPlugin {
	return &TampPlugin{
		Name:    name,
		Version: version,
		Enabled: true,
	}
}

func (p *TampPlugin) Status() string {
	if p.Enabled {
		return "active"
	}
	return "inactive"
}

func (p *TampPlugin) Enable() {
	p.Enabled = true
}

func (p *TampPlugin) Disable() {
	p.Enabled = false
}

type Installer struct {
	RepoURL string `json:"repo_url"`
}

func NewInstaller(repoURL string) *Installer {
	if repoURL == "" {
		repoURL = "https://github.com/GrayCodeAI/tokman"
	}
	return &Installer{RepoURL: repoURL}
}

func (i *Installer) OneLiner() string {
	return `curl -sSL ` + i.RepoURL + `/install.sh | bash`
}

func (i *Installer) InstallScript() string {
	return `#!/bin/bash
set -e
echo "Installing TokMan..."
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi
VERSION="latest"
URL="` + i.RepoURL + `/releases/download/$VERSION/tokman-${OS}-${ARCH}"
curl -sSL "$URL" -o /usr/local/bin/tokman
chmod +x /usr/local/bin/tokman
echo "TokMan installed successfully!"
tokman --version`
}

func (i *Installer) InteractiveSetup(yes bool) string {
	if yes {
		return "tokman init --yes"
	}
	return "tokman init"
}
