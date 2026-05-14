package main

import (
	"fmt"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

func cmdRemove(args []string) error {
	fs := newFlagSet("remove")
	host := fs.String("host", currentHostname(), "host whose manifest to mutate")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after staging")
	dryRun := fs.Bool("dry-run", false, "print planned changes; do not write")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: glix remove <name> [flags]")
	}
	name := fs.Arg(0)

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	if !r.HostExists(*host) {
		return fmt.Errorf("host %q not found in %s", *host, r.Root)
	}

	m, err := manifest.Load(r.ManifestPath(*host))
	if err != nil {
		return err
	}
	if _, ok := m.Packages[name]; !ok {
		return fmt.Errorf("package %q not in manifest", name)
	}

	if *dryRun {
		fmt.Printf("would remove package %q from %s\n", name, r.ManifestPath(*host))
		return nil
	}

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}

	delete(m.Packages, name)
	if err := manifest.Save(r.ManifestPath(*host), m); err != nil {
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
	if err := r.Commit(fmt.Sprintf("glix remove %s: %s", *host, name)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Printf("removed %s from host %s\n", name, *host)

	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	return nil
}
