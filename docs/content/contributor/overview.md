---
title: Contributor overview
sidebar_position: 1
---

# Contributor overview

glixos is a small project with a few moving parts and strong opinions
about how they fit together. This page is the orientation; read it once,
then dive into the rest of the contributor docs.

## Components

```
glixos/
├── core/                     # glixos-core: the OS layer flake
│   ├── flake.nix             # nixosModules.{glixos,desktop}; lib.importManifest
│   ├── lib/
│   │   └── importManifest.nix
│   ├── modules/core/         # base system (boot, nix, locale, networking, users)
│   ├── modules/desktop/      # optional desktop profile
│   └── hosts/vm/             # internal smoketest host
├── glix/                     # Go CLI
│   ├── go.mod
│   ├── cmd/glix/             # one file per command
│   └── internal/
│       ├── manifest/         # glix.toml read/write
│       ├── flake/            # anchored-region patcher
│       ├── nix/              # nix / nixos-rebuild wrappers
│       ├── registry/         # registry JSON fetch/cache
│       ├── repo/             # user-packages repo handle + git + snapshot
│       ├── resolver/         # ref → canonical flake URI
│       ├── templates/        # //go:embed-ed init templates
│       └── ui/               # table renderer
├── registry/
│   └── registry.json         # default registry contents
├── examples/
│   ├── pkg-hello/            # reference: packages.default fallback
│   ├── pkg-greeting/         # reference: homeModules.default with glixConfig
│   └── user-packages/        # hand-written reference deployment
└── docs/                     # this Docusaurus site
```

Two languages, one direction of dependency: `glix` writes `glix.toml`,
`core` (Nix) reads it. The Go side never parses Nix; the Nix side never
shells out to glix.

## Where to start

| Goal                                                           | Read                                              |
|----------------------------------------------------------------|---------------------------------------------------|
| Understand the architecture in one sitting                     | [Architecture](./architecture)                    |
| Understand why we made the design choices we did               | [Design principles](./design-principles), [ADRs](../adr/) |
| Find your way around the Go code                               | [Codebase tour](./codebase-tour)                  |
| Add or change a CLI command                                    | [CLI internals](./cli-internals)                  |
| Understand the Nix side                                        | [importManifest](./importManifest), [Flake contract](./flake-contract) |
| Understand how the manifest stays consistent on errors         | [Snapshots](./snapshots)                          |
| Build, run tests, ship a release                               | [Build & test](./build-test), [Release](./release)|
| See what's planned                                             | [Milestones](./milestones)                        |

## Working agreement

- **One mutation, one commit.** Every glix command that changes state
  produces exactly one git commit. This is a hard invariant; rollback
  relies on it.
- **Manifest is the truth.** If a feature can be expressed in
  `glix.toml`, it should be. If it can't, it goes in user-owned Nix
  (`hosts/<host>/default.nix`, `shared/`).
- **No Nix parsing in Go.** Anchored regions only. If you find yourself
  reaching for an AST, step back.
- **Schema additive.** Add fields with `omitempty`; only bump `schema =`
  for breaking changes.
- **Tests live next to the code.** `_test.go` in the same package. No
  separate test tree.
- **Errors are explicit and rolled back.** A command must either complete
  fully or restore state. The `repo.Snapshot` type exists to make this
  easy; use it.

## Communication

- Issues: [github.com/powerreddude/glixos/issues](https://github.com/powerreddude/glixos/issues)
- PRs: reference the relevant ADR(s) in the description. If you're
  introducing a new design choice, file the ADR with the PR.

## License

glixos is GPLv3. By contributing you agree to license your contributions
under the same terms. See `LICENSE` at the repo root.
