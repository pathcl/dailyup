package cmd

import (
	"fmt"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/spf13/cobra"
)

var (
	copyIDs    []int
	copyParent int
	toArea     string
	toSprint   string
	copyDryRun bool
	copyDebug  bool
	copyCfg    string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy work items to a different area and sprint",
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().IntSliceVar(&copyIDs, "id", nil, "work item ID(s) to copy (repeatable)")
	copyCmd.Flags().StringVar(&toArea, "to-area", "", "target area path, e.g. \"MyProject\\Area B\"")
	copyCmd.Flags().StringVar(&toSprint, "to-sprint", "", "target iteration path, e.g. \"Team B\\Iteration 2\"")
	copyCmd.Flags().IntVar(&copyParent, "parent", 0, "parent work item ID to link copied items under")
	copyCmd.Flags().BoolVar(&copyDryRun, "dry-run", false, "print what would be created without writing")
	copyCmd.Flags().BoolVar(&copyDebug, "debug", false, "print raw HTTP requests and responses to stderr")
	copyCmd.Flags().StringVar(&copyCfg, "config", config.DefaultPath(), "path to config file")
	_ = copyCmd.MarkFlagRequired("id")
	_ = copyCmd.MarkFlagRequired("to-area")
	_ = copyCmd.MarkFlagRequired("to-sprint")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	setupLogger(copyDebug)

	cfg, err := config.Load(copyCfg)
	if err != nil {
		return err
	}

	client, err := azdevops.NewClientFromAzCLI(cfg.Organization, cfg.Project, copyDebug)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	items, err := azdevops.FetchWorkItemsByIDs(client, copyIDs)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	for _, item := range items {
		if copyDryRun {
			fmt.Printf("[dry-run] Would copy #%d %q (%s) → %s / %s\n",
				item.ID, item.Title, item.Type, toArea, toSprint)
			continue
		}
		newID, err := azdevops.CreateWorkItem(client, item, toArea, toSprint, copyParent)
		if err != nil {
			return fmt.Errorf("copy #%d: %w", item.ID, err)
		}
		fmt.Printf("Copied #%d → #%d %q (%s)\n", item.ID, newID, item.Title, item.Type)
	}
	return nil
}
