# ADR-002: Sprint (iteration path) as primary filter, date range as fallback

## Status

Accepted

## Context

Work items in Azure DevOps are organised by iteration path (sprint), not by calendar date. A work item assigned to "Sprint 68" stays in that iteration regardless of when it was created or modified. Querying by `ChangedDate` therefore crosses sprint boundaries and produces ambiguous results — an item modified during sprint review to close it appears in the next sprint's date window.

Two filtering strategies are possible:

- **Sprint-first**: Use `System.IterationPath UNDER '<project>\<sprint>'` (single sprint) or `System.IterationPath IN (...)` (multiple). Semantically clean; matches what the user sees in the ADO board.
- **Date-first**: Use `System.ChangedDate >= '<date>'`. Simpler for open-ended time ranges; the only option for PRs and commits, which have no iteration path.

Pull requests and commits do not have iteration paths — they can only be queried by date.

## Decision

Work items default to the current iteration (`@CurrentIteration`). Users can specify one or more sprint names via `--sprint`, which maps to an iteration path query. The `--weeks` flag is a date-based fallback for work items when sprint names are not given, and the primary filter for PRs and commits in all modes.

Sprint takes priority over weeks: if both are specified, sprint wins for work items.

## Consequences

- Sprint names must match the iteration path exactly as configured in ADO (everything after the project prefix, e.g. `Sprint 68` not `Project\Sprint 68`).
- The report heading reflects what was actually queried: sprint names joined by comma, a date range, or "Current Sprint".
- Multi-sprint queries use `IN` instead of `UNDER` because `UNDER` only accepts a single path.
- PR and commit sections always use the date window derived from `--weeks` (or the 2-week default), independent of which sprint filter is active for work items.
