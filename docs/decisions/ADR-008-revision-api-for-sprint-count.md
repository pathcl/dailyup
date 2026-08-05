# ADR-008: Per-item revision history for sprint carry-over count

## Status

Accepted

## Context

Users want to know if a work item has been carried across multiple sprints — a signal that it is blocked, under-estimated, or neglected. Azure DevOps does not expose a "sprint count" or "carry-over count" field on work items. Options for computing it:

1. **One WIQL per item per sprint**: Query whether item ID X was ever in Sprint N, Sprint N-1, etc. Produces O(items × sprints) API calls.
2. **Revision history API**: `GET /_apis/wit/workitems/{id}/revisions?$select=System.IterationPath` returns all historical revisions of an item. Count the distinct `System.IterationPath` values to get the number of sprints the item has been in.
3. **Accept the limitation**: Do not show sprint count. Users lose carry-over visibility.

## Decision

Use the revision history API (option 2). After the batch work item fetch, issue one revision request per item concurrently using goroutines and `sync.WaitGroup`. The result — a `map[int]int` of item ID to distinct sprint count — is merged back into each `WorkItem` as the `SprintCount` field. The report only surfaces `SprintCount` when it is greater than 1 ("carried over N sprints").

## Consequences

- The revision fetch adds N concurrent HTTP calls (one per work item). For typical sprint sizes (10–30 items) this is fast; for large queries it may take a few extra seconds.
- Goroutine-per-item is safe at this scale; ADO does not enforce aggressive rate limiting for this endpoint.
- Errors fetching revisions for any single item are logged as warnings and treated as `SprintCount = 0` (the item still appears in the report, just without the carry-over annotation).
- The revision API returns all historical revisions including field changes unrelated to the iteration path. Only the `System.IterationPath` field is requested via `$select`, minimising payload size.
- New items created in the current sprint show `SprintCount = 1`, which is correct — they have been in exactly one sprint.
