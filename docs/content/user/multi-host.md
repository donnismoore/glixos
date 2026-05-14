---
title: Multi-host
sidebar_position: 9
---

# Multi-host

One user-packages repo can describe any number of machines. Each host has
its own subdirectory under `hosts/`, its own manifest, and its own
`[settings]` (including system tuple and primary user).

See [ADR-005](../adr/ADR-005-multi-host) for the design rationale.

## Layout

```
~/.config/glixos/
├── flake.nix
├── flake.lock
├── hosts/
│   ├── laptop/
│   │   ├── default.nix
│   │   └── glix.toml          # x86_64-linux, alice
│   ├── desktop/
│   │   ├── default.nix
│   │   └── glix.toml          # x86_64-linux, alice
│   └── pi/
│       ├── default.nix
│       └── glix.toml          # aarch64-linux, pi-user
└── shared/                    # optional, your own modules
    └── base.nix
```

## Adding a host

```bash
glix init --host=desktop --user=alice --system=x86_64-linux
glix init --host=pi --user=pi-user --system=aarch64-linux
```

`glix init` is host-additive: it leaves existing hosts alone and registers
the new one in the glix-managed hosts region of `flake.nix`.

## Operating on a specific host

Every read and write command takes `--host`:

```bash
glix add --host=pi --scope=system raspi-firmware
glix list --host=desktop
glix show --host=pi raspi-firmware
glix rebuild --host=desktop switch
```

If you omit `--host`, the default is `$(hostname)` on the current machine.

## Listing across hosts

```bash
glix list --all-hosts
```

```
HOST     NAME       SCOPE   STATE    USER   FLAKE
desktop  firefox    system  enabled         github:NixOS/nixpkgs#firefox
laptop   firefox    system  enabled         github:NixOS/nixpkgs#firefox
laptop   greeting   home    enabled  alice  github:owner/pkg-greeting
pi       raspi-fw   system  enabled         github:owner/raspi-firmware
```

```bash
glix info
```

```
hosts:
  desktop      system=x86_64-linux primary_user=alice    packages=12/14 enabled
  laptop       system=x86_64-linux primary_user=alice    packages=10/11 enabled
  pi           system=aarch64-linux primary_user=pi-user packages=3/3   enabled
```

## Sharing modules

Anything outside `hosts/<host>/` is yours to organise. A common pattern is
a `shared/` directory imported from each host's `default.nix`:

```nix
# hosts/laptop/default.nix
{ ... }:
{
  imports = [ ../../shared/base.nix ];
  networking.hostName = "laptop";
  # ...
}
```

```nix
# shared/base.nix
{ pkgs, ... }:
{
  time.timeZone = "America/Los_Angeles";
  environment.systemPackages = with pkgs; [ git tmux ];
}
```

glix never edits `shared/` or the parts of `hosts/<host>/default.nix`
outside the glix-managed regions of `flake.nix`. Everything is yours
to refactor.

## Building a remote host

`nixos-rebuild` supports `--target-host` and `--build-host`. Pass them
through your usual NixOS workflow; glix only owns the manifest side.

```bash
nixos-rebuild switch --flake .#pi --target-host root@pi.local
```

## Mixing architectures

The Nix system tuple lives in each host's `[settings] system`. `glix init`
writes it; you can change it later by editing `glix.toml` directly (no
mutation flag yet) and running any glix command to trigger a flake regen.
