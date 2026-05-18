---
title: Milestones
sidebar_position: 12
---

# Milestones

A running log of what's shipped and what's next. Each milestone is a
self-contained chunk that ends in a tested, committed state.

## Shipped

### M1 — Core flake skeleton

A minimal NixOS system boots in a QEMU VM. No glix yet; `glix.toml` was
stubbed by hand.

- `core/flake.nix` with `nixosModules.glixos`, `nixosModules.desktop`.
- `core/modules/core/*.nix` focused submodules.
- `nix build .#vm` works.

### M2 — Manifest contract end-to-end

`importManifest.nix` ingested a hand-written `glix.toml` and produced
module lists. Proved the fallback path and the `homeModules.default`
path both work.

- `core/lib/importManifest.nix` v1.
- `examples/pkg-hello/` and `examples/pkg-greeting/`.
- `examples/user-packages/` — a hand-written reference deployment.

### M3 — glix CLI MVP

The Go CLI with `init`, `add`, `remove`, `list`, `rebuild`. TOML
round-trip preserving deterministic order. Anchored-region patching.

- `glix/internal/manifest/`, `glix/internal/flake/`, `glix/internal/repo/`.
- Embedded init templates.
- First end-to-end `glix add` → `nixos-rebuild`.

### M4 — Registry + resolver + transactional rollback

Short-name resolution via the chain (URI → glixos registry →
`nix registry list`). In-memory snapshots cover every mutation.

- `glix/internal/resolver/`.
- `glix/internal/registry/` with on-disk cache.
- `glix/internal/repo/snapshot.go`.

### M5 — Lifecycle commands

`update`, `enable`, `disable`, `set`, `show`, `doctor`, `rollback`.

- One file per command in `cmd/glix/`.
- `glix doctor` covers toolchain, repo, flake, and manifest checks.
- `glix rollback` reverts the last commit and relocks.

### M6 — Multi-host, multi-user, pinning, gc

- `Settings.system`, `Settings.primary_user`, `Package.user`,
  `Package.pin`.
- `homeModulesByUser` in `importManifest.nix`.
- `glix gc` wrapping `nix-collect-garbage`.
- Per-host system tuples allow mixed-arch repos.

### M7 — Per-package config and repo introspection

- `Package.config` (`[packages.<name>.config]` flat string map).
- `_module.args.glixConfig` wiring via `withGlixConfig`.
- `glix set <pkg> config.<key>=<value>` (empty value deletes).
- `glix info` showing repo state and `flake.lock` inputs.
- `pkg-greeting` consumes `glixConfig.message`.

### M8 — Package the CLI as a derivation

The repo root becomes the canonical flake; `core/` is now a plain
subdirectory of modules + lib (no longer its own flake).

- Root `flake.nix` exposes `packages.${system}.glix` via
  `buildGoModule`, plus the existing `lib`, `nixosModules`, `vm`
  outputs.
- `nixosModules.glixos` is now a function of `inputs`; it applies
  `overlays.default` (which adds `pkgs.glix`) and installs the CLI via
  `environment.systemPackages`. Gated by `glixos.glix.enable` (default
  true) with a `glixos.glix.package` override.
- `glix --version` / `glix -v` aliases; `glix version` prints
  `glix <ver> (<commit>)`, with both stamped at build time via
  ldflags.
- `defaultCoreURL` in `glix init` is now `github:powerreddude/glixos`
  (no `?dir=core`).

## Planned

### M9 — ISO installer

Installer ISO via `nixos-generators`. First-boot wizard runs `glix init`.

### M10 — Secrets stories documented

Worked examples for `sops-nix` and `agenix` integration. (ADR-006 keeps
secrets out of glix itself; this milestone is doc-only.)

### M11 — Public registry

Move the registry into its own repo. Establish a contribution process
for adding packages.

### Beyond

- `glix rollback --to <generation>` with tag-based mapping.
- A real `--dry-run` for every mutation, with a render of the planned
  diff.
- A test harness that boots `nixos-generators` images for end-to-end
  CI.
