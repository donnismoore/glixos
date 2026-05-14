---
title: 'ADR-009: Per-package config'
sidebar_label: 'ADR-009 Per-package config'
---

# ADR-009 — Per-package config injected as `glixConfig`

**Status:** Accepted
**Added in:** M7

## Context

Most non-trivial packages need a small amount of configuration (a theme
name, a port, a feature toggle). Two extremes are bad:

- **Edit Nix by hand.** Breaks ADR-003: glix would lose track of state and
  users would learn divergent override patterns per package.
- **A typed schema in `glix.toml`.** Pushes every package's option surface
  into the manifest format and forces glix to validate or transit types.

## Decision

Each package may carry a `[packages.<name>.config]` sub-table in
`glix.toml`. It is a flat `string -> string` map. `importManifest` exposes
its contents to the package's module as the `glixConfig` module argument
(via `_module.args.glixConfig`):

```toml
[packages.greeting]
flake   = "github:owner/pkg-greeting"
scope   = "home"
enabled = true

[packages.greeting.config]
message = "Hello, world!"
```

```nix
homeModules.default = { pkgs, glixConfig ? { }, ... }: {
  home.file.".greeting".text =
    (glixConfig.message or "Hello from glixos.") + "\n";
};
```

`glix set <pkg> config.<key>=<value>` mutates the table; an empty value
deletes the key.

## Consequences

- The manifest's typed surface stays trivially small (`schema = 1` was not
  bumped; this is an additive optional field).
- Casting is the package module's responsibility, which keeps glix language-
  and toolchain-agnostic and matches how every other Nix flake exposes
  options.
- Default values live in the package module, not in `glix.toml`, so a fresh
  install with no config still works.
- `glix show` and `glix info` surface the config table for visibility.
- `glix set foo config.bar=` is the explicit "delete" form, which is
  consistent with how `glix set foo user=` already cleared the user field.
