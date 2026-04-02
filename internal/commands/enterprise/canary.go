package enterprise

import (
	"context"
	"fmt"

	"github.com/GrayCodeAI/tokman/internal/canary"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/spf13/cobra"
)

func init() {
	canaryCmd := &cobra.Command{
		Use:   "canary",
		Short: "Manage canary deployments",
		Long:  "Create and manage canary deployments for gradual rollouts",
	}

	canaryCmd.AddCommand(&cobra.Command{
		Use:   "create [name]",
		Short: "Create a new canary deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			service, _ := cmd.Flags().GetString("service")
			targetVersion, _ := cmd.Flags().GetString("target")
			currentVersion, _ := cmd.Flags().GetString("current")

			manager := canary.NewManager()
			config := canary.DeploymentConfig{
				Name:           name,
				Service:        service,
				CurrentVersion: currentVersion,
				TargetVersion:  targetVersion,
				Strategy:       canary.StrategyStepped,
			}

			deploy, err := manager.CreateDeployment(config)
			if err != nil {
				return err
			}

			fmt.Printf("Created canary deployment: %s (ID: %s)\n", deploy.Name, deploy.ID)
			return nil
		},
	})

	canaryCmd.AddCommand(&cobra.Command{
		Use:   "start [deployment-id]",
		Short: "Start a canary deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := canary.NewManager()
			ctx := context.Background()

			deploy, err := manager.GetDeployment(args[0])
			if err != nil {
				return err
			}

			if err := deploy.Start(ctx); err != nil {
				return err
			}

			fmt.Printf("Started canary deployment: %s\n", deploy.Name)
			return nil
		},
	})

	canaryCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List canary deployments",
		Run: func(cmd *cobra.Command, args []string) {
			manager := canary.NewManager()
			deployments := manager.ListDeployments()

			fmt.Println("Canary Deployments:")
			for _, d := range deployments {
				fmt.Printf("  %s - %s (%s)\n", d.ID, d.Name, d.Status)
			}
		},
	})

	canaryCmd.Flags().String("service", "", "Service name")
	canaryCmd.Flags().String("target", "", "Target version")
	canaryCmd.Flags().String("current", "", "Current version")

	registry.Add(func() { registry.Register(canaryCmd) })
}
