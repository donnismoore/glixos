---
title: Multi-user
sidebar_position: 10
---

# Multi-user

A glixos host can run home-manager configs for any number of users. Each
home-scope package is routed to a specific user; the default is
`[settings] primary_user`, and any package can override with its own
`user` field.

See [ADR-010](../adr/ADR-010-multi-user-routing) for the design.

## Choosing a target user

```bash
# Goes to settings.primary_user (alice on this host).
glix add --scope=home github:owner/pkg-greeting

# Explicit per-package override.
glix add --scope=home --user=bob github:owner/pkg-greeting --name=greeting-bob
glix set greeting-bob config.message="Hi, Bob."
```

The two packages must have distinct names. By convention, suffix the
override package with `-<user>` so it's obvious in `glix list`.

## Making sure the user exists

`importManifest` happily routes packages to whatever user names appear in
the manifest, but home-manager will fail to evaluate if the user isn't
declared in NixOS. The init template declares your primary user:

```nix
# hosts/laptop/default.nix
users.users."@USER@" = {
  isNormalUser = true;
  extraGroups = [ "wheel" "networkmanager" ];
  initialPassword = "changeme";
};
```

For additional users, either:

1. Add a `users.users.bob = { ... };` block in `hosts/<host>/default.nix`,
   or
2. Put it in a `shared/users.nix` module imported from each host.

Either approach is normal NixOS; glix has no opinions here beyond
"the user must exist when you rebuild".

## Listing the routing

```bash
glix list
```

```
NAME           SCOPE   STATE    USER   FLAKE
greeting       home    enabled  alice  github:owner/pkg-greeting
greeting-bob   home    enabled  bob    github:owner/pkg-greeting
firefox        system  enabled         github:NixOS/nixpkgs#firefox
```

```bash
glix show greeting-bob
```

```
name    greeting-bob
flake   github:owner/pkg-greeting
scope   home
state   enabled
user    bob
host    laptop
config
  message = Hi, Bob.
```

## Changing the target user

```bash
glix set greeting user=alice    # set
glix set greeting user=         # clear; fall back to primary_user
```

`glix set` validates that the value is a legal POSIX-ish ident
(`[A-Za-z][A-Za-z0-9_-]*`).

## Inside the manifest

A multi-user manifest looks like:

```toml
[settings]
default_scope = "home"
primary_user  = "alice"

[packages.greeting]
flake   = "github:owner/pkg-greeting"
scope   = "home"
enabled = true
# no user field → routed to alice

[packages.greeting-bob]
flake   = "github:owner/pkg-greeting"
scope   = "home"
enabled = true
user    = "bob"

[packages.greeting-bob.config]
message = "Hi, Bob."
```

## What this generates in Nix

`importManifest` returns a `homeModulesByUser` attrset:

```nix
homeModulesByUser = {
  alice = [ <pkg-greeting module wrapped with glixConfig = {}> ];
  bob   = [ <pkg-greeting module wrapped with glixConfig.message = "Hi, Bob."> ];
}
```

The host template feeds this directly to home-manager:

```nix
home-manager.users = lib.mapAttrs
  (_: mods: { imports = mods ++ [ { home.stateVersion = "24.11"; } ]; })
  manifest.homeModulesByUser;
```

So adding a new user takes nothing more than declaring them in NixOS and
running `glix add --user=<name>`.
