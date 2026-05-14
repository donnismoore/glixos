---
title: 'ADR-004: Resolver chain'
sidebar_label: 'ADR-004 Resolver chain'
---

# ADR-004 — Resolver chain: URI → glixos registry → nix registry

**Status:** Accepted

## Context

Users want short names (`glix add firefox`) but full flake URIs must remain
authoritative. A curated registry adds value without locking out arbitrary
flakes.

## Decision

Resolution order for `glix add <ref>`:

1. Anything that looks like a flake URI (`github:`, `gitlab:`, `git+`,
   `path:`, `http(s)://`, contains `#`) is used verbatim.
2. The glixos registry (cached JSON, configurable URL).
3. The user's `nix registry list` entries (unless `--no-nix-registry`).
4. Error with suggestions.

## Consequences

- Curated names get a fast path; everything else still works.
- The registry is just a JSON blob hosted anywhere; there is no server to
  operate.
- `--registry-url` and `--refresh` give users an escape hatch for testing
  alternate registries.
- The chain is symmetric for `glix search`, so the same names resolve in
  both commands.
