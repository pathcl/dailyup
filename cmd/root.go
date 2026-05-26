package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pathcl/dailyup/internal/azdevops"
	"github.com/pathcl/dailyup/internal/config"
	"github.com/pathcl/dailyup/internal/report"
	"github.com/spf13/cobra"
)

var (
	cfgPath        string
	weeks          int
	sprintName     string
	assignedTo     string
	debug          bool
	noPullRequests bool
	noCommits      bool
)

var rootCmd = &cobra.Command{
	Use:   "dailyup",
	Short: "Summarise your Azure DevOps activity for the current (or a named) sprint",
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Generate a Markdown work summary",
	RunE:  runSummary,
}

func init() {
	summaryCmd.Flags().StringVar(&cfgPath, "config", config.DefaultPath(), "path to config file")
	summaryCmd.Flags().IntVar(&weeks, "weeks", 0, "look-back window for PRs/commits (overrides config)")
	summaryCmd.Flags().StringVar(&sprintName, "sprint", "", `sprint name, e.g. "Sprint 68" (default: current sprint)`)
	summaryCmd.Flags().StringVar(&assignedTo, "assigned-to", "", `filter work items by person, e.g. "@Me" or display name (overrides config)`)
	summaryCmd.Flags().BoolVar(&debug, "debug", false, "print raw HTTP requests and responses to stderr")
	summaryCmd.Flags().BoolVar(&noPullRequests, "no-pull-requests", false, "skip fetching pull requests")
	summaryCmd.Flags().BoolVar(&noCommits, "no-commits", false, "skip fetching commits")
	rootCmd.AddCommand(summaryCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runSummary(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	if weeks > 0 {
		cfg.Weeks = weeks
	}
	// CLI flags override config
	if assignedTo != "" {
		cfg.AssignedTo = assignedTo
	}
	if noPullRequests {
		cfg.PullRequests = false
	}
	if noCommits {
		cfg.Commits = false
	}

	since := time.Now().UTC().Add(-time.Duration(cfg.Weeks) * 7 * 24 * time.Hour)
	to := time.Now().UTC()

	client, err := azdevops.NewClientFromAzCLI(cfg.Organization, cfg.Project, debug)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	wiOpts := azdevops.WorkItemOpts{
		Sprint:     sprintName,
		Tags:       cfg.Tags,
		AssignedTo: cfg.AssignedTo,
	}

	var (
		wg                  sync.WaitGroup
		workItems           []azdevops.WorkItem
		prs                 []azdevops.PullRequest
		commits             []azdevops.Commit
		wiErr, prErr, cmErr error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		workItems, wiErr = azdevops.FetchWorkItems(client, wiOpts)
	}()

	if cfg.PullRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prs, prErr = azdevops.FetchPullRequests(client, since)
		}()
	}

	if cfg.Commits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			commits, cmErr = azdevops.FetchCommits(client, cfg.Email, since)
		}()
	}

	wg.Wait()

	for _, e := range []error{wiErr, prErr, cmErr} {
		if e != nil {
			return e
		}
	}

	fmt.Print(report.Render(since, to, workItems, prs, commits))
	return nil
}
