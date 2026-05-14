---
title: Architectural Decision Records
sidebar_label: Overview
---

# Architectural Decision Records

Every non-trivial design choice in glixos is captured here. Each ADR follows
the same lightweight shape: **context → decision → consequences**. ADRs are
append-only: once accepted, they are amended by adding new ADRs rather than
editing old ones.

## Index

| ID                                            | Title                                                | Status   |
|-----------------------------------------------|------------------------------------------------------|----------|
| [ADR-001](./ADR-001-two-flakes.md)               | Two flakes, one direction of dependency              | Accepted |
| [ADR-002](./ADR-002-flake-contract.md)           | Flake contract: optional `nixosModule` / `homeModule`| Accepted |
| [ADR-003](./ADR-003-manifest-driven.md)          | Manifest-driven, glix writes one file                | Accepted |
| [ADR-004](./ADR-004-resolver-chain.md)           | Resolver chain: URI → glixos registry → nix registry | Accepted |
| [ADR-005](./ADR-005-multi-host.md)               | Multi-host: one repo, `hosts/<name>/` subdir         | Accepted |
| [ADR-006](./ADR-006-secrets-out-of-scope.md)     | Secrets are out of scope for v1                      | Accepted |
| [ADR-007](./ADR-007-flakes-only.md)              | Flakes-only, no `nix-channel`                        | Accepted |
| [ADR-008](./ADR-008-rollback-coupling.md)        | `glix rollback` reverts manifest and generation together | Accepted |
| [ADR-009](./ADR-009-per-package-config.md)       | Per-package config injected as `glixConfig`          | Accepted |
| [ADR-010](./ADR-010-multi-user-routing.md)       | Home-scope routing keyed by `Package.user`           | Accepted |

## Writing a new ADR

1. Copy an existing file and increment the number.
2. Use front-matter `title` and `sidebar_label` consistent with the index
   above.
3. Keep the three-section shape; resist adding a long preamble.
4. Add the row to the index table and to `sidebars.js`.
