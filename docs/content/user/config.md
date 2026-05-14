---
title: Per-package config
sidebar_position: 7
---

# Per-package config

Many packages need a small amount of configuration: a theme, a port, a
feature flag. glixos surfaces this via `[packages.<name>.config]` in
`glix.toml`. Package modules read it as the `glixConfig` module argument.

See [ADR-009](../adr/ADR-009-per-package-config) for why the design is
shaped this way.

## The shape

```toml
[packages.greeting]
flake   = "github:owner/pkg-greeting"
scope   = "home"
enabled = true

[packages.greeting.config]
message = "Hello, world!"
theme   = "dark"
```

- **Flat.** No nested tables. A path-like name like `theme.color` is a
  single key, not nested config.
- **String values only.** Everything is a TOML string. The package module
  casts as needed (Nix has `lib.toInt`, `lib.toBool`, etc.).

## Setting and clearing

```bash
glix set greeting config.message="Hello, world!"
glix set greeting config.theme="catppuccin-mocha"
glix set greeting config.message=                   # delete
```

`glix set` accepts multiple `key=value` pairs in one invocation; they all
land in the same commit.

## Reading from a module

The wrapper module that `importManifest.nix` builds injects
`_module.args.glixConfig` with the contents of the config table:

```nix
homeModules.default = { pkgs, glixConfig ? { }, ... }: {
  home.file.".greeting".text =
    (glixConfig.message or "Hello from glixos.") + "\n";

  programs.starship.enable =
    (glixConfig.starship or "false") == "true";
};
```

The `glixConfig ? { }` default keeps the package buildable outside glixos
(e.g. for `nix flake check`).

## Inspect

```bash
glix show greeting
```

```
name    greeting
flake   github:owner/pkg-greeting
scope   home
state   enabled
user    alice
host    laptop
config
  message = Hello, world!
  theme   = dark
```

## Patterns

- **Booleans.** Convention: `"true"`/`"false"`. Compare in Nix with
  `glixConfig.foo or "false" == "true"`.
- **Integers.** Pass as strings, parse with `lib.toInt`.
- **Lists.** Pass as a delimiter-separated string and `lib.splitString`.
  If you find yourself doing this often, consider promoting the option to
  a real NixOS option in the package module and exposing only a single
  boolean toggle through `glixConfig`.

## When **not** to use `glixConfig`

If a setting changes the package's build (not just its runtime config),
prefer pinning a different revision (`pin`) or pointing at a fork of the
flake (`flake=...`). `glixConfig` is for module-time configuration, not
package selection.

If a setting is host-wide rather than package-specific (e.g. system
timezone), put it in `hosts/<host>/default.nix` or in a `shared/` module.
