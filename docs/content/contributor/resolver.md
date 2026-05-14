---
title: Resolver
sidebar_position: 8
---

# Resolver

How `glix add <ref>` turns a possibly-short user input into a canonical
flake reference.

Location: `glix/internal/resolver/`.

See [ADR-004](../adr/ADR-004-resolver-chain).

## The chain

```
input string
  │
  ├── looks like a flake URI?
  │     yes → SourceURI, ref=input
  │     no  → continue
  │
  ├── present in glixos registry?
  │     yes → SourceRegistry, ref=registry[input].flake
  │     no  → continue
  │
  ├── present in `nix registry list`?
  │     yes → SourceNixRegistry, ref=<resolved>
  │     no  → return error with suggestions
```

## What counts as a URI

`looksLikeURI(s)` returns true if `s` contains a `:` whose left side is a
recognised flake scheme (`github`, `gitlab`, `sourcehut`, `srht`,
`path`, `git`, `git+ssh`, `git+http`, `git+https`, `http`, `https`,
`file`, `tarball`, `flake`) **or** if `s` contains a `#` (the
`pkg.foo#attr` form). Anything else is treated as a short name.

## `Options`

```go
type Options struct {
    Registry    *registry.Registry
    NixRegistry NixRegistryAccessor
}
```

- `Registry` is loaded eagerly by the caller (see `registry_helpers.go`).
- `NixRegistry` is an interface so tests can swap in a fake. Concrete
  implementations:
  - `NixCLI{}` shells out to `nix registry list`.
  - `NoNixRegistry{}` always returns "not found"; activated by
    `glix add --no-nix-registry`.

## Suggestions on miss

If nothing resolves, the error includes up to N close matches from the
registry's package names (basic Levenshtein-style ranking). The bar is
"helpful when you typo a known name", not "smart fuzzy search".

## Tests

`glix/internal/resolver/resolver_test.go` covers:

- URI passthrough for each recognised scheme.
- Registry hit with and without a trailing slash.
- `NoNixRegistry` fallback path.
- "not found" error with suggestions.

The `nix registry list` integration is mocked via the `NixRegistryAccessor`
interface; we don't shell out in unit tests.
