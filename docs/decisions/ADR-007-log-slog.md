# ADR-007: log/slog over zerolog or zap

## Status

Accepted

## Context

The tool benefits from structured logging to aid debugging of API calls, authentication failures, and query construction. Options:

1. **`log` (stdlib)**: Unstructured. Cannot set levels programmatically. Insufficient for attaching key-value pairs to log lines.
2. **`log/slog` (stdlib, Go 1.21+)**: Structured, levelled, zero-dependency. API is intentionally similar to `log` with added key-value support.
3. **`zerolog`**: Zero-allocation structured logger. Widely used; adds a dependency.
4. **`go.uber.org/zap`**: High-performance structured logger. Adds a dependency; API is more complex (fields must be typed explicitly).

## Decision

Use `log/slog`. The default log level is `Warn` so normal usage is silent. Passing `--debug` switches the level to `Debug`, which logs every HTTP request URL, response status code, and response body to stderr.

## Consequences

- No additional dependency. The module stays lean.
- `slog` is the stdlib's long-term answer to structured logging; it will be maintained and improved as part of the Go release cycle.
- Performance is not a concern for a CLI tool that makes tens of API calls per run — the allocation overhead of `zerolog` or `zap` would be imperceptible.
- All log output goes to stderr; stdout is reserved for the Markdown report, so the output can be piped directly to a file or another tool without mixing log noise into the report.
- The `--debug` flag wires directly to a `slog.LevelDebug` handler; there is no config-file equivalent (debug is a developer/troubleshooting concern, not a persistent setting).
