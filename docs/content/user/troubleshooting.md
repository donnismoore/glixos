---
title: Troubleshooting
sidebar_position: 14
---

# Troubleshooting

Quick reference for common errors. Start with `glix doctor` if you're not
sure where the problem is.

## `glix doctor` reports a FAIL

`glix doctor` runs a series of read-only checks. The fix usually follows
directly from the failed check:

| Check         | Fix                                                                 |
|---------------|---------------------------------------------------------------------|
| `nix`         | Install nix (or upgrade; need 2.18+).                                |
| `flakes`      | Add `experimental-features = nix-command flakes` to `nix.conf`.      |
| `discovery`   | `cd` into a glixos repo, or pass `--dir`.                            |
| `flake.nix`   | The repo is corrupt; re-run `glix init` in an empty directory.       |
| `flake.lock`  | Run any mutating glix command to relock, or `nix flake lock`.        |
| `<host>`      | The manifest is missing or invalid; check the printed error.         |

## "host X not found"

Every mutating command needs a host directory. Either:

```bash
glix init --host=X --user=...      # create it
glix --host=Y add ...               # target an existing host instead
```

`hosts/<host>/` must contain both `glix.toml` and `default.nix`. `glix
init` creates them; if either is missing, the host is incomplete.

## "package X already present"

```bash
glix add foo
# glix add: package "foo" already present; remove it first or pass --name
```

Either remove the existing entry (`glix remove foo`) or add the new one
with a different name (`glix add foo --name=foo-v2`).

## "nix flake lock failed; rolled back"

glix already restored your manifest. Inspect the underlying nix error
printed alongside the message — common causes:

- Pin that doesn't exist upstream (`pin=garbage`).
- The flake ref doesn't have a `flake.nix` (typo, wrong `?dir=`).
- The flake's own inputs are broken upstream.

Fix the input or pin and try again. Nothing was saved.

## "package X has invalid scope"

The manifest contains `scope = "..."` where the value isn't `system` or
`home`. Hand-edit `glix.toml` (or restore from git history) and try
again.

## Home-manager: "Cannot find user X"

A package is `scope = "home"` with `user = "X"` but `X` is not declared in
NixOS. Either add the user to `hosts/<host>/default.nix`
(`users.users.X = { ... };`) or change the package to a user that does
exist (`glix set <name> user=...`).

## glix wrote my file unexpectedly

glix only ever writes:

- `hosts/<host>/glix.toml`
- Anchored regions in `flake.nix`:
  ```
  # >>> glix-managed inputs >>>
  # <<< glix-managed inputs <<<

  # >>> glix-managed hosts >>>
  # <<< glix-managed hosts <<<
  ```
- `flake.lock` (via `nix flake lock`).
- One git commit per mutation.

If something else changed, it wasn't glix. Inspect with `git diff HEAD~1`.

## I want to undo my last change

```bash
glix rollback                  # revert the last commit, relock
glix rollback --apply          # also nixos-rebuild switch
```

If the system itself is broken and you can't run glix:

```bash
sudo nixos-rebuild --rollback switch   # boot back into the previous gen
# then on the working system:
glix rollback
```

## glix produces a non-deterministic `glix.toml`

It shouldn't. The encoder sorts packages alphabetically and emits fields
in a fixed order. If you see a diff that's "just reordering", file an
issue — that's a bug. Until then, `git diff` the file against `HEAD` to
see what actually changed.

## I want to debug a flake eval failure

```bash
cd ~/.config/glixos
nix flake show
nix eval .#nixosConfigurations.<host>.config.environment.systemPackages \
  --apply 'pkgs: builtins.length pkgs'
```

Standard Nix tooling works against glixos repos without modification.

## More

Open an issue at [github.com/donnismoore/glixos/issues](https://github.com/donnismoore/glixos/issues). Include
`glix doctor` output and the contents of the relevant `hosts/<host>/glix.toml`.
