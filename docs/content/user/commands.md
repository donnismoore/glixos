---
title: Command reference
sidebar_position: 4
---

# Command reference

Every command takes `--dir <path>` (override repo discovery) and most take
`--host <name>` (default: `$(hostname)`). Run `glix <command> -h` for the
authoritative flag list.

## Bootstrap

### `glix init`

Bootstrap a user-packages flake repo and seed a host.

```
glix init [--host=<name>] [--user=<name>] [--system=<tuple>]
          [--dir=<path>] [--core=<flake-url>]
```

| Flag       | Default                          | Notes                                    |
|------------|----------------------------------|------------------------------------------|
| `--host`   | `$(hostname)`                    | Host subdirectory under `hosts/`         |
| `--user`   | `$USER`                          | Primary user (`Settings.primary_user`)   |
| `--system` | `x86_64-linux`                   | Nix system tuple                         |
| `--dir`    | `$XDG_CONFIG_HOME/glixos`        | Repo root                                |
| `--core`   | `github:donnismoore/glixos`     | glixos repo flake URL                    |

`glix init` is **idempotent for new hosts**. Running it again with a
different `--host` on an existing repo adds that host without touching the
existing one.

## Mutation

### `glix add`

```
glix add <ref-or-name> [--scope=system|home] [--user=<name>]
                       [--pin=<rev>] [--name=<override>]
                       [--apply] [--dry-run]
                       [--registry-url=<url>] [--refresh]
                       [--no-nix-registry]
```

| Flag                | Meaning                                                          |
|---------------------|------------------------------------------------------------------|
| `--scope`           | `system` (default from `[settings] default_scope`) or `home`     |
| `--user`            | Only for `scope=home`; defaults to `Settings.primary_user`       |
| `--pin`             | Lock to a revision via `?rev=<pin>` (github/gitlab/sourcehut)    |
| `--name`            | Override the inferred package name                               |
| `--apply`           | Run `nixos-rebuild switch` after staging                         |
| `--dry-run`         | Print the planned change without writing                         |
| `--registry-url`    | Override the registry URL for this invocation                    |
| `--refresh`         | Force a refetch of the registry                                  |
| `--no-nix-registry` | Skip the `nix registry list` fallback                            |

Resolution chain (ADR-004): URI passthrough → glixos registry →
`nix registry list` → error with suggestions.

### `glix remove <name>`

Removes the package from `glix.toml` and the `flake.nix` inputs region,
relocks, and commits. Flags: `--apply`, `--dry-run`, `--host`, `--dir`.

### `glix enable <name>` / `glix disable <name>`

Toggle the `enabled` flag. Disabled packages stay in the manifest but
contribute nothing to the generated module list.

### `glix set <name> <key>=<value>...`

Mutate one or more fields on an existing package. Multiple `key=value`
pairs apply atomically.

Keys:

| Key                  | Effect                                                       |
|----------------------|--------------------------------------------------------------|
| `flake=<uri>`        | Replace the flake reference                                  |
| `scope=system\|home` | Move between scopes                                          |
| `enabled=true\|false`| Same as `glix enable`/`disable`                              |
| `pin=<rev>`          | Set the revision pin (empty value clears)                    |
| `user=<name>`        | For home-scope packages (empty value clears)                 |
| `config.<key>=<val>` | Set a per-package config entry (empty value deletes the key) |

Examples:

```bash
glix set greeting scope=home user=alice
glix set greeting config.message="Hello, world!"
glix set greeting config.message=        # delete the key
glix set ferment pin=v1.2.3 enabled=true # multiple in one commit
```

### `glix update [<input>...]`

Run `nix flake update`. With no args, updates every input. With named
inputs, updates only those.

Recognised input names:

- Any package name from the manifest.
- `nixpkgs`, `home-manager`, `glixos-core`.

Flags: `--apply`, `--host`, `--dir`.

## Inspection

### `glix list`

```
glix list [--host=<name>] [--all] [--all-hosts]
```

| Flag           | Meaning                                          |
|----------------|--------------------------------------------------|
| `--all`        | Include disabled packages                        |
| `--all-hosts`  | Show every host (adds a `HOST` column)           |

### `glix show <name>`

Per-package detail: flake, scope, state, user (for home), pin, host, and
the config table if non-empty.

### `glix info`

Repo summary: root, git status and HEAD subject, per-host stats
(`system`, `primary_user`, enabled/total package counts), and the
resolved `flake.lock` inputs.

### `glix search <query>`

Substring match against the cached registry. `--refresh` forces a refetch.

### `glix doctor`

Read-only environment and repo health checks. Exits non-zero if any
check is `FAIL`.

## System

### `glix rebuild [action]`

Wraps `nixos-rebuild --flake .#<host> <action>`. Actions: `switch`
(default), `boot`, `test`, `build`, `dry-build`, `dry-activate`.

### `glix rollback`

Revert the last manifest commit and relock. See
[Rollback](./rollback) for the full story (and the relationship with
`nixos-rebuild --rollback`).

### `glix gc [-delete-old]`

Wraps `nix-collect-garbage`. With `-delete-old`, passes `-d` to also
remove old generations.

## Meta

### `glix version`

Prints the CLI version and the commit it was built from
(e.g. `glix 0.1.0-m7 (abc1234)`). `glix --version` and `glix -v` are
aliases. When `glix` is built outside Nix (`go build` directly) the
commit field reads `unknown`.

### `glix help`

Lists every command with a one-line summary. `glix <cmd> -h` shows flags
for a specific command.
