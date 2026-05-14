---
title: Manifest reference
sidebar_position: 5
---

# Manifest reference

`hosts/<host>/glix.toml` is the only file glix writes. Its schema is
**versioned**; the current version is `schema = 1`. Additive fields ship
without bumping the schema.

## Top level

```toml
schema = 1

[settings]
# ...

[packages.<name>]
# ...
```

## `[settings]`

| Key             | Type    | Default          | Notes                                                                 |
|-----------------|---------|------------------|-----------------------------------------------------------------------|
| `default_scope` | string  | `"system"`       | Scope used by `glix add` when `--scope` is not given                  |
| `auto_apply`    | bool    | `false`          | If true, every mutating command implicitly does `--apply`             |
| `registry_url`  | string  | (empty)          | URL of `registry.json`; if empty, glix uses its compiled-in default   |
| `system`        | string  | `"x86_64-linux"` | Nix system tuple for this host. Written by `glix init --system`       |
| `primary_user`  | string  | (empty)          | Default target user for home-scope packages without an explicit `user` |

## `[packages.<name>]`

| Key       | Type            | Required | Notes                                                                            |
|-----------|-----------------|----------|----------------------------------------------------------------------------------|
| `flake`   | string          | yes      | Flake reference. Anything `nix` accepts.                                         |
| `scope`   | `system`\|`home`| yes      | Where to install the package.                                                    |
| `enabled` | bool            | yes      | If false, the package stays in the manifest but is excluded from module output.  |
| `pin`     | string          | no       | Folded into the flake URI as `?rev=<pin>`. Github/Gitlab/Sourcehut style refs.   |
| `user`    | string          | no       | Only meaningful for `scope = "home"`. Empty means use `Settings.primary_user`.   |

### `[packages.<name>.config]`

A flat `string -> string` table. Surfaced to the package's module as the
`glixConfig` module argument. See [Per-package config](./config).

```toml
[packages.greeting.config]
message = "Hello, world!"
theme   = "dark"
```

## Example

```toml
# glixos manifest. Managed by glix; hand edits preserved on best-effort.
schema = 1

[settings]
default_scope = "system"
auto_apply    = false
system        = "x86_64-linux"
primary_user  = "alice"
registry_url  = "https://raw.githubusercontent.com/glixos/registry/main/registry.json"

[packages.pkg-hello]
flake   = "github:powerreddude/glixos?dir=examples/pkg-hello"
scope   = "system"
enabled = true

[packages.firefox]
flake   = "github:NixOS/nixpkgs/nixos-unstable#firefox"
scope   = "system"
enabled = true
pin     = "da5ad661ba4e5ef59ba743f0d112cbc30e474f32"

[packages.greeting]
flake   = "path:/home/alice/code/pkg-greeting"
scope   = "home"
enabled = true
user    = "alice"

[packages.greeting.config]
message = "Hello, world!"

[packages.work-tools]
flake   = "github:work/tooling-pack"
scope   = "home"
enabled = true
user    = "alice-work"
```

## Hand-edits

glix is the writer-of-record, but it does its best to preserve hand edits.
The encoder emits a deterministic ordering and a constant set of fields,
so re-running `glix list` after a manual `glix.toml` change produces a
clean diff. If you commit a hand edit and then run any mutating glix
command, the encoder will rewrite the file in canonical form on the next
write.

If you need a field glix doesn't know about, prefer adding it to the
package's own module (read from `glixConfig`) rather than the manifest.
