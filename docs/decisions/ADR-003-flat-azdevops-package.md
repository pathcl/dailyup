# ADR-003: Flat azdevops package with three focused files

## Status

Accepted

## Context

The Azure DevOps data layer needs to query three distinct resource types: work items (via WIQL + batch), pull requests, and commits. Options for structuring this code:

1. **One monolithic file**: Everything in `azdevops.go`. Simple to navigate for small codebases; grows unwieldy as each resource type has its own request/response structs, query logic, and tests.
2. **Sub-packages** (`azdevops/workitems`, `azdevops/pullrequests`, `azdevops/commits`): Maximum isolation, but creates cross-import complexity. The `*Client` struct would need to live in a shared base package, adding indirection.
3. **Flat package, multiple files**: One package (`azdevops`), files split by resource type (`workitems.go`, `pullrequests.go`, `commits.go`, `revisions.go`). Each file is independently readable; all share the same `*Client` without any extra import.

## Decision

Use option 3: a single `azdevops` package with four focused files. Each file owns its resource type's structs, API calls, and helper functions. All files share the `*Client` declared in `client.go`.

## Consequences

- Each file can be read and understood in isolation without chasing imports.
- Tests are co-located: `workitems_test.go`, `pullrequests_test.go`, `commits_test.go`, `revisions_test.go`.
- Adding a new resource type (e.g. pipelines) means adding a new file in the same package — no structural change required.
- The package surface area is flat; callers import a single path (`internal/azdevops`) regardless of which resource they use.
- Test helpers shared across files live in `testhelpers_test.go` (package `azdevops_test`), avoiding duplication while keeping helpers out of the production binary.
