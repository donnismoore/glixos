# glixos

> Modular, flake-driven NixOS with a manifest you can read.

**glixos** is a NixOS distribution organised around two ideas:

1. **One file declares your system.** A single `glix.toml` per host lists
   every package you've installed, what scope it lives in (system or
   home), which user a home-scope package targets, and any per-package
   config. Everything else is static Nix that reads that file.
2. **A small Go CLI is the only writer.** `glix` mutates `glix.toml` and
   an anchored region of one generated `flake.nix`. It never parses Nix;
   it never edits user code. Every mutation is one git commit, so
   rollback is trivial.

| | |
|---|---|
| **Status** | Pre-release. CLI version `0.1.0-m7`. Schema `1`. |
| **Docs** | [powerreddude.github.io/glixos](https://powerreddude.github.io/glixos/) |
| **License** | [GPLv3](./LICENSE) |

---

## Quickstart

### 1. Install `glix`

On an existing glixos host, `nixosModules.glixos` already installs `glix`
into `environment.systemPackages`; rebuild and you're done. To get
`glix` on a machine that isn't a glixos host yet, pick one:

```bash
# Run on demand (no install).
nix run github:powerreddude/glixos -- version

# Build from source.
git clone https://github.com/powerreddude/glixos.git
cd glixos/glix
go build -o glix ./cmd/glix
sudo install -m 0755 glix /usr/local/bin/glix
glix version
# → glix 0.1.0-m7 (<commit>)
```

`glix` requires Nix 2.18+ with `nix-command` and `flakes` enabled, plus
`git` on `$PATH`.

### 2. Bootstrap a host

```bash
glix init \
  --host laptop \
  --user alice \
  --system x86_64-linux
```

This creates `~/.config/glixos/` with a `flake.nix`,
`hosts/laptop/{default.nix,glix.toml}`, runs `git init`, and makes the
first commit.

### 3. Add a package

```bash
cd ~/.config/glixos
glix add github:powerreddude/glixos?dir=examples/pkg-hello
glix list
# NAME      SCOPE   STATE    USER  FLAKE
# pkg-hello system  enabled        github:powerreddude/glixos?dir=examples/pkg-hello
```

### 4. Rebuild

```bash
sudo glix rebuild switch
```

This wraps `nixos-rebuild --flake .#laptop switch`. If something goes
wrong, recover with:

```bash
sudo nixos-rebuild --rollback switch
glix rollback
```

The [Quickstart guide](https://powerreddude.github.io/glixos/user/quickstart)
goes into more detail.

---

## What's in this repo

```
glixos/
├── flake.nix            # repo-root flake: nixosModules, lib, packages.glix
├── core/                # the OS layer (modules, lib, smoke-test host)
├── glix/                # the Go CLI
├── registry/            # default registry (registry.json)
├── examples/            # reference package flakes
│   ├── pkg-hello/
│   ├── pkg-greeting/
│   └── user-packages/
├── docs/                # Docusaurus documentation site
└── LICENSE              # GNU GPLv3
```

Each subdirectory has its own README explaining its role.

---

## Documentation

The full docs are deployed to GitHub Pages:

**[powerreddude.github.io/glixos](https://powerreddude.github.io/glixos/)**

Two paths:

- **[Users](https://powerreddude.github.io/glixos/user/getting-started)** —
  install, manage packages, share modules between hosts, per-package
  config, rollback, GC, troubleshooting.
- **[Contributors](https://powerreddude.github.io/glixos/contributor/overview)** —
  architecture, design principles, codebase tour, CLI internals, the
  Nix layer, build & test, release process.

ADRs live alongside the docs at
[powerreddude.github.io/glixos/adr](https://powerreddude.github.io/glixos/adr/).

To run the docs locally:

```bash
cd docs
npm install
npm start    # http://localhost:3000
```

---

## License

glixos is licensed under the GNU General Public License v3.0. See
[LICENSE](./LICENSE) for the full text. By contributing, you agree to
license your contributions under the same terms.
