# ADR-010: Absent TOML boolean treated as true (opt-out semantics)

## Status

Accepted

## Context

The config file supports `pull_requests` and `commits` boolean fields. Both default to enabled. When a user does not include these keys in their config file, the desired behaviour is that they remain enabled. The challenge: Go's TOML decoders (including `BurntSushi/toml`) decode a missing boolean field as `false`, the zero value for `bool`. This means an absent `pull_requests` key would silently disable pull requests — the opposite of the intended default.

Options:

1. **Pointer fields** (`*bool`): A missing key decodes as `nil`, distinguishable from `false`. Adds pointer indirection to the `Config` struct everywhere these fields are used.
2. **Double decode**: Decode once into the typed struct (for field parsing), then decode again into `map[string]interface{}` to check which keys are actually present in the file. Absent keys are then overridden to `true`.
3. **Opt-in semantics**: Change the default to disabled; users must explicitly set `pull_requests = true`. Breaks the principle of least surprise — running the tool out of the box should show everything.

## Decision

Decode the config file twice. The first decode populates the typed `Config` struct. The second decode populates a `map[string]interface{}`. For each boolean field whose key is absent from the map, the struct field is reset to `true`. Users who explicitly set `pull_requests = false` will find the key present in the map, so the explicit `false` is respected.

## Consequences

- The `Config` struct uses plain `bool` fields — no pointer indirection elsewhere in the codebase.
- The double-decode adds a negligible second file read at startup.
- The rule is clear: absent key = feature enabled. Explicit `false` = feature disabled. There is no third state.
- New boolean features added to the config should follow the same pattern: add the field, then add the absence-check in the same block in `Load`.
- The double-decode block is localised to `config.Load`; nothing else in the codebase needs to know about this behaviour.
