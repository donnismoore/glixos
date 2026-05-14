---
title: Build & test
sidebar_position: 11
---

# Build & test

How to build glixos and run its tests locally.

## Prerequisites

- Go 1.22+.
- Nix 2.18+ with `nix-command` and `flakes` enabled.
- Node 18+ (only if you're working on the docs site).

## The Go CLI

```bash
cd glix
go build -o glix ./cmd/glix
./glix version
```

Quick install for your user:

```bash
cd glix
go install ./cmd/glix
# ensure $(go env GOBIN) is on $PATH; defaults to ~/go/bin
```

## Running tests

```bash
cd glix
go test ./...
```

All packages have unit tests. Network-touching tests (the registry HTTP
fetch) skip themselves if no network is available; you can force a full
run with `GLIX_TEST_NET=1 go test ./...`.

`go vet ./...` and `go test -race ./...` are both expected to pass on
every push.

## The Nix core

```bash
cd core
nix flake check          # build the smoke-test host
nix build .#vm            # produces a runnable QEMU image
./result/bin/run-*-vm     # boot it
```

`nix fmt` formats every Nix file via `nixpkgs-fmt`.

## End-to-end smoke test

A throwaway, full-stack exercise that's safe to run anywhere:

```bash
TMP=$(mktemp -d)
go build -o $TMP/glix ./glix/cmd/glix

$TMP/glix init --dir $TMP/repo --host laptop --user me --system x86_64-linux \
  --core path:$(pwd)/core

cd $TMP/repo
$TMP/glix add --scope home path:$(pwd)/../../examples/pkg-greeting
$TMP/glix set greeting config.message="hello from CI"
$TMP/glix list
$TMP/glix info

nix eval --raw .#nixosConfigurations.laptop.config.home-manager.users.me.home.file.\".glixos-greeting\".text
```

This is roughly what the integration testing of M6/M7 exercised.

## The docs site

```bash
cd docs
npm install
npm start    # http://localhost:3000
```

To build a static bundle:

```bash
cd docs
npm run build
npm run serve   # serves ./build on localhost:3000
```

The CI workflow at `.github/workflows/docs.yml` builds and publishes to
GitHub Pages on every push to `main`.

## Pre-commit checklist

```bash
go test ./...            # all Go tests pass
go vet ./...             # no vet complaints
nix flake check          # core flake builds
nix flake check path:./examples/pkg-hello
nix flake check path:./examples/pkg-greeting
cd docs && npm run build # docs site builds
```

## Versioning

The CLI version lives in `glix/cmd/glix/main.go` as a `const version`.
Bump it when shipping a milestone. The pattern is
`0.1.0-m<milestone>` until 1.0.

## Tagging a release

```bash
git tag v0.1.0-m7 -m "M7: per-package config and repo introspection"
git push --tags
```

A future release workflow will build `glix` binaries and publish them
alongside the tag.
