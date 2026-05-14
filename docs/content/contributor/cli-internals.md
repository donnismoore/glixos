---
title: CLI internals
sidebar_position: 5
---

# CLI internals

How the Go side is put together at the level you need to add features.

## Dispatch

`cmd/glix/main.go` holds a `[]command` slice. Each entry has a name,
summary, and a `func(args []string) error`. `main()` matches the first
argv and dispatches.

To add a new command:

1. Create `cmd/glix/cmd_<verb>.go` with `func cmdVerb(args []string) error`.
2. Add an entry to the `commands` slice in `main.go`.
3. That's it.

## Flag parsing

Every command uses `newFlagSet(name)` from `main.go`, which returns a
`*flag.FlagSet` configured to print usage to stderr. Stick to the stdlib
`flag` package; we don't use cobra.

Positional arguments come **after** flags in `args`. The `flag` package
stops at the first non-flag argument, so callers can mix flags and
positionals as long as flags come first. Document this in the command's
usage string.

## The mutation pattern

Every command that changes state follows the same shape:

```go
// 1. Parse flags and positionals.
// 2. Resolve the repo, validate input.
// 3. Load the manifest.
// 4. Compute the new manifest.
// 5. Take a snapshot.
// 6. Save manifest → regenerate flake → flake lock.
//    On any failure, snapshot.Restore() and return the error.
// 7. r.Commit(message).
// 8. Optionally rebuild if --apply or settings.AutoApply.
```

Read `cmd_set.go` for the canonical example.

## Snapshots

```go
snap, err := r.TakeSnapshot(*host)
if err != nil { return err }

// ... write things ...

if somethingFailed {
    _ = snap.Restore()
    return err
}
```

The snapshot captures three files in memory. `Restore()` re-writes them
atomically. The intent is: between `TakeSnapshot` and the final
`r.Commit`, the working tree may be inconsistent; outside that window
it must be a valid glixos repo.

## Commit messages

By convention: `glix <verb> <host>: <summary>`.

Examples:

- `glix init: bootstrap host laptop`
- `glix add laptop: pkg-hello (path:..., system, via uri)`
- `glix remove laptop: pkg-hello`
- `glix set laptop: pkg-hello (config.message=Hello)`
- `glix update laptop: all inputs`

`glix rollback` relies on these only being subject lines — the full
commit message is fine to extend with a body if needed.

## Validation

`cmd/glix/util.go` has `requireValidIdent(kind, name) error` which
checks the name against `^[A-Za-z][A-Za-z0-9_-]*$`. Use it for any
field that becomes a TOML bare key or a Nix attribute name.

For scope validation: `manifest.Scope(v).Valid()`.

For host existence: `r.HostExists(host)`.

## Reading vs writing

Read-only commands (`list`, `show`, `info`, `search`, `doctor`, `version`)
must not write anything, must not invoke git, and must not invoke nix
except for `nix flake metadata`/`nix registry list` calls that don't
modify state. They should not take snapshots.

## Errors

- Return Go errors. `main()` prints them as
  `glix <cmd>: <message>` and exits 1.
- Wrap with `%w` when you want the caller's error chain preserved
  (rarely needed for a CLI).
- Don't `log.Fatal`. Always return.

## Adding a new manifest field

1. Add a field to `manifest.Package` (or `manifest.Settings`) with an
   `omitempty` TOML tag.
2. Extend `Encode` to emit it (sorted, deterministic).
3. Add a round-trip test in `manifest_test.go`.
4. If a command should mutate it, add a `case "field":` to `cmdSet` in
   `cmd_set.go`.
5. Surface it in `cmd_show.go` and (if relevant) `cmd_list.go`.
6. If it changes Nix evaluation, thread it through
   `core/lib/importManifest.nix` and the host template.

This is the M6/M7 playbook (see ADR-009, ADR-010 for examples).

## Adding a new region

If a Nix-side change needs glix to manage a new region of `flake.nix`:

1. Add a `RegionFoo` constant in `internal/flake/patcher.go`.
2. Add the begin/end markers to `templates/flake.nix.tmpl`.
3. Write a `RenderFoo(...) string` builder.
4. Have `regenerateFlake` in `cmd_init.go` pass it to `flake.PatchFile`.

Don't add regions casually; each one is forever, since old user repos
will outlive the addition. Prefer adding fields to existing structures.
