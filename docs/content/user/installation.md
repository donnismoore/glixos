---
title: Installation
sidebar_position: 2
---

# Installation

glixos is two things you install separately:

- **`glix`** — the Go CLI you run from your shell.
- **`glixos-core`** — the OS layer, pulled in as a flake input by the
  user-packages repo `glix init` creates.

You install `glix` first; it bootstraps the rest.

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

## Option A — Build `glix` from source

```bash
git clone https://github.com/powerreddude/glixos.git
cd glixos/glix
go build -o glix ./cmd/glix
sudo install -m 0755 glix /usr/local/bin/glix
glix version
# → 0.1.0-m7
```

To install for just your user:

```bash
go install ./cmd/glix
# ensure $(go env GOBIN) (or $HOME/go/bin) is on $PATH
```

## Option B — `nix run`

If you don't want a permanent install:

```bash
nix run github:powerreddude/glixos?dir=glix -- version
```

This builds `glix` on the fly and runs it. Drop the `-- version` for
interactive use, but for repeated work you'll want a real install.

## Option C — Run from a Nix shell

```bash
nix shell github:powerreddude/glixos?dir=glix
glix version
```

`glix shell` exits the shell when you're done; nothing is installed
permanently.

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
