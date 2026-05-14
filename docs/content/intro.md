---
slug: /
title: glixos
sidebar_position: 1
---

# glixos

> Modular, flake-driven NixOS with a manifest you can read.

**glixos** is a NixOS distribution organised around two ideas:

1. **One file declares your system.** A single `glix.toml` per host lists every
   package you've installed, what scope it lives in (system or home), which
   user a home-scope package targets, and any per-package config. Everything
   else is static Nix that reads that file.
2. **A small Go CLI is the only writer.** `glix` is a thin tool that mutates
   `glix.toml` and an anchored region of one generated `flake.nix`. It never
   parses Nix; it never edits user code. Every mutation is a single git
   commit, which makes rollback trivial.

The result is a NixOS where the source of truth is small, diffable, and
recoverable, while the full power of the flake ecosystem stays on the table.

## Pick your path

<div className="row" style={{ marginTop: '1.5rem' }}>
  <div className="col col--6">
    <div className="card padding--md margin-bottom--md">
      <h3>I want to use glixos</h3>
      <p>Install it, manage packages, share modules between hosts, recover
      from a bad rebuild. Start here.</p>
      <a className="button button--primary" href="user/getting-started">
        Get started →
      </a>
    </div>
  </div>
  <div className="col col--6">
    <div className="card padding--md margin-bottom--md">
      <h3>I want to understand or hack on glixos</h3>
      <p>Architecture, design decisions, codebase tour, and how to build
      and test the Go CLI plus the Nix layer.</p>
      <a className="button button--primary" href="contributor/overview">
        Contribute →
      </a>
    </div>
  </div>
</div>

## What state is the project in?

glixos is pre-release. The schema, command surface, and on-disk layout are
stabilising but the project is still iterating on milestones. The current
CLI version is `0.1.0-m7`, which covers:

- Bootstrap, add/remove/enable/disable/set/show/list/info/doctor/rollback/gc
- Multi-host (`hosts/<name>/glix.toml`), multi-user (`Package.user`), per-host
  Nix system tuple
- Per-package config (`[packages.<name>.config]`) wired to the package's
  module via `_module.args.glixConfig`
- Pinning via `?rev=<pin>` for github/gitlab/sourcehut refs
- Registry + resolver chain (URI → glixos registry → `nix registry list`)
- Transactional rollback: every mutation snapshots manifest + flake + lock
  and atomically restores them if `nix flake lock` fails

See [Milestones](contributor/milestones) for the roadmap.
