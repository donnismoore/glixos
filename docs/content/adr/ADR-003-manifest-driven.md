---
title: 'ADR-003: Manifest-driven, glix writes one file'
sidebar_label: 'ADR-003 Manifest-driven'
---

# ADR-003 — Manifest-driven, glix writes one file

**Status:** Accepted

## Context

Letting a Go program edit arbitrary Nix is fragile. Letting users hand-edit
generated Nix invites drift.

## Decision

glix writes only `hosts/<host>/glix.toml`. A static `lib/importManifest.nix`
in core reads the TOML and produces module lists. The user-packages
`flake.nix` is generated once by `glix init` and only ever mutated by `glix`
inside delimited regions:

```
# >>> glix-managed inputs >>>
# <<< glix-managed inputs <<<

# >>> glix-managed hosts >>>
# <<< glix-managed hosts <<<
```

These markers are the only surface the patcher touches.

## Consequences

- No Nix parser in Go.
- Users can freely add Nix files under `shared/` or `hosts/<host>/` without
  touching glix-managed regions.
- Manifest schema becomes a versioned interface (`schema = 1`). Additive
  fields preserve compat (`omitempty` in the encoder, permissive decode in
  the loader); a breaking change bumps `schema` and forces an explicit
  migration.
- The manifest is human-readable, diffable, and grep-friendly.
