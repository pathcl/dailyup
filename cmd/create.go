package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/pathcl/dailyup/internal/editor"
	"github.com/spf13/cobra"
)

var (
	createParent   int
	createArea     string
	createSprint   string
	createTags     string
	createTask     bool
	createType     string
	createTemplate string
	createDebug    bool
	createCfg      string
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
	createCmd.Flags().BoolVar(&createTask, "task", false, "create a Task (shorthand for --type task)")
	createCmd.Flags().StringVar(&createType, "type", "", "work item type: story (default), task, or feature")
	createCmd.Flags().StringVar(&createTemplate, "template", "", "path to a custom editor template file")
	createCmd.Flags().BoolVar(&createDebug, "debug", false, "print raw HTTP requests and responses to stderr")
	createCmd.Flags().StringVar(&createCfg, "config", config.DefaultPath(), "path to config file")
	_ = createCmd.MarkFlagRequired("parent")
	rootCmd.AddCommand(createCmd)
}

// defaultTemplates holds the hardcoded editor template for each ADO item type.
var defaultTemplates = map[string]string{
	"User Story": `Title:

Description:
## Story
As a [platform engineer / on-call / service team],
I want [capability],
so that [outcome].

## Context
Why this sprint, why this priority.

## Acceptance Criteria
- [ ] Given [state], when [trigger], then [observable result]
- [ ] Alert fires within [X]s of the condition
- [ ] No false positives in staging for [N] hours
- [ ] Runbook covers this scenario

## Technical Approach
[Key decisions, constraints, ADRs referenced]

## Definition of Done
- [ ] Code reviewed & merged
- [ ] Tests pass
- [ ] Deployed to prod
- [ ] Error budget unaffected (verify dashboard)
- [ ] On-call notified if behavior changes

# Lines starting with '#' are ignored.
`,
	"Task": `Title:

Description:
## What
Single paragraph. What is being done and why.

## Steps
1. [Step]
2. [Step]

## Done When
- [ ] [Verifiable condition]
- [ ] PR merged / runbook updated / config deployed

## Notes
[Links, gotchas — delete if empty]

# Lines starting with '#' are ignored.
`,
	"Feature": `Title:

Description:
## Objective
One paragraph: what this delivers and why it matters now.

## Success Metrics
- Availability target: [e.g. 99.9%]
- Error budget allocated: [X minutes/month]
- Latency target: [p99 < Xms]
- Leading indicator to watch: [metric name]

## Scope
**In:** [what's explicitly included]
**Out:** [what's explicitly excluded]

## Dependencies
- [ ] [Team / service] — [what we need]

## Risks
| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| [risk] | High/Med/Low | [action] |

## Definition of Done
- [ ] Runbook written and linked
- [ ] Dashboards updated
- [ ] Alerts tuned (signal:noise acceptable)
- [ ] Post-rollout review scheduled

# Lines starting with '#' are ignored.
`,
}

// ItemTypeFromFlag maps --type and --task flag values to the ADO work item type
// string. --type takes precedence over --task when both are set.
func ItemTypeFromFlag(typeFlag string, taskFlag bool) (string, error) {
	if typeFlag != "" {
		switch strings.ToLower(typeFlag) {
		case "story":
			return "User Story", nil
		case "task":
			return "Task", nil
		case "feature":
			return "Feature", nil
		default:
			return "", fmt.Errorf("unknown type %q: use story, task, or feature", typeFlag)
		}
	}
	if taskFlag {
		return "Task", nil
	}
	return "User Story", nil
}

// LoadTemplate returns the editor template content for the given item type.
// Lookup order: customPath → templatesDir/<type>.md → hardcoded default.
// typeFlag is the flag value (story/task/feature); templatesDir is the
// directory to search (pass config.TemplatesDir() in production).
func LoadTemplate(typeFlag, customPath, templatesDir string) string {
	if customPath != "" {
		if b, err := os.ReadFile(customPath); err == nil {
			return string(b)
		}
	}
	if templatesDir != "" {
		path := filepath.Join(templatesDir, strings.ToLower(typeFlag)+".md")
		if b, err := os.ReadFile(path); err == nil {
			return string(b)
		}
	}
	// Map flag value to ADO type to look up default.
	adoType, _ := ItemTypeFromFlag(typeFlag, false)
	if t, ok := defaultTemplates[adoType]; ok {
		return t
	}
	return defaultTemplates["User Story"]
}

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
	sprintLeaf := cfg.Sprint
	if createSprint != "" {
		sprintLeaf = createSprint
	}
	iterationBase := cfg.IterationBase

	if area == "" {
		return fmt.Errorf("area path required: set 'area' in config or pass --area")
	}

	iterationPath := config.BuildIterationPath(iterationBase, sprintLeaf)
	if iterationPath == "" {
		return fmt.Errorf("sprint required: set 'sprint' in config or pass --sprint")
	}

	itemType, err := ItemTypeFromFlag(createType, createTask)
	if err != nil {
		return err
	}

	typeFlag := strings.ToLower(createType)
	if typeFlag == "" {
		if createTask {
			typeFlag = "task"
		} else {
			typeFlag = "story"
		}
	}
	tmpl := LoadTemplate(typeFlag, createTemplate, config.TemplatesDir())

	content, err := editor.Open(tmpl)
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

	newID, err := azdevops.CreateNewWorkItem(client, itemType, title, description, createTags, area, iterationPath, createParent)
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
		if strings.HasPrefix(line, "#") {
			continue
		}
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
