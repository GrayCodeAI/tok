package configcmd

import (
	"path/filepath"

	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/config"
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
