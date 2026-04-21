package configcmd

import (
	"path/filepath"

	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/config"
)

func effectiveConfigPath() string {
	if shared.CfgFile != "" {
		return shared.CfgFile
	}
	return config.ConfigPath()
}

func effectiveConfigDir() string {
	return filepath.Dir(effectiveConfigPath())
}
