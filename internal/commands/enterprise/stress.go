package enterprise

import (
	"context"
	"fmt"
	"time"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/stress"
	"github.com/spf13/cobra"
)

func init() {
	stressCmd := &cobra.Command{
		Use:   "stress",
		Short: "Run stress tests",
		Long:  "Run stress tests to evaluate system performance under load",
	}

	stressCmd.AddCommand(&cobra.Command{
		Use:   "run [scenario]",
		Short: "Run a stress test scenario",
		RunE: func(cmd *cobra.Command, args []string) error {
			scenarioName := "basic_load"
			if len(args) > 0 {
				scenarioName = args[0]
			}

			config := stress.DefaultConfig()
			config.Duration = 5 * time.Minute

			runner := stress.NewRunner(config)

			// Register standard scenarios
			for _, s := range stress.StandardScenarios() {
				runner.RegisterScenario(s)
			}

			ctx := context.Background()
			result, err := runner.Run(ctx, scenarioName)
			if err != nil {
				return err
			}

			fmt.Println(result.GenerateReport())

			return nil
		},
	})

	stressCmd.AddCommand(&cobra.Command{
		Use:   "scenarios",
		Short: "List available stress test scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available stress test scenarios:")
			fmt.Println("  basic_load  - Basic load test with steady traffic")
			fmt.Println("  spike_test  - Sudden traffic spike simulation")
			fmt.Println("  soak_test   - Extended duration stability test")
		},
	})

	registry.Add(func() { registry.Register(stressCmd) })
}
