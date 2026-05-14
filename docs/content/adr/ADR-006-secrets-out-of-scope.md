---
title: 'ADR-006: Secrets are out of scope for v1'
sidebar_label: 'ADR-006 Secrets'
---

# ADR-006 — Secrets are out of scope for v1

**Status:** Accepted

## Context

Secret management is a deep problem with mature solutions (`sops-nix`,
`agenix`). Building a glix-native one would be a project on its own.

## Decision

glix never reads, writes, or transports secret material. Documentation
describes how to add `sops-nix` or `agenix` as ordinary flake inputs and
reference their modules from `shared/` or `hosts/*/`.

## Consequences

- Smallest attack surface for `glix`.
- A first-class `glix secret` subcommand is deferred; if added later it
  must wrap an established tool rather than reinvent.
- Putting secret references in `glix.toml` is fine (they're just strings
  that name a sops file or agenix path); the secret material itself is
  managed by the chosen tool outside the manifest.
