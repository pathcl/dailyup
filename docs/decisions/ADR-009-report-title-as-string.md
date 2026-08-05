# ADR-009: Report title passed as a pre-formatted string

## Status

Accepted

## Context

The Markdown report heading describes what was queried. The heading content depends on which mode was used:

- One sprint: `"Sprint 68"`
- Multiple sprints: `"Sprint 67, Sprint 68"`
- Date range: `"Apr 28, 2026 – May 26, 2026"`
- Current iteration: `"Current Sprint"`

Options for the `Render` function signature:

1. **`Render(from, to time.Time, sprints []string, ...)`**: The report package decides how to format the heading based on mode. The caller must still communicate the mode, making the signature complex.
2. **`Render(title string, ...)`**: The caller formats the heading and passes it as a string. The report package treats it verbatim.

## Decision

Use `Render(title string, items []WorkItem, prs []PullRequest, commits []Commit) string`. The caller (`cmd/root.go`) builds the title string based on the active filter mode before calling `Render`.

## Consequences

- The report package has no knowledge of sprint names, date ranges, or query modes. It does one thing: render data as Markdown.
- The title string is the single source of truth for the heading. If the heading is wrong, the bug is in the caller, not the renderer.
- Tests for the report package can pass any string as the title without caring about the upstream query logic.
- Adding a new query mode in the future (e.g. label-based) requires only a new `title =` assignment in the caller; the report package needs no change.
