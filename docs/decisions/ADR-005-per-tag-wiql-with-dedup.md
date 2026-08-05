# ADR-005: One WIQL query per tag with client-side deduplication

## Status

Accepted

## Context

When the user configures multiple tags, the tool must return work items that match any of those tags. WIQL (Azure DevOps Work Item Query Language) supports `[System.Tags] CONTAINS '<tag>'` but does not have a direct `OR` operator for multiple `CONTAINS` conditions without nesting into a subquery. Options:

1. **Single query with OR subquery**: Technically possible but WIQL subquery syntax is poorly documented, version-sensitive, and fragile.
2. **Single query with IN clause**: `[System.Tags] IN (...)` is valid for exact equality, not for substring matching (`CONTAINS`). Tags in ADO are stored as a semicolon-delimited string, so `CONTAINS` is the correct operator.
3. **One WIQL per tag, client-side dedup**: Issue N queries (one per tag), collect all returned IDs, deduplicate on the client before the batch fetch.

## Decision

Issue one WIQL query per tag. Collect the resulting IDs in a `map[int]struct{}` keyed by work item ID to eliminate duplicates. Then fetch all unique IDs in a single batch request.

If no tags are configured, a single query is run with no tag condition.

## Consequences

- N API calls for N tags. In practice the number of tags is small (2–5), so the overhead is acceptable.
- The deduplication is exact: a work item matching multiple tags appears exactly once in the result.
- Each per-tag query returns the same set of fields and respects all other filters (sprint, assignee, type) — the tag condition is simply appended to the same condition list.
- The batch fetch (`workitemsbatch`) accepts an array of IDs, so there is no additional cost for deduplication happening client-side.
