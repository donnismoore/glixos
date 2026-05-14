---
title: Quickstart
sidebar_position: 3
---

# Quickstart

Five minutes to a working glixos repo. Assumes you've already installed
`glix` (see [Installation](./installation)) and that you're on a machine
where `nix flake` commands work.

## 1. Initialise the repo

```bash
glix init \
  --host laptop \
  --user alice \
  --system x86_64-linux
```

This creates `$XDG_CONFIG_HOME/glixos` (or `~/.config/glixos`), populates
it with `flake.nix`, `hosts/laptop/{default.nix,glix.toml}`, runs
`git init`, and makes the first commit:

```
initialized glixos user-packages repo at /home/alice/.config/glixos
  host:   laptop
  user:   alice
  system: x86_64-linux
  core:   github:glixos/glixos?dir=core
```

Flags:

| Flag        | Default                          | Meaning                                       |
|-------------|----------------------------------|-----------------------------------------------|
| `--host`    | `$(hostname)`                    | `hosts/<name>/` directory to create           |
| `--user`    | `$USER`                          | Primary user for home-scope packages          |
| `--system`  | `x86_64-linux`                   | Nix system tuple for this host                |
| `--dir`     | `$XDG_CONFIG_HOME/glixos`        | Repo root                                     |
| `--core`    | `github:glixos/glixos?dir=core`  | Flake URL for `glixos-core`                   |

## 2. Add a package

```bash
cd ~/.config/glixos
glix add github:powerreddude/glixos?dir=examples/pkg-hello
```

What just happened:

1. The ref was used verbatim (it's a URI).
2. The package was added to `hosts/laptop/glix.toml` and to the
   glix-managed inputs region of `flake.nix`.
3. `nix flake lock` ran to pin the new input.
4. A git commit was created.

```
added pkg-hello -> github:powerreddude/glixos?dir=examples/pkg-hello (scope=system) to host laptop
staged. run `glix rebuild` to apply.
```

By default, packages are **staged** — they live in the manifest but the
system is not rebuilt yet.

Want to apply immediately? Add `--apply`:

```bash
glix add --apply github:powerreddude/glixos?dir=examples/pkg-hello
```

## 3. Inspect

```bash
glix list
```

```
NAME      SCOPE   STATE    USER  FLAKE
pkg-hello system  enabled        github:powerreddude/glixos?dir=examples/pkg-hello
```

```bash
glix show pkg-hello
```

```
name    pkg-hello
flake   github:powerreddude/glixos?dir=examples/pkg-hello
scope   system
state   enabled
host    laptop
```

```bash
glix info
```

```
root    /home/alice/.config/glixos
git     clean
head    glix add laptop: pkg-hello (...)
hosts:
  laptop       system=x86_64-linux primary_user=alice packages=1/1 enabled
flake.lock:
  glixos-core          github:powerreddude/glixos
  home-manager         github:nix-community/home-manager @ ...
  nixpkgs              github:NixOS/nixpkgs @ ...
  pkg-hello            github:powerreddude/glixos
```

## 4. Rebuild

```bash
sudo glix rebuild switch
```

This is just a wrapper around `nixos-rebuild --flake .#laptop switch`. The
`switch` action is the default and can be omitted.

Other actions are forwarded as-is: `boot`, `test`, `build`, `dry-build`,
`dry-activate`.

## 5. Recover from a bad rebuild

If a rebuild fails or activates something you didn't want:

```bash
sudo nixos-rebuild --rollback switch    # boot back into the previous generation
glix rollback                           # revert the manifest to match
```

If `nix flake lock` itself fails during `glix add` or `glix update`, glix
already rolled back the manifest before returning the error — no manual
intervention needed.

## Where next

- [Commands](./commands) — full surface.
- [Manifest](./manifest) — schema reference.
- [Multi-host](./multi-host) — share a repo across machines.
- [Multi-user](./multi-user) — home-scope packages for more than one user.
- [Per-package config](./config) — wire options through `glixConfig`.
