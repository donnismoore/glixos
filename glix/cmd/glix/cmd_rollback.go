package main

import (
	"fmt"

	"github.com/glixos/glix/internal/nix"
)

// cmdRollback implements ADR-008: undo the most recent glix-managed
// manifest commit via `git revert HEAD`, then relock the flake. With
// --apply, also runs `nixos-rebuild switch` so the system follows the
// reverted manifest.
func cmdRollback(args []string) error {
	fs := newFlagSet("rollback")
	host := fs.String("host", currentHostname(), "host to relock against")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after revert")
	if err := fs.Parse(args); err != nil {
		return err
	}

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	if !r.HostExists(*host) {
		return fmt.Errorf("host %q not found in %s", *host, r.Root)
	}
	if !r.HasGit() {
		return fmt.Errorf("rollback requires a git repo at %s", r.Root)
	}
	clean, err := r.IsClean()
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("working tree is dirty; commit or discard changes before rollback")
	}

	subj, err := r.HeadSubject()
	if err != nil {
		return fmt.Errorf("read HEAD: %w", err)
	}
	fmt.Printf("reverting: %s\n", subj)

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}
	if err := r.RevertHead(); err != nil {
		_ = snap.Restore()
		return fmt.Errorf("git revert: %w", err)
	}
	// `git revert` already commits, so the manifest file on disk now
	// reflects the prior state. Regenerate the flake regions from it so
	// flake.nix matches and run lock to fix flake.lock.
	if err := regenerateFlake(r); err != nil {
		_ = snap.Restore()
		return err
	}
	if err := nix.FlakeLock(r.Root); err != nil {
		if rerr := snap.Restore(); rerr != nil {
			return fmt.Errorf("nix flake lock failed (%w) and rollback also failed: %v", err, rerr)
		}
		return fmt.Errorf("nix flake lock failed after revert; restored on-disk files: %w", err)
	}
	// Amend the revert commit so flake.nix/flake.lock changes are bundled
	// with it. We do this by creating a second commit ("relock after
	// rollback") rather than amending to keep history append-only.
	if err := r.Commit(fmt.Sprintf("glix rollback %s: relock after revert", *host)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Println("rolled back to previous manifest")
	if *apply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	fmt.Println("run `glix rebuild` to apply, or `nixos-rebuild switch --rollback` to switch to the prior generation directly.")
	return nil
}
