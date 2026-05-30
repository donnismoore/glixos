---
title: Installation
sidebar_position: 2
---

# Installation

glixos ships everything from one flake at `github:donnismoore/glixos`:

- **`glix`** — the Go CLI you run from your shell.
- **`nixosModules.glixos`** — the OS layer, pulled in as a flake input
  by the user-packages repo `glix init` creates. It installs `glix`
  itself into `environment.systemPackages`, so once a glixos host is up
  it always has the CLI.

You install `glix` first to bootstrap. After the first rebuild, the
module keeps it in sync.

## Requirements

- A working NixOS or nix-with-flakes installation:
  - Nix 2.18 or newer.
  - `nix-command` and `flakes` enabled. On non-NixOS systems, add to
    `/etc/nix/nix.conf`:
    ```ini
    experimental-features = nix-command flakes
    ```
- `git` available on `$PATH` (glix auto-commits every mutation).
- Go 1.22+ **only if you're building `glix` from source**.

You do **not** need glixos itself before installing `glix`; the CLI is just
a Go binary that drives nix and nixos-rebuild for you.

## Option A — Already on a glixos host

Nothing to do — `nixosModules.glixos` installs `glix` for you. Verify:

```bash
glix --version
# → glix <ver> (<commit>)
```

If a host should *not* ship `glix` (e.g. minimal appliances), set
`glixos.glix.enable = false` in that host's module.

## Option B — `nix run` (no install)

```bash
nix run github:donnismoore/glixos -- version
```

This builds `glix` on the fly and runs it. Drop the `-- version` for
interactive use, but for repeated work you'll want a real install.

## Option C — `nix shell` (ephemeral)

```bash
nix shell github:donnismoore/glixos
glix version
```

Exit the shell when you're done; nothing is installed permanently.

## Option D — Build `glix` from source

```bash
git clone https://github.com/donnismoore/glixos.git
cd glixos/glix
go build -o glix ./cmd/glix
sudo install -m 0755 glix /usr/local/bin/glix
glix version
# → glix 0.1.0-m7 (unknown)
```

To install for just your user:

```bash
go install ./cmd/glix
# ensure $(go env GOBIN) (or $HOME/go/bin) is on $PATH
```

The `(unknown)` commit field appears because `go build` doesn't stamp
the revision; the Nix build does.

## Verify

```bash
glix doctor
```

Expected output is a series of `[OK]` lines covering nix, flakes, repo
discovery, and per-host manifests (none yet on a fresh machine). `WARN`
about a missing repo is fine on first install — that's what `glix init`
fixes.

## Next

Head to [Quickstart](./quickstart) to bootstrap your first host.
