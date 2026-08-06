package cmd

import (
	"fmt"
	"strings"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/pathcl/dailyup/internal/editor"
	"github.com/spf13/cobra"
)

var (
	createParent int
	createArea   string
	createSprint string
	createTags   string
	createTask   bool
	createDebug  bool
	createCfg    string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new work item by opening your editor",
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().IntVar(&createParent, "parent", 0, "parent work item ID")
	createCmd.Flags().StringVar(&createArea, "area", "", "area path (overrides config default)")
	createCmd.Flags().StringVar(&createSprint, "sprint", "", "iteration path (overrides config default)")
	createCmd.Flags().StringVar(&createTags, "tags", "", "comma-separated tags to apply, e.g. \"backend,api\"")
	createCmd.Flags().BoolVar(&createTask, "task", false, "create a Task instead of a User Story")
	createCmd.Flags().BoolVar(&createDebug, "debug", false, "print raw HTTP requests and responses to stderr")
	createCmd.Flags().StringVar(&createCfg, "config", config.DefaultPath(), "path to config file")
	_ = createCmd.MarkFlagRequired("parent")
	rootCmd.AddCommand(createCmd)
}

const createTemplate = `Title:

Description:

# Enter the title on the 'Title:' line (required).
# Add an optional description below 'Description:'.
# Lines starting with '#' are ignored.
`

func runCreate(cmd *cobra.Command, args []string) error {
	setupLogger(createDebug)

	cfg, err := config.Load(createCfg)
	if err != nil {
		return err
	}

	area := cfg.Area
	if createArea != "" {
		area = createArea
	}
	sprint := cfg.Sprint
	if createSprint != "" {
		sprint = createSprint
	}
	if area == "" {
		return fmt.Errorf("area path required: set 'area' in config or pass --area")
	}
	if sprint == "" {
		return fmt.Errorf("sprint required: set 'sprint' in config or pass --sprint")
	}

	itemType := "User Story"
	if createTask {
		itemType = "Task"
	}

	content, err := editor.Open(createTemplate)
	if err != nil {
		return fmt.Errorf("editor: %w", err)
	}

	title, description, err := ParseCreateContent(content)
	if err != nil {
		return err
	}

	client, err := azdevops.NewClientFromAzCLI(cfg.Organization, cfg.Project, createDebug)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	newID, err := azdevops.CreateNewWorkItem(client, itemType, title, description, createTags, area, sprint, createParent)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	fmt.Printf("Created #%d %q (%s) under parent #%d\n", newID, title, itemType, createParent)
	return nil
}

// ParseCreateContent extracts title and description from the editor output.
func ParseCreateContent(content string) (title, description string, err error) {
	lines := strings.Split(content, "\n")
	var descLines []string
	inDesc := false

	for _, line := range lines {
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			continue
		}
		if strings.HasPrefix(line, "Description:") {
			inDesc = true
			rest := strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			if rest != "" {
				descLines = append(descLines, rest)
			}
			continue
		}
		if inDesc {
			descLines = append(descLines, line)
		}
	}

	if title == "" {
		return "", "", fmt.Errorf("title is required — add text after 'Title:'")
	}
	description = strings.TrimSpace(strings.Join(descLines, "\n"))
	return title, description, nil
}
