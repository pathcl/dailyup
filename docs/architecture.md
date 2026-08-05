# Architecture

## Overview

`dailyup` is a CLI tool that queries Azure DevOps and emits a Markdown summary of
work done during a sprint or a rolling date window. It has no persistent state,
no database, and no background process — each invocation does its work and exits.

---

## Data Flow

```
User invokes: dailyup summary [flags]
        │
        ▼
cmd/root.go          Load config (TOML), merge CLI flags, build query options
        │
        ├─── goroutine ──► azdevops.FetchWorkItems
        │                       │
        │                       ├── runWIQL  (POST /_apis/wit/wiql)
        │                       │       Returns list of work item IDs
        │                       │
        │                       ├── fetchBatch  (POST /_apis/wit/workitemsbatch)
        │                       │       Fetches fields for all IDs at once
        │                       │
        │                       └── fetchIterationCounts  [concurrent, one goroutine per item]
        │                               GET /_apis/wit/workitems/{id}/revisions
        │                               Returns SprintCount per item
        │
        ├─── goroutine ──► azdevops.FetchPullRequests
        │                       GET /_apis/git/pullrequests
        │                       Client-side date filter applied
        │
        └─── goroutine ──► azdevops.FetchCommits
                                GET /_apis/git/repositories          (list all repos)
                                GET /_apis/git/repositories/{id}/commits  (per repo, sequential)
                                Client-side author filter applied
        │
        ▼
  sync.WaitGroup.Wait() — collect results, surface first error
        │
        ▼
report.Render(title, workItems, prs, commits)
        │
        ▼
stdout  (Markdown)       stderr  (slog structured logs, Warn level by default)
```

---

## Package Responsibilities

### `main`
Entry point. Delegates immediately to `cmd.Execute()`. Contains no logic.

### `cmd`
Cobra command wiring. Two subcommands:
- **`summary`**: Orchestrates concurrent data fetching and Markdown rendering.
- **`types`**: Lists available work item types in the configured project.

Owns all flag definitions and the flag-to-config override logic. Responsible for building `WorkItemOpts` from the active filter mode (sprint vs. date). See [ADR-002](decisions/ADR-002-sprint-first-filtering.md).

### `internal/config`
Loads and validates `~/.config/dailyup/config.toml`. Performs tilde expansion and boolean defaulting. See [ADR-010](decisions/ADR-010-toml-bool-absent-equals-true.md).

### `internal/azdevops`
All Azure DevOps API calls. Structured as one package, four files:

| File | Responsibility |
|------|---------------|
| `client.go` | `*Client` struct, auth via Azure CLI, composable HTTP transports |
| `workitems.go` | WIQL queries, batch item fetch, filter construction |
| `pullrequests.go` | PR list fetch, date-window filter, dual date-format parsing |
| `commits.go` | Repo enumeration, per-repo commit fetch, author filter |
| `revisions.go` | Concurrent revision history fetch for sprint carry-over count |
| `workitemtypes.go` | Catalog of available work item types |

See [ADR-003](decisions/ADR-003-flat-azdevops-package.md).

### `internal/report`
Stateless Markdown renderer. Takes pre-fetched data, groups work items by type,
and writes to a `strings.Builder`. No knowledge of query mode or dates. See [ADR-009](decisions/ADR-009-report-title-as-string.md).

---

## Authentication

The tool uses `azidentity.NewAzureCLICredential` to obtain a short-lived Bearer
token from an existing `az login` session. The token is scoped to the Azure
DevOps resource (`499b84ac-1321-427f-aa17-267ca6975798`). No PAT is stored or
required. See [ADR-001](decisions/ADR-001-az-cli-auth.md).

---

## API Version

All Azure DevOps REST endpoints use `api-version=7.1`. This is the current
stable version as of the tool's creation. The version is embedded in each URL
individually (not a global header) so individual endpoints can be upgraded
without a flag day.

---

## Key Design Decisions

| Decision | ADR |
|----------|-----|
| Use Azure CLI auth, no PAT | [ADR-001](decisions/ADR-001-az-cli-auth.md) |
| Sprint as primary filter, --weeks as fallback | [ADR-002](decisions/ADR-002-sprint-first-filtering.md) |
| Flat package layout (workitems / pullrequests / commits) | [ADR-003](decisions/ADR-003-flat-azdevops-package.md) |
| Concurrent top-level fetches via goroutines + WaitGroup | [ADR-004](decisions/ADR-004-concurrent-fetching.md) |
| One WIQL per tag with client-side dedup | [ADR-005](decisions/ADR-005-per-tag-wiql-with-dedup.md) |
| httptest.NewServer for tests, no mock framework | [ADR-006](decisions/ADR-006-httptest-no-mock-framework.md) |
| log/slog over zerolog/zap | [ADR-007](decisions/ADR-007-log-slog.md) |
| Revision API for sprint carry-over count | [ADR-008](decisions/ADR-008-revision-api-for-sprint-count.md) |
| Report title as pre-formatted string | [ADR-009](decisions/ADR-009-report-title-as-string.md) |
| Absent TOML boolean = true (opt-out semantics) | [ADR-010](decisions/ADR-010-toml-bool-absent-equals-true.md) |

---

## Extension Points

### Adding a new data source (e.g. pipelines)
1. Add `pipelines.go` in `internal/azdevops` with a `FetchPipelines` function.
2. Add the corresponding type and method to the `report` package or extend `Render`.
3. Add a `Pipelines bool` field to `Config` with the absent-equals-true pattern.
4. Wire a new goroutine in `cmd/root.go` alongside the existing three.
5. Add tests in `pipelines_test.go` using `newMockServer` and `newClient`.

### Adding a new output format (e.g. JSON)
1. Add `json.go` in `internal/report`.
2. Add a `--format` flag in `cmd/root.go`; switch on it before calling `Render`.
3. No changes to the azdevops package are required.

### Adding a new filter (e.g. area path)
1. Add the field to `WorkItemOpts` in `workitems.go`.
2. Add the WIQL condition in `buildQuery`.
3. Add the flag in `cmd/root.go` and wire it into `wiOpts`.
