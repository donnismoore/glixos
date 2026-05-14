---
title: Snapshots
sidebar_position: 9
---

# Snapshots

How glix keeps the working tree consistent when a mutation fails mid-way.

Location: `glix/internal/repo/snapshot.go`.

See [ADR-008](../adr/ADR-008-rollback-coupling).

## The problem

A `glix add` does several side effects in order:

1. Write `glix.toml`.
2. Rewrite anchored regions of `flake.nix`.
3. Run `nix flake lock` (which mutates `flake.lock`).

Step 3 can fail for reasons outside glix's control: a malformed pin, a
network blip, an upstream flake with broken inputs. If we stop at the
failure, the user is left with a manifest that mentions a package
`flake.nix` doesn't have a working input for. Worse, half the inputs
might be locked and half not.

## The mechanism

```go
type Snapshot struct {
    manifestPath string
    manifest     []byte // nil if file did not exist

    flakePath    string
    flake        []byte

    lockPath     string
    lock         []byte // nil if file did not exist
}

func (r *Repo) TakeSnapshot(host string) (*Snapshot, error)
func (s *Snapshot) Restore() error
```

`TakeSnapshot` reads each file into memory at the moment of capture.
`Restore` writes them back atomically, recreating or deleting as needed
so the file system reaches exactly the captured state.

`Restore` is intentionally **best-effort and idempotent**: each file is
written via the atomic tmp+rename dance. A failed restore returns an
error but does not abort the rollback for the other files.

## The usage contract

Every mutating command:

```go
snap, err := r.TakeSnapshot(*host)
if err != nil { return err }

if err := manifest.Save(...); err != nil {
    _ = snap.Restore()
    return err
}
if err := regenerateFlake(r); err != nil {
    _ = snap.Restore()
    return err
}
if err := nix.FlakeLock(r.Root); err != nil {
    if rerr := snap.Restore(); rerr != nil {
        return fmt.Errorf("nix flake lock failed (%w) and rollback also failed: %v", err, rerr)
    }
    return fmt.Errorf("nix flake lock failed; rolled back: %w", err)
}
// success: commit
```

After `r.Commit` succeeds, the snapshot is discarded implicitly (no
finalizer needed; it's just memory).

## What snapshots do **not** cover

- They don't touch git history. If you've already `r.Commit`-ed, use
  `glix rollback` (which does `git revert HEAD`) instead.
- They don't track the running NixOS generation. `nixos-rebuild
  --rollback` is the only mechanism for that.
- They don't snapshot the registry cache. That cache is read-only
  during mutations.

## Tests

`glix/internal/repo/snapshot_test.go` covers:

- Round-trip: capture → modify → restore → contents match capture.
- Files that didn't exist at capture time are deleted on restore.
- Files that existed but had been deleted come back on restore.
- Atomic write characteristics (no partial writes visible).

If you add a new piece of state that glix mutates (a new file, a new
region), update `Snapshot` to capture it. The cost is small; the
guarantee is "every mutation is fully reversible until the commit
succeeds".
