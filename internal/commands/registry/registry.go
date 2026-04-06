package registry

import (
	"fmt"
	"log"
	"os"

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
				fmt.Fprintf(os.Stderr, "warning: command %q already registered, skipping (first-wins policy)\n", cmd.Name())
				log.Printf("DEBUG: duplicate command registration skipped: %q", cmd.Name())
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
