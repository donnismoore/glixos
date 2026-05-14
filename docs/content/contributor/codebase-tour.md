---
title: Codebase tour
sidebar_position: 4
---

# Codebase tour

A guided walk through the directories you'll most often touch.

## `core/` (Nix)

The OS layer flake.

| Path                          | What's there                                            |
|-------------------------------|---------------------------------------------------------|
| `flake.nix`                   | `nixosModules.glixos`, `nixosModules.desktop`, `lib.importManifest`, a smoketest `nixosConfigurations.vm`. |
| `lib/importManifest.nix`      | Pure function: `(manifestPath, inputs, defaultUser) → { systemModules, homeModules, homeModulesByUser, manifest }`. |
| `modules/core/default.nix`    | Aggregator. Imports the focused submodules below.       |
| `modules/core/boot.nix`       | Boot loader defaults.                                   |
| `modules/core/locale.nix`     | Locale / time defaults.                                 |
| `modules/core/networking.nix` | NetworkManager etc.                                     |
| `modules/core/nix.nix`        | Flakes-on, channels-off, sane GC defaults.              |
| `modules/core/users.nix`      | `mutableUsers = false`-ish defaults.                    |
| `modules/desktop/default.nix` | Optional desktop profile.                               |
| `hosts/vm/default.nix`        | Smoketest host (`nix build .#vm`).                      |

The core flake has its own `nixpkgs` input. `user-packages` overrides it
via `inputs.nixpkgs.follows` so there's one nixpkgs in the lock file.

## `glix/` (Go)

The CLI. Module path: `github.com/glixos/glix`.

```
glix/
├── cmd/glix/             # main + one file per subcommand
│   ├── main.go           # registry of commands, dispatch
│   ├── util.go           # resolveRepo, currentHostname, requireValidIdent
│   ├── registry_helpers.go
│   ├── cmd_init.go
│   ├── cmd_add.go        # …and one cmd_<verb>.go per command
│   └── ...
└── internal/
    ├── manifest/         # TOML schema, Load/Save/Encode
    ├── flake/            # anchored-region patcher + input/host renderers
    ├── nix/              # exec wrappers: FlakeLock, FlakeUpdate, Rebuild, CollectGarbage
    ├── registry/         # JSON fetch + cache
    ├── repo/             # Repo handle, git, Snapshot for rollback
    ├── resolver/         # URI/registry/nix-registry chain
    ├── templates/        # //go:embed-ed flake.nix.tmpl, host.nix.tmpl, glix.toml.tmpl
    └── ui/               # tiny table renderer
```

### Commands

One file per command, all in `cmd/glix/`. Common pattern:

```go
func cmdFoo(args []string) error {
    fs := newFlagSet("foo")
    host := fs.String("host", currentHostname(), "...")
    dir := fs.String("dir", "", "...")
    apply := fs.Bool("apply", false, "...")
    if err := fs.Parse(args); err != nil { return err }

    r, err := resolveRepo(*dir)
    if err != nil { return err }

    m, err := manifest.Load(r.ManifestPath(*host))
    if err != nil { return err }

    // … mutate pkg in m …

    snap, err := r.TakeSnapshot(*host)
    if err != nil { return err }

    if err := manifest.Save(r.ManifestPath(*host), m); err != nil {
        _ = snap.Restore()
        return err
    }
    if err := regenerateFlake(r); err != nil {
        _ = snap.Restore()
        return err
    }
    if err := nix.FlakeLock(r.Root); err != nil {
        if rerr := snap.Restore(); rerr != nil {
            return fmt.Errorf("...: %v / %v", err, rerr)
        }
        return fmt.Errorf("nix flake lock failed; rolled back: %w", err)
    }
    if err := r.Commit(fmt.Sprintf("glix foo %s: ...", *host)); err != nil {
        return fmt.Errorf("git commit: %w", err)
    }

    // optionally apply
    return nil
}
```

If you're adding a new mutating command, copy `cmd_set.go` and adapt.

### Manifest

`internal/manifest/manifest.go`:

- `Manifest`, `Settings`, `Package`, `Scope`.
- `Load(path) (*Manifest, error)` — permissive TOML decode (BurntSushi).
- `Encode(io.Writer, *Manifest) error` — hand-rolled, deterministic.
- `Save(path, *Manifest) error` — atomic write (tmpfile + rename).
- `PackageNameFromRef(ref) (string, bool)` — derive a package name from a
  flake URI. Used by `glix add` when `--name` is not given.

The encoder emits fields in a fixed order, with `omitempty`-style
suppression for empties.

### Flake patcher

`internal/flake/patcher.go`:

- `RegionInputs`, `RegionHosts` — string constants for the anchor names.
- `PatchFile(path, map[region]string) error` — rewrites the file with
  each named region replaced.
- `RenderInputs(m *manifest.Manifest) string` — builds the
  `<name> = { url = "..."; };` block from the manifest's packages.
- `RenderHosts(entries []HostEntry) string` — builds the
  `<name> = mkHost "..." "..."` block.
- `applyPin(flake, pin string) string` — internal; folds `?rev=` into
  github/gitlab/sourcehut URLs.

### Resolver

`internal/resolver/resolver.go`:

- `Resolve(input string, opts Options) (Resolution, error)`.
- Three sources: `SourceURI`, `SourceRegistry`, `SourceNixRegistry`.
- `NixCLI` shells out to `nix registry list`; `NoNixRegistry` disables
  that fallback.

### Snapshot

`internal/repo/snapshot.go`:

- `(*Repo).TakeSnapshot(host)` returns a `*Snapshot` capturing the
  bytes of `glix.toml`, `flake.nix`, and `flake.lock` in memory.
- `(*Snapshot).Restore()` atomically rewrites them back.

Use this in every mutating command between "validate input" and "perform
side effects".

### Nix wrappers

`internal/nix/exec.go`:

- `FlakeLock(root)`, `FlakeUpdate(root, inputs...)`, `Rebuild(root, host, action)`,
  `CollectGarbage(deleteOld bool)`, `Version()`, `HasFlakes()`.
- All shell out to `nix` / `nixos-rebuild` / `nix-collect-garbage`.

## `examples/`

Reference flakes. Treat them as the source of truth for the contract.

| Path                       | Demonstrates                                            |
|----------------------------|---------------------------------------------------------|
| `pkg-hello/`               | `packages.${system}.default` fallback (no module).      |
| `pkg-greeting/`            | `homeModules.default` with `glixConfig` opt-in.         |
| `user-packages/`           | A hand-written user-packages flake (pre-`glix init`).   |

## `registry/`

`registry.json` — the seed registry. In production this would live in
its own repo; for now it's beside the code so we can iterate together.

## `docs/`

This Docusaurus site. See [Build & test](./build-test) for how to run it
locally.
