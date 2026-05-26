package cmd

import (
	"fmt"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/spf13/cobra"
)

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "List all work item types available in the configured project",
	RunE:  runTypes,
}

func init() {
	typesCmd.Flags().StringVar(&cfgPath, "config", config.DefaultPath(), "path to config file")
	typesCmd.Flags().BoolVar(&debug, "debug", false, "print raw HTTP requests and responses to stderr")
	rootCmd.AddCommand(typesCmd)
}

func runTypes(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	client, err := azdevops.NewClientFromAzCLI(cfg.Organization, cfg.Project, debug)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	types, err := azdevops.FetchWorkItemTypes(client)
	if err != nil {
		return err
	}
	fmt.Printf("Work item types in %s/%s:\n\n", cfg.Organization, cfg.Project)
	for _, t := range types {
		if t.Description != "" {
			fmt.Printf("  %-25s %s\n", t.Name, t.Description)
		} else {
			fmt.Printf("  %s\n", t.Name)
		}
	}
	return nil
}
