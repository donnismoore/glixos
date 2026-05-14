---
title: 'ADR-010: Multi-user routing for home-scope packages'
sidebar_label: 'ADR-010 Multi-user routing'
---

# ADR-010 — Home-scope routing keyed by `Package.user`

**Status:** Accepted
**Added in:** M6

## Context

Originally, every `scope = "home"` package was wired into the home-manager
configuration of a single user (`@USER@`, substituted at `glix init`).
That works for single-user laptops but fails for shared workstations,
servers with multiple service accounts, or households where two users want
different home environments.

## Decision

- `Settings.primary_user` records the default target user for home-scope
  packages.
- `Package.user` overrides that default per package.
- `importManifest` returns `homeModulesByUser`: an attrset keyed by
  resolved user (`Package.user` → `Settings.primary_user` → `defaultUser`
  argument), with a list of modules each.
- The host template uses `lib.mapAttrs` over `homeModulesByUser` to
  populate `home-manager.users.<user>.imports`.
- The legacy flat `homeModules` list is preserved for back-compat.

## Consequences

- Existing single-user setups keep working unchanged: every package without
  `user = "..."` lands on `primary_user`.
- Adding a package for another user is `glix add --scope=home --user=alice
  <ref>`. The host's `users.users.alice` must exist (either via the
  template's `users.users."@USER@"` block, a `shared/` module, or a stanza
  the user adds by hand to `hosts/<host>/default.nix`).
- The host template now derives the home-manager `users` attrset
  generically; users no longer hand-list specific accounts.
- `glix list` and `glix show` surface the resolved user so the routing is
  visible without grepping the manifest.
