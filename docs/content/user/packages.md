---
title: Packages
sidebar_position: 6
---

# Packages

A "package" in glixos is just **a flake**. glixos imposes a tiny optional
contract on top of standard flake outputs; any flake that builds with
`nix build` is usable.

## The contract

| Output                       | Required | Used when                              |
|------------------------------|----------|----------------------------------------|
| `packages.${system}.default` | fallback | Plain package. glix wraps it.          |
| `nixosModules.default`       | optional | Scope `system`. Imported as a module.  |
| `homeModules.default`        | optional | Scope `home`. Imported as a module.    |

Resolution:

- **scope = system** → use `nixosModules.default` if present, otherwise
  add `packages.${system}.default` to `environment.systemPackages`.
- **scope = home** → use `homeModules.default` if present, otherwise add
  `packages.${system}.default` to `home.packages`.

That's the entire contract. No required metadata, no glixos-specific
attrset.

## A bare package (fallback path)

```nix
# pkg-hello/flake.nix
{
  description = "glixos example: bare hello package";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.hello;
      });
    };
}
```

Add it:

```bash
glix add github:owner/pkg-hello
```

glix puts `hello` into `environment.systemPackages`.

## A home-manager module

```nix
# pkg-greeting/flake.nix
{
  description = "glixos example: home module installing cowsay and a greeting file";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    homeModules.default = { pkgs, glixConfig ? { }, ... }: {
      home.packages = [ pkgs.cowsay ];
      home.file.".glixos-greeting".text =
        (glixConfig.message or "Hello from a glixos package home module.") + "\n";
    };
  };
}
```

Add it:

```bash
glix add --scope=home --user=alice github:owner/pkg-greeting
glix set greeting config.message="Hi, alice."
```

## A full NixOS module

```nix
{
  outputs = { self, nixpkgs }: {
    nixosModules.default = { pkgs, glixConfig ? { }, ... }: {
      services.myservice = {
        enable = true;
        listenAddress = glixConfig.listen or "127.0.0.1:8080";
      };
      environment.systemPackages = [ pkgs.myservice ];
    };
  };
}
```

The pattern is unchanged from standard NixOS modules; the only glixos
addition is `glixConfig`.

## How references resolve

`glix add <ref>` runs the resolver chain:

1. **URI passthrough.** Anything with a flake-style scheme (`github:`,
   `gitlab:`, `git+`, `path:`, `http(s)://`) is used verbatim. The
   package name is derived from the URI:
   - `github:owner/repo` → `repo`
   - `github:owner/repo?dir=subpkg` → `subpkg`
   - `path:./examples/pkg-hello` → `pkg-hello`
2. **glixos registry.** A short name like `firefox` is looked up in the
   registry's `packages` table.
3. **`nix registry list`.** Final fallback.

If two of these would produce the same name, glix refuses to add it; pass
`--name` to disambiguate.

## Pinning

Pass `--pin=<rev>` (or `glix set <name> pin=<rev>`) to lock a package to a
specific revision. The pin is folded into the flake URI as `?rev=<pin>`,
which nix understands for `github:`, `gitlab:`, and `sourcehut:` refs. See
[Pinning](./pinning).

## Choosing a scope

| Scope    | Wires into                                     | Use for                                   |
|----------|------------------------------------------------|-------------------------------------------|
| `system` | NixOS module list (or `environment.systemPackages`) | Services, kernel/firmware, daemons, system-wide programs |
| `home`   | home-manager imports for `Package.user`        | Dotfiles, per-user programs, shell config |

When in doubt: if the upstream flake ships a `homeModules.default`, use
`home`; otherwise `system`.
