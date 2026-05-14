---
title: Anchored-region patcher
sidebar_position: 7
---

# Anchored-region patcher

How glix writes to `flake.nix` without parsing Nix.

Location: `glix/internal/flake/patcher.go`.

## The anchors

```nix
# >>> glix-managed inputs >>>
# <<< glix-managed inputs <<<

# >>> glix-managed hosts >>>
# <<< glix-managed hosts <<<
```

These markers are placed by `flake.nix.tmpl` at `glix init` time.
glix promises:

- Never write outside any pair of markers.
- Always replace the contents between a matched pair atomically.
- Refuse to operate if a marker is missing or out of order.

## `PatchFile(path, regions map[string]string) error`

Reads the file, locates each requested region, replaces the content
between markers with the supplied string, and writes back atomically
(temp file + fsync + rename).

If multiple regions are passed, the file is read once and rewritten
once.

## `RenderInputs(m *manifest.Manifest) string`

For each enabled package, produces:

```nix
<name> = {
  url = "<flake-with-optional-?rev=pin>";
};
```

Pin application: `applyPin(flake, pin)` returns:

- `flake` unchanged if `pin == ""` or `flake` already has `rev=` in its
  query.
- `flake + "?rev=" + pin` if `flake` has no `?` (no query yet).
- `flake + "&rev=" + pin` if `flake` already has a `?` (preserve the
  existing query).

The output is **sorted alphabetically** by package name so the diff is
deterministic.

## `RenderHosts(entries []HostEntry) string`

For each host:

```nix
<name> = mkHost "<name>" "<system>";
```

Sorted alphabetically. `mkHost` is defined statically in `flake.nix.tmpl`.

## Why this, not a real parser

A full Nix parser is huge, takes a slow dependency, and gives glix the
ability to mangle code it shouldn't touch. The anchored-region approach
is:

- **Bounded.** glix can only write inside the markers; user code outside
  is untouchable.
- **Auditable.** A `git diff` after any glix command shows exactly what
  changed.
- **Inspectable.** Users can read the markers and the content between
  them like any other config file.

The cost is that any new piece of state glix needs to push into
`flake.nix` requires a new region, and regions are append-only (you can't
remove one without breaking older user repos). In practice we've kept it
to two: inputs and hosts.

## Tests

`glix/internal/flake/patcher_test.go` covers:

- Round-trip stability (writing the same render twice is a no-op).
- Pin matrix (`applyPin` with empty, with existing query, with existing
  `rev=`, with sourcehut variant).
- Marker-missing → error.
- Multiple regions in one pass.

## Future-proofing

Adding a new region is a small ceremony; see [CLI internals → Adding a
new region](./cli-internals#adding-a-new-region) for the exact steps.
