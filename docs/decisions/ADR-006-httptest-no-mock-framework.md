# ADR-006: httptest.NewServer over a mock framework

## Status

Accepted

## Context

The azdevops package makes HTTP calls that need to be tested without hitting real Azure DevOps endpoints. Options:

1. **Interface abstraction + mock**: Define an `HTTPClient` interface, inject a mock implementation in tests. Requires either a hand-written mock or a framework like `gomock` or `testify/mock`. The interface must be maintained as the real client evolves.
2. **`httptest.NewServer`**: Start a real HTTP server in the test process, point the client at it. No interface required; tests exercise the full HTTP client stack including JSON encoding, status code handling, and URL construction.
3. **Record/replay** (e.g. `go-vcr`): Record real API responses once, replay in tests. Adds recorded fixture files; fixtures go stale when API shape changes.

## Decision

Use `net/http/httptest.NewServer` directly. Each test file starts a mock server with a handler that switches on URL path to serve canned JSON responses. A shared `testhelpers_test.go` file provides `newMockServer` and `newClient` helpers used by all test files in the package.

## Consequences

- Zero additional test dependencies. `net/http/httptest` is part of the Go standard library.
- Tests exercise the real HTTP stack: JSON marshalling, status code branching, header handling, and URL construction are all covered.
- Mock server handlers are explicit Go code — they can assert on request content, return error status codes, and simulate partial failures.
- The production `Client` struct does not expose an interface; there is no seam for dependency injection. This is intentional — adding an interface for testability alone would be an abstraction with no production value.
- Shared helpers (`testhelpers_test.go`) prevent duplicated setup code across test files without leaking anything into the production binary (package `azdevops_test`).
