package cmd

import (
	"fmt"
	"strings"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/spf13/cobra"
)

var (
	viewID    int
	viewDebug bool
	viewCfg   string
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Print all details of a work item",
	RunE:  runView,
}

func init() {
	viewCmd.Flags().IntVar(&viewID, "id", 0, "work item ID")
	viewCmd.Flags().BoolVar(&viewDebug, "debug", false, "print raw HTTP requests and responses to stderr")
	viewCmd.Flags().StringVar(&viewCfg, "config", config.DefaultPath(), "path to config file")
	_ = viewCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(viewCmd)
}

func runView(cmd *cobra.Command, args []string) error {
	setupLogger(viewDebug)

	cfg, err := config.Load(viewCfg)
	if err != nil {
		return err
	}

	client, err := azdevops.NewClientFromAzCLI(cfg.Organization, cfg.Project, viewDebug)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	item, err := azdevops.FetchWorkItemDetail(client, viewID)
	if err != nil {
		return err
	}

	fmt.Print(RenderDetail(item))
	return nil
}

func RenderDetail(item *azdevops.WorkItemDetail) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "#%d — %s\n", item.ID, item.Title)
	fmt.Fprintf(&sb, "%s\n\n", strings.Repeat("─", 60))

	field := func(label, value string) {
		if value != "" {
			fmt.Fprintf(&sb, "%-16s %s\n", label+":", value)
		}
	}

	field("Type", item.Type)
	field("State", item.State)
	field("Priority", formatPriority(item.Priority))
	field("Assigned To", item.AssignedTo)
	field("Created By", item.CreatedBy)
	field("Created", formatDate(item.CreatedDate))
	field("Last Updated", formatDate(item.ChangedDate))
	field("Area", item.AreaPath)
	field("Sprint", item.IterationPath)
	field("Tags", item.Tags)

	if item.Description != "" {
		fmt.Fprintf(&sb, "\n%s\n%s\n", strings.Repeat("─", 60), "Description:")
		fmt.Fprintf(&sb, "%s\n", stripHTML(item.Description))
	}

	return sb.String()
}

func formatPriority(p int) string {
	if p == 0 {
		return ""
	}
	return fmt.Sprintf("%d", p)
}

func formatDate(s string) string {
	// ADO returns RFC3339-ish: "2026-05-01T10:30:00Z" — trim to date only.
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// stripHTML removes HTML tags from ADO description fields.
func stripHTML(s string) string {
	var out strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			out.WriteRune(r)
		}
	}
	return strings.TrimSpace(out.String())
}
