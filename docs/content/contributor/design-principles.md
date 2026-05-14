---
title: Design principles
sidebar_position: 2
---

# Design principles

Five rules. They are the rules; everything else follows. If a proposed
change violates one of these, that's grounds for rejection or a new ADR.

## 1. Two flakes, one direction

The core flake never depends on the package layer. The package layer
imports the core and a set of third-party flakes. This keeps the OS
maintainable as a unit, and keeps user state out of the project repo.

See [ADR-001](../adr/ADR-001-two-flakes).

## 2. Flakes own their config

Each package flake may ship its own `nixosModules.default` and/or
`homeModules.default`. The core never knows what packages exist; it
just imports module lists produced by `importManifest`.

A package's config surface is **its own module's** options surface
(plus the glixConfig hatch for the flat-string case). glix does not
maintain a per-package option schema.

See [ADR-002](../adr/ADR-002-flake-contract),
[ADR-009](../adr/ADR-009-per-package-config).

## 3. One file, one writer

glix writes exactly one file: `hosts/<host>/glix.toml`. All Nix glue is
static and reads that manifest. `flake.nix` has two anchored regions
glix maintains (inputs and hosts); everything else in `flake.nix` is
yours forever.

See [ADR-003](../adr/ADR-003-manifest-driven).

## 4. No Nix parsing in Go

glix only edits delimited regions in template files it generated itself.
Semantic logic lives in Nix.

If you find yourself wanting an AST, you've stepped outside the contract.
Either move the semantics into Nix (where there's already a parser
infrastructure), or extend the manifest schema so the data lives in TOML.

## 5. Reproducible by construction

The user-packages flake is a git repo. `flake.lock` plus `glix.toml`
plus user-owned Nix fully describe the system. Every glix mutation is
one commit; if any step fails, snapshots restore manifest + `flake.nix`
+ `flake.lock` before the error returns.

See [ADR-007](../adr/ADR-007-flakes-only),
[ADR-008](../adr/ADR-008-rollback-coupling).

## How to apply these

When reviewing a change:

1. Does it preserve the dependency direction? (No core → user imports.)
2. Does it require glix to understand Nix semantics it doesn't already?
   (If yes, push the semantics into a Nix function.)
3. Does it add or change a writable file outside `glix.toml` and the
   anchored regions of `flake.nix`?
4. Does it bump the manifest schema? (Then it needs a migration plan
   and an ADR.)
5. Is every state transition still one commit? (Then rollback works.)
