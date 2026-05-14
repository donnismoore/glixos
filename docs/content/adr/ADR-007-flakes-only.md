---
title: 'ADR-007: Flakes-only, no nix-channel'
sidebar_label: 'ADR-007 Flakes-only'
---

# ADR-007 — Flakes-only, no `nix-channel`

**Status:** Accepted

## Context

Reproducibility is a core promise. Channels introduce implicit, mutable
state outside the flake graph.

## Decision

The core module sets `nix.channel.enable = false` and
`nix.settings.experimental-features = [ "nix-command" "flakes" ]`.
Channels are not supported.

## Consequences

- A single source of truth for inputs (`flake.lock`).
- Tools that assume `<nixpkgs>` won't work without explicit overrides;
  users are directed to use flake-based equivalents.
- `glix update` is the canonical way to advance any input: a registered
  package, `nixpkgs`, `home-manager`, or `glixos-core`.
