---
title: 'ADR-008: Rollback couples manifest and generation'
sidebar_label: 'ADR-008 Rollback coupling'
---

# ADR-008 — `glix rollback` reverts manifest and generation together

**Status:** Accepted

## Context

`nixos-rebuild --rollback` switches to the previous generation but leaves
the manifest pointing at the bad state, so the next `glix rebuild` would
re-apply it. Reverting the manifest alone leaves the running system on the
bad generation.

## Decision

`glix rollback` reverts the **manifest** to its previous state by reverting
the last git commit in the user-packages repo and relocking. If you also
want the running generation rolled back, run `nixos-rebuild --rollback`
(or pass `--apply` to `glix rollback`, which rebuilds against the reverted
manifest).

## Consequences

- Manifest and live system never silently diverge as long as every mutation
  is one commit.
- Every mutating glix command must produce exactly one git commit for this
  invariant to hold; this is already required by ADR-003.
- Transactional rollback at the **command** level (see
  `internal/repo/snapshot.go`) is a separate layer that protects against
  partial failures within a single mutation (manifest written, but
  `nix flake lock` fails). It restores manifest, `flake.nix`, and
  `flake.lock` from in-memory snapshots before returning the error.
- Future `glix rollback --to <generation>` can layer on top using
  generation→commit tags.
