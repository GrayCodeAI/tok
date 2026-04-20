package registry

import (
	"os"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"
)

var globalRoot *cobra.Command

func Init(root *cobra.Command) {
	globalRoot = root
}

func Register(cmd *cobra.Command) {
	if globalRoot != nil {
		for _, existing := range globalRoot.Commands() {
			if existing.Name() == cmd.Name() {
				// Keep first registration deterministically and ignore subsequent duplicates.
				// Avoid noisy startup warnings when optional command packs overlap.
				if os.Getenv("TOK_DEBUG_REGISTRY") == "1" {
					out.Global().Errorf("registry debug: duplicate command %q skipped\n", cmd.Name())
				}
				return
			}
		}
		globalRoot.AddCommand(cmd)
	}
}

type RegisterFunc func()

var AllPackages []RegisterFunc

func Add(fn RegisterFunc) {
	AllPackages = append(AllPackages, fn)
}

func RegisterAll() {
	for _, fn := range AllPackages {
		fn()
	}
}
