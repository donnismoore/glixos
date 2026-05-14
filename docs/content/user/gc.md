---
title: Garbage collection
sidebar_position: 13
---

# Garbage collection

glixos uses Nix's standard garbage collector. `glix gc` is a thin wrapper.

## Common use

Free space without deleting old generations:

```bash
glix gc
```

This is equivalent to `nix-collect-garbage`. Old generations stay, which
means you can still roll back via `nixos-rebuild --rollback`.

Free more space by also removing old generations:

```bash
sudo glix gc -delete-old
```

Equivalent to `nix-collect-garbage -d`. After this, only the current
generation is left; rollback is no longer possible.

## When to run it

- After a `glix update` that pulled significant nixpkgs deltas.
- After removing a large package that you don't plan to reinstall.
- On low-disk-space alerts.

You don't need to run `glix gc` after every mutation. Generations are
cheap; they only consume disk for whatever store paths they reference
that aren't reachable from the current generation.

## Scheduling

If you want garbage collection on a timer, NixOS has it natively:

```nix
# hosts/laptop/default.nix
nix.gc = {
  automatic = true;
  dates = "weekly";
  options = "--delete-older-than 30d";
};
```

That's a host-level concern (not a glix package), so it belongs in
`hosts/<host>/default.nix`, not in `glix.toml`.

## Verifying

```bash
nix-store --gc --print-roots | head
df -h /nix
```
