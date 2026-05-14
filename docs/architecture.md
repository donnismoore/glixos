# Glixos Architecture

> Status: design draft (pre-M1). Source of truth until M1 lands.

Glixos is a NixOS-based operating system whose distinguishing feature is a strict
separation between the immutable OS core and a user-owned package layer driven
by a Go CLI called `glix`. Packages are flakes; flakes own their own
configuration; `glix` only manipulates a single declarative manifest.

## 1. Design principles

1. **Two flakes, one direction.** The core flake never depends on the package
   layer. The package layer imports the core and a set of third-party flakes.
2. **Flakes own their config.** Each package flake may ship its own
   `nixosModule` / `homeModule`. The core never knows what packages exist.
3. **One file, one writer.** `glix` writes exactly one file: `glix.toml`. All
   Nix glue is static and reads that manifest.
4. **No Nix parsing in Go.** `glix` only edits delimited regions in template
   files it generated itself. Semantic logic lives in Nix.
5. **Reproducible by construction.** The user-packages flake is a git repo;
   `flake.lock` plus `glix.toml` plus user overrides fully describe the system.

## 2. High-level architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        glixos-core                          │
│   (flake repo — owned by the project, pinned by user)       │
│   ├─ flake.nix           # exposes nixosConfigurations,     │
│   │                      #   nixosModules.glixos            │
│   ├─ modules/core/       # base system: boot, networking,   │
│   │                      #   users, locales, glix CLI pkg   │
│   ├─ modules/desktop/    # optional opt-in desktop profile  │
│   └─ lib/                # helpers: importManifest, etc.    │
└──────────────────────┬──────────────────────────────────────┘
                       │ imported as input
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  user-packages flake                        │
│   (git repo on the user's machine — owned by glix)          │
│   ├─ flake.nix           # one nixosConfigurations per host │
│   ├─ hosts/              # per-host configuration           │
│   │   └─ <hostname>/                                        │
│   │       ├─ glix.toml   # the ONLY file glix writes        │
│   │       ├─ system.nix  # static: importManifest "system"  │
│   │       └─ home.nix    # static: importManifest "home"    │
│   ├─ shared/             # modules imported by every host   │
│   └─ overrides/          # free-form user .nix files        │
└──────────────────────┬──────────────────────────────────────┘
                       │ nixos-rebuild switch --flake .
                       ▼
                  Running system
```

## 3. Flake contract for glixos packages

Any flake is consumable, but a flake is **first-class** if it exposes:

| Output                        | Required | Purpose                                                                                  |
|-------------------------------|----------|------------------------------------------------------------------------------------------|
| `packages.${system}.default`  | fallback | Plain package; glix wraps it into `environment.systemPackages` or `home.packages`        |
| `nixosModules.default`        | optional | Self-contained NixOS module (services, kernel modules, system config)                    |
| `homeModules.default`         | optional | Self-contained home-manager module (dotfiles, user services)                             |
| `glixos.meta` (passthru/attr) | optional | Metadata: description, scope hints, conflicts, version, schema                           |

Resolution rules inside the generated module:

- If `scope = system` and `nixosModules.default` exists → import it.
- Else if `scope = system` → add `packages.${system}.default` to
  `environment.systemPackages`.
- If `scope = home` and `homeModules.default` exists → import it into
  home-manager.
- Else if `scope = home` → add to `home.packages`.

## 4. `glix.toml` — the manifest

The single file `glix` writes. Everything else is static Nix that reads it.

```toml
schema = 1

[settings]
default_scope = "system"      # or "home"
auto_apply    = false          # stage by default
registry_url  = "https://raw.githubusercontent.com/glixos/registry/main/registry.json"

[packages.firefox]
flake   = "github:nixos/nixpkgs/nixos-unstable#firefox"
scope   = "system"
enabled = true
pin     = "sha256-..."         # optional informational mirror of flake.lock

[packages.helix]
flake   = "github:helix-editor/helix"
scope   = "home"
config  = { theme = "catppuccin_mocha" }   # passed to homeModules.default
```

## 5. Static Nix glue: `lib/importManifest.nix`

A pure function that, given the parsed TOML and the flake inputs, returns a
list of NixOS modules and a list of home-manager modules.

```nix
{ lib, inputs }:
let
  manifest    = builtins.fromTOML (builtins.readFile ../glix.toml);
  enabledPkgs = lib.filterAttrs (_: p: p.enabled or true) manifest.packages;

  mkPkgModule = input: { config, pkgs, ... }: {
    environment.systemPackages = [ input.packages.${pkgs.system}.default ];
  };

  pickSystem = name: pkg:
    if pkg.scope or "system" != "system" then null
    else inputs.${name}.nixosModules.default or (mkPkgModule inputs.${name});
in {
  systemModules = lib.filter (m: m != null)
    (lib.mapAttrsToList pickSystem enabledPkgs);

  # homeModules: symmetric, scope == "home"
}
```

The user-packages `flake.nix` calls this and feeds the result into
`nixosConfigurations.<host>` and the home-manager config. Users never edit
generated Nix — only `glix.toml` (via `glix`) or files under `overrides/`.

## 6. The `glix` Go CLI

All mutating commands accept `--host NAME` to target a specific host directory
under `hosts/`. Default is `$(hostname)`.

### Command surface (initial)

| Command                                                            | Behavior                                                                                              |
|--------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| `glix init [--host NAME]`                                          | Bootstrap user-packages flake repo in `$XDG_CONFIG_HOME/glixos`; create `hosts/<NAME>/` and `git init`. |
| `glix add <ref> [--scope=system\|home] [--name=N] [--apply]`       | Resolve ref, add to `flake.nix` inputs, append to `glix.toml`, lock. Stage only unless `--apply`.     |
| `glix remove <name>`                                               | Reverse of add.                                                                                       |
| `glix list [--enabled\|--all]`                                     | Read `glix.toml`, render table.                                                                       |
| `glix enable/disable <name>`                                       | Toggle `enabled` flag.                                                                                |
| `glix set <name>.<key> <value>`                                    | Edit `[packages.X.config]` for that package.                                                          |
| `glix rebuild [switch\|test\|boot]`                                | Wraps `nixos-rebuild --flake .#$host <action>`.                                                       |
| `glix update [<name>...]`                                          | `nix flake lock --update-input` for selected (or all) inputs.                                         |
| `glix search <query>`                                              | Query the glixos registry.                                                                            |
| `glix doctor`                                                      | Sanity checks: flake parses, manifest valid, registry reachable, generation consistency.              |
| `glix rollback`                                                    | `nixos-rebuild --rollback` + revert last manifest commit (see §10).                                   |
| `glix gc`                                                          | Wraps `nix-collect-garbage -d`.                                                                       |

### Resolver chain for `glix add <ref>`

1. If `<ref>` looks like a flake URI (`github:`, `gitlab:`, `git+`, `path:`,
   `http(s)://`, contains `#`) → use as-is.
2. Look up in glixos registry (cached locally with TTL).
3. Fall back to `nix registry list`.
4. Error with suggestions.

### Internal layout (Go)

```
cmd/glix/                main.go (cobra root)
internal/
  config/                XDG paths, settings loader
  manifest/              glix.toml read/write (preserves comments + key order)
  flake/                 flake.nix patcher (anchored regions, no full Nix parser)
  registry/              fetch + cache registry JSON; resolution chain
  resolver/              ref → canonical flake URI
  nix/                   thin wrapper over `nix` and `nixos-rebuild`
  state/                 transient cache, last-good generation tracking
  ui/                    table rendering, prompts
pkg/glixos/              public types for library reuse
```

### Implementation rules

- **Atomic writes:** write `glix.toml.tmp`, fsync, rename.
- **Anchored regions:** glix mutates only between
  `# >>> glix-managed >>>` … `# <<< glix-managed <<<` markers in `flake.nix`.
- **Auto-commit (configurable):** every mutating command makes a git commit in
  the user-packages repo. Enables `glix rollback`.
- **Dry-run:** every mutating command supports `--dry-run`.
- **Lockfile authority:** `flake.lock` is the truth for pins; `glix.toml.pin`
  is an optional informational mirror.

## 7. Registry

A separate `glixos/registry` git repo containing `registry.json`:

```json
{
  "schema": 1,
  "packages": {
    "firefox": { "flake": "github:glixos/pkg-firefox", "description": "..." },
    "helix":   { "flake": "github:helix-editor/helix",  "description": "..." }
  }
}
```

Cached locally under `$XDG_CACHE_HOME/glix/registry.json` with TTL.

## 8. Project repo layout

```
glixos/
├── core/                     # the glixos-core flake
│   ├── flake.nix
│   ├── modules/core/
│   ├── modules/desktop/
│   └── lib/importManifest.nix
├── glix/                     # Go CLI
│   ├── go.mod
│   ├── cmd/glix/
│   └── internal/...
├── registry/                 # seed registry (later moved to its own repo)
│   └── registry.json
├── templates/                # files glix init drops into user repo
│   ├── flake.nix.tmpl
│   ├── system.nix
│   └── home.nix
├── examples/
│   └── pkg-hello/            # reference glixos-compatible flake
└── docs/
    ├── architecture.md
    ├── flake-contract.md
    └── glix-cli.md
```

The `glix` binary is packaged as a Nix derivation inside `core/` and included
in `modules/core/`, so a glixos system always has `glix` available.

## 9. Milestones

1. **M1 — Core flake skeleton.** Minimal NixOS system boots in a VM. No glix
   yet; manifest stubbed.
2. **M2 — Manifest import.** `lib/importManifest.nix` + `templates/flake.nix.tmpl`.
   Hand-edit `glix.toml`; prove a package installs system-wide and a
   `homeModule` wires up.
3. **M3 — Glix MVP.** `init`, `add`, `remove`, `list`, `rebuild`. TOML
   round-trip preserving order/comments. Anchored region patching.
4. **M4 — Registry + resolver.** Short-name resolution, search, caching.
5. **M5 — Home-manager integration.** `--scope=home` end-to-end.
6. **M6 — Polish.** `doctor`, `update`, `gc`, `--dry-run`, atomicity,
   auto-commit, `rollback`.
7. **M7 — ISO.** Installer ISO via `nixos-generators` with glix preinstalled.

## 10. Resolved questions

See [`decisions.md`](./decisions.md) for full rationale and implications.

- **Multi-host:** one repo, `hosts/<name>/glix.toml`, shared modules in
  `shared/`. Default host = `$(hostname)`; selectable with `glix --host NAME`.
- **Secrets:** out of scope. Document `sops-nix` / `agenix` as supported
  patterns; `glix` never touches secret material.
- **Channels:** flakes-only. Core module sets `nix.channel.enable = false`.
- **Rollback:** `glix rollback` runs `nixos-rebuild --rollback` and reverts
  the last manifest commit so manifest and running generation stay in sync.
