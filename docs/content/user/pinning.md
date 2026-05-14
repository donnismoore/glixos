---
title: Pinning
sidebar_position: 11
---

# Pinning

Two ways to lock a package to a specific revision:

1. **`flake.lock`** — the canonical pinning mechanism. `nix flake lock`
   resolves each input to a specific revision and stores it. glix relocks
   on every mutation.
2. **`Package.pin`** — a hint glix folds into the flake URI as
   `?rev=<pin>` before locking. Useful when you want the manifest itself
   to advertise the pin, or when you want to force a specific revision
   regardless of upstream branch tip.

## When to use which

| Goal                                                       | Mechanism                                                                                  |
|------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| Reproducible builds across machines                        | `flake.lock` (commit it; that's what glix does for you)                                    |
| Pin a single package to a specific tag/SHA in the manifest | `Package.pin` (set with `glix add --pin=` or `glix set <name> pin=`)                       |
| Roll a package forward                                     | `glix update <name>` (refreshes `flake.lock` for that input)                               |
| Roll a package back                                        | `glix set <name> pin=<old-sha>` *or* edit `flake.lock` and `glix rebuild`                  |

## `Package.pin`

```bash
glix add --pin=v1.2.3 github:owner/cool-tool
```

Resulting manifest:

```toml
[packages.cool-tool]
flake   = "github:owner/cool-tool"
scope   = "system"
enabled = true
pin     = "v1.2.3"
```

When glix renders the inputs region of `flake.nix`, it emits:

```nix
cool-tool = {
  url = "github:owner/cool-tool?rev=v1.2.3";
};
```

This works for ref types where nix understands `?rev=`:

- `github:owner/repo`
- `gitlab:owner/repo`
- `sourcehut:owner/repo` / `srht:owner/repo`

For other ref types (`path:`, `git+ssh://`, `http://`), `Package.pin` is
preserved in the manifest but **not** folded into the URL — those refs
don't have a `?rev=` semantic in nix. Use `flake.lock` for them.

## Pinning failures are transactional

If you supply a malformed pin, `nix flake lock` will refuse. glix catches
the failure and restores the manifest, `flake.nix`, and `flake.lock` from
an in-memory snapshot before returning the error. You don't end up with a
half-applied change.

```
$ glix set cool-tool pin=garbage
glix set: nix flake lock failed; rolled back: ...
```

## Auditing pins

```bash
glix info
```

```
flake.lock:
  glixos-core         github:owner/glixos
  home-manager        github:nix-community/home-manager @ 9760b31dab30
  nixpkgs             github:NixOS/nixpkgs @ da5ad661ba4e
  cool-tool           github:owner/cool-tool @ v1.2.3
```

`glix info` reads `flake.lock` directly, so it's the authoritative view of
what's actually pinned.
