package system

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/filter"
)

var toonCmd = &cobra.Command{
	Use:   "toon <file>",
	Short: "Apply TOON columnar encoding to JSON output",
	Long: `Compress JSON arrays using TOON columnar encoding.
Achieves 40-80% compression on structured data.

Examples:
  cat data.json | tok toon
  tok toon data.json`,
	RunE: runToon,
}

func runToon(cmd *cobra.Command, args []string) error {
	var input string
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		input = string(data)
	} else {
		data := make([]byte, 1024*1024)
		n, _ := os.Stdin.Read(data)
		input = string(data[:n])
	}

	encoder := filter.NewTOONEncoder(filter.DefaultTOONConfig())
	result, orig, comp, isToon := encoder.Encode(strings.TrimSpace(input))
	if !isToon {
		fmt.Print(input)
		return nil
	}

	fmt.Println(result)
	fmt.Fprintf(os.Stderr, "TOON: %d → %d tokens (%.1f%% saved)\n",
		orig, comp, float64(orig-comp)/float64(orig)*100)
	return nil
}

var tddCmd = &cobra.Command{
	Use:   "tdd <file>",
	Short: "Apply Token Dense Dialect encoding",
	Long: `Replace common programming terms with Unicode symbols
for 8-25% extra token savings.

Examples:
  cat code.go | tok tdd
  tok tdd code.go`,
	RunE: runTDD,
}

var tddDecode bool

func init() {
	registry.Add(func() { registry.Register(toonCmd) })
	registry.Add(func() { registry.Register(tddCmd) })

	tddCmd.Flags().BoolVar(&tddDecode, "decode", false, "Decode TDD back to original")
}

func runTDD(cmd *cobra.Command, args []string) error {
	var input string
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		input = string(data)
	} else {
		data := make([]byte, 1024*1024)
		n, _ := os.Stdin.Read(data)
		input = string(data[:n])
	}

	tdd := filter.NewTokenDenseDialect(filter.DefaultTDDConfig())

	if tddDecode {
		fmt.Print(tdd.Decode(input))
		return nil
	}

	result, stats := tdd.EncodeWithStats(input)
	fmt.Print(result)
	fmt.Fprintf(os.Stderr, "\n%s\n", filter.FormatTDDStats(stats))
	return nil
}
