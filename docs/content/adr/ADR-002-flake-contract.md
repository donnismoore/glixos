---
title: 'ADR-002: Flake contract'
sidebar_label: 'ADR-002 Flake contract'
---

# ADR-002 — Flake contract: optional `nixosModule` / `homeModule`

**Status:** Accepted

## Context

Packages need to ship their own configuration (services, dotfiles, kernel
modules) without the core knowing about them, but arbitrary upstream flakes
won't have a glixos-specific schema.

## Decision

A "first-class" glixos package flake may expose `nixosModules.default`
and/or `homeModules.default`. If absent, `glix` falls back to wrapping
`packages.${system}.default` into `environment.systemPackages` or
`home.packages` based on declared scope.

## Consequences

- Any vanilla nixpkgs-style flake is usable on day one.
- Packages that want richer integration get full module power.
- No required `glixos.meta`; metadata is opportunistic.
- ADR-009 later layered per-package config on top of this contract without
  changing the shape of the contract itself: the wrapper module passes
  `_module.args.glixConfig`, so package modules opt in only if they care.
