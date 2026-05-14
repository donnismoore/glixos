---
title: Rollback
sidebar_position: 12
---

# Rollback

glixos has two layers of recovery; both are designed so that the manifest
and the running system never disagree silently.

See [ADR-008](../adr/ADR-008-rollback-coupling).

## Layer 1: in-flight transactional rollback

Every mutating glix command (`add`, `remove`, `enable`, `disable`, `set`,
`update`) follows the same pattern:

1. Take an in-memory snapshot of `glix.toml`, `flake.nix`, and
   `flake.lock`.
2. Write the desired changes.
3. Run `nix flake lock`.
4. If step 3 fails, restore from the snapshot and return the error.
5. If everything succeeded, `git add -A && git commit -m "..."`.

You get a working tree every time, even on failures. Nothing for you to
do.

## Layer 2: `glix rollback`

If a *successful* mutation produces a system you don't want, run:

```bash
glix rollback
```

This reverts the last commit in the user-packages repo (`git revert
HEAD`) and runs `nix flake lock` against the restored manifest. Now
`flake.nix` and `glix.toml` describe the system as it was before the
last change.

To also roll back the running generation:

```bash
sudo nixos-rebuild --rollback switch
glix rollback
```

Order matters: if you `glix rollback` first, then your manifest matches
the previous-previous generation, which is fine if you also rebuild.
The general rule: **reach a clean state on both sides** before doing
anything else.

You can also chain them:

```bash
glix rollback --apply       # revert manifest + nixos-rebuild switch
```

`--apply` runs `nixos-rebuild switch` after the revert, so the running
system tracks the manifest.

## What `glix rollback` does **not** do

- It does not call `nixos-rebuild --rollback` automatically. That's a
  separate, system-level operation; you may want a generation rollback
  without touching the manifest, or vice versa. Layer 2 keeps the two
  decisions independent.
- It does not revert through arbitrary history. Repeat
  `glix rollback` to step back multiple commits. For surgical changes,
  use `git revert <sha>` followed by `nix flake lock`.

## Auditing your history

Every glix mutation is one commit with a descriptive subject:

```
glix add laptop: pkg-greeting (path:..., home, via uri)
glix set laptop: pkg-greeting (config.message=Hello)
glix remove laptop: pkg-hello
glix update laptop: all inputs
```

```bash
cd ~/.config/glixos
git log --oneline
```

This is the audit trail. Treat it like a normal git history; `git revert`,
`git diff`, and tagging all behave normally.
