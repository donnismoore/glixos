---
title: 'ADR-005: Multi-host layout'
sidebar_label: 'ADR-005 Multi-host'
---

# ADR-005 — Multi-host: one repo, `hosts/<name>/` subdir

**Status:** Accepted

## Context

Most users have more than one machine and want to share overrides between
them, but per-host divergence (graphics drivers, hostname, services) is the
norm.

## Decision

A single user-packages repo holds one flake at the root and one subdirectory
per host under `hosts/<hostname>/`. Each host has its own `glix.toml` and
`default.nix`. Shared Nix may live anywhere outside `hosts/<host>/`; common
practice is a `shared/` directory imported from each host's `default.nix`.
The default host is `$(hostname)`; `glix --host NAME` selects another.

## Consequences

- Cross-host sharing is a normal Nix import, no special tooling.
- glix always reads/writes a host-qualified manifest path; no global
  manifest exists.
- A user with one machine sees `hosts/<hostname>/` with no extra burden.
- Per-host overrides extend to the Nix system tuple via
  `[settings] system = "..."`, so a single repo can declare both
  `x86_64-linux` and `aarch64-linux` hosts.
