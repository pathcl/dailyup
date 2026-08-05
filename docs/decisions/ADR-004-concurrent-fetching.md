# ADR-004: Concurrent data fetching with goroutines and WaitGroup

## Status

Accepted

## Context

A full `dailyup summary` run makes three independent top-level API calls: work items, pull requests, and commits. Each involves at least one HTTP round-trip; work items also fan out into per-item revision history calls. Running them sequentially means total latency equals the sum of all calls.

Options:

1. **Sequential**: Simple, easy to read. Total latency = sum of all call durations.
2. **Goroutines + `sync.WaitGroup` in the caller** (`cmd/root.go`): Each top-level fetch runs in its own goroutine. Total latency ≈ max of all call durations.
3. **Channel pipeline**: More composable but adds complexity with no benefit here since all results must be collected before rendering.
4. **`errgroup`**: Cleaner error propagation but adds a dependency (`golang.org/x/sync`) for a problem that the stdlib already solves adequately.

## Decision

Use goroutines and `sync.WaitGroup` in `cmd/root.go` to run the three top-level fetches concurrently. Each goroutine captures its error into a dedicated variable; errors are checked after `wg.Wait()`. Concurrency lives at the caller layer; the azdevops functions themselves are synchronous.

Within `revisions.go`, the same pattern is applied for per-item revision history: one goroutine per item, a mutex-protected result map, and a single `WaitGroup`.

## Consequences

- Total latency is bounded by the slowest individual call rather than the sum of all calls.
- No additional dependencies beyond stdlib.
- Error handling is explicit: all three errors are checked after the wait, so the first non-nil error is returned.
- The azdevops functions remain simple and synchronous — concurrency is an orchestration concern, not an implementation concern.
- If one section (e.g. PRs) is disabled via config, its goroutine is never started; `cfg.PullRequests = false` skips the `wg.Add(1)` entirely.
