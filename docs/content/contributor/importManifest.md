---
title: importManifest
sidebar_position: 6
---

# `importManifest.nix`

The pure Nix function that turns `glix.toml` plus a set of flake inputs
into module lists for NixOS and home-manager.

Location: `core/lib/importManifest.nix`.

## Signature

```nix
{ lib }:
{ manifestPath
, inputs ? { }
, defaultUser ? "user"
}:
{
  manifest;          # the raw decoded TOML
  systemModules;     # list of NixOS modules
  homeModules;       # flat list of home-manager modules (legacy)
  homeModulesByUser; # { <user> = [ module ... ]; ... }
}
```

Outer parameter (`lib`) is curried at the flake's `lib =` boundary.
Inner parameters come from the caller (the user-packages `flake.nix`).

## Behaviour

1. Read and decode the manifest. Reject anything other than `schema = 1`.
2. Filter to `enabled = true` packages.
3. For each package, derive the user-facing **module** via the contract:
   - If `inputs.<name>.nixosModules.default` exists and scope is system, use it.
   - Else if scope is system, wrap `inputs.<name>.packages.${pkgs.system}.default`
     into `environment.systemPackages`.
   - Symmetric for `home`.
4. Wrap the resolved module with `withGlixConfig`:
   ```nix
   { imports = [ mod ]; _module.args.glixConfig = pkg.config or { }; }
   ```
   This is how `[packages.<name>.config]` reaches the package module.
5. For home-scope packages, tag the resolved+wrapped module with its
   target user (`Package.user` → `Settings.primary_user` → `defaultUser`).
   Group by user into `homeModulesByUser`.

## Why a wrapper module

Each package module may be a function, an attrset, or even a list. The
`withGlixConfig` wrapper imports the original module and sets one
`_module.args` entry, which works uniformly for all module shapes:

```nix
withGlixConfig = cfg: mod: {
  imports = [ mod ];
  _module.args.glixConfig = cfg;
};
```

A package that doesn't care about `glixConfig` can ignore it; its module
function signature simply doesn't accept the argument, and Nix happily
passes it without complaint (because `_module.args` only requires
declared modules to opt in).

## Why two outputs (`homeModules` and `homeModulesByUser`)

Older host templates used `homeModules` as a flat list under a single
home-manager user. The M6 multi-user routing (see
[ADR-010](../adr/ADR-010-multi-user-routing)) introduced
`homeModulesByUser`. The flat list is preserved so older user
repositories keep evaluating; new ones (post-M6 `glix init`) consume
the attrset variant.

## Error handling

- **Unknown schema:** the function throws with a clear message naming
  the manifest path.
- **Missing flake input:** when the manifest mentions a package but the
  user-packages `flake.nix` has no matching input (e.g., the user
  hand-edited the manifest without running glix), the function throws
  via `requireInput` with a hint to re-run `glix add`.
- **Empty manifest:** returns empty module lists. Useful for
  bootstrapping a host that hasn't had any packages added yet.

## Testing it

`importManifest` is a pure function over disk-resident state, so it's
straightforward to evaluate directly:

```bash
cd ~/.config/glixos
nix eval --raw .#nixosConfigurations.laptop.config.environment.systemPackages \
  --apply 'builtins.toString'

nix eval --json .#nixosConfigurations.laptop.config.home-manager.users \
  --apply 'lib: builtins.attrNames lib'
```

The smoke-test pattern used during development is:

```bash
nix eval --raw .#nixosConfigurations.<host>.config.home-manager.users.<user>.home.file.\".whatever\".text
```

…which exercises the full chain (manifest → glixConfig → module).
