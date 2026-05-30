# glix

The glixos CLI. A small Go program that mutates one TOML file plus the
two anchored regions of one generated `flake.nix`, and shells out to
`nix` / `nixos-rebuild` for everything else.

Module path: `github.com/glixos/glix`.

## What's here

```
glix/
├── go.mod
├── cmd/glix/                 # main + one file per subcommand
│   ├── main.go               # command registry, dispatch, usage
│   ├── util.go               # resolveRepo, currentHostname, requireValidIdent
│   ├── registry_helpers.go
│   ├── cmd_init.go
│   ├── cmd_add.go
│   ├── cmd_remove.go
│   ├── cmd_enable.go
│   ├── cmd_set.go
│   ├── cmd_show.go
│   ├── cmd_list.go
│   ├── cmd_info.go
│   ├── cmd_search.go
│   ├── cmd_update.go
│   ├── cmd_rebuild.go
│   ├── cmd_rollback.go
│   ├── cmd_gc.go
│   └── cmd_doctor.go
└── internal/
    ├── manifest/             # glix.toml read/write (deterministic encode)
    ├── flake/                # anchored-region patcher; input/host renderers
    ├── nix/                  # exec wrappers: FlakeLock, FlakeUpdate, Rebuild, GC
    ├── registry/             # JSON fetch + on-disk cache
    ├── repo/                 # Repo handle; git auto-commit; in-memory Snapshot
    ├── resolver/             # URI → name resolution chain
    ├── templates/            # //go:embed-ed flake.nix.tmpl, host.nix.tmpl, glix.toml.tmpl
    └── ui/                   # tiny table renderer
```

## Build

```bash
cd glix
go build -o glix ./cmd/glix
./glix version
# → 0.1.0-m7
```

## Test

```bash
cd glix
go test ./...
go vet ./...
go test -race ./...
```

All tests are unit tests next to their packages. There is no separate
test tree.

## Run on demand

```bash
nix run github:donnismoore/glixos -- version
```

## Adding a command

1. Drop `cmd_<verb>.go` in `cmd/glix/`.
2. Add it to the `commands` slice in `main.go`.
3. If it mutates state, follow the pattern in `cmd_set.go`:
   `TakeSnapshot` → mutate → on any failure `Restore` → on success
   `r.Commit(...)`.

See the [CLI internals](https://donnismoore.github.io/glixos/contributor/cli-internals)
guide for the full pattern.

## See also

- [Architecture](https://donnismoore.github.io/glixos/contributor/architecture)
- [Codebase tour](https://donnismoore.github.io/glixos/contributor/codebase-tour)
- [Command reference (user docs)](https://donnismoore.github.io/glixos/user/commands)
