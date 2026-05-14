---
title: Flake contract
sidebar_position: 10
---

# Flake contract

What a flake must (and may) export to be a first-class glixos package.

See [ADR-002](../adr/ADR-002-flake-contract),
[ADR-009](../adr/ADR-009-per-package-config).

## Required

Nothing strictly. Any flake with `packages.${system}.default` is
installable.

## Strongly recommended

| Output                       | Purpose                                                    |
|------------------------------|------------------------------------------------------------|
| `packages.${system}.default` | Fallback path. Wrapped into `environment.systemPackages` or `home.packages` based on scope. |
| `nixosModules.default`       | Used when scope is `system`. Lets the package own services, kernel modules, etc. |
| `homeModules.default`        | Used when scope is `home`. Lets the package own dotfiles, user services, etc. |

A package can ship any subset. Resolution chooses the right one based on
the scope declared in `glix.toml`.

## `glixConfig` argument

If a package module accepts `glixConfig` in its argument list, it receives
the contents of `[packages.<name>.config]` from the manifest as an
attrset of strings:

```nix
nixosModules.default = { pkgs, glixConfig ? { }, ... }: {
  services.foo = {
    enable = true;
    listenAddress = glixConfig.listen or "127.0.0.1:8080";
  };
};
```

The `? { }` default is what makes the package buildable outside glixos
(`nix flake check`, plain `nixos-rebuild` with hand-written modules).

## What a package should **not** do

- **Modify global system state implicitly.** Package modules should
  only declare options; they should not perform IO at evaluation time.
- **Depend on glix being on PATH at build time.** Packages are built by
  plain `nix`. glixos-specific behaviour comes from glixos's own modules.
- **Override `system.stateVersion`.** That's the host's call.
- **Hard-code a `pkgs.system`.** Always parametrise over `${system}`.

## Testing a package locally

```bash
nix flake check path:./pkg-foo
nix build path:./pkg-foo
nix eval path:./pkg-foo#homeModules.default --apply 'm: builtins.typeOf m'
```

For a glixos integration test:

```bash
glix init --dir=/tmp/glix-test --host=test --user=test
cd /tmp/glix-test
glix add path:/abs/path/to/pkg-foo
nix eval .#nixosConfigurations.test.config.environment.systemPackages \
  --apply 'p: builtins.length p'
```

The `examples/` directory exercises both the fallback and the module
paths; clone them as a starting point.

## Naming

`PackageNameFromRef` in `glix/internal/manifest/manifest.go` derives a
short name from a ref. The rules:

| Ref form                                  | Inferred name |
|-------------------------------------------|---------------|
| `github:owner/repo`                       | `repo`        |
| `github:owner/repo/branch`                | `repo`        |
| `github:owner/repo?dir=subpkg`            | `subpkg`      |
| `path:./examples/pkg-hello`               | `pkg-hello`   |
| `https://example.com/foo.git`             | `foo`         |

If you want a different name, pass `--name`. Names must match
`^[A-Za-z][A-Za-z0-9_-]*$` (TOML bare key and Nix attribute name
compatible).
