package main

import (
	"fmt"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

func cmdEnable(args []string) error  { return setEnabled("enable", true, args) }
func cmdDisable(args []string) error { return setEnabled("disable", false, args) }

func setEnabled(verb string, want bool, args []string) error {
	fs := newFlagSet(verb)
	host := fs.String("host", currentHostname(), "host whose manifest to mutate")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after staging")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: glix %s <name>", verb)
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
	pkg, ok := m.Packages[name]
	if !ok {
		return fmt.Errorf("package %q not in manifest", name)
	}
	if pkg.Enabled == want {
		fmt.Printf("package %q already %sd\n", name, verb)
		return nil
	}

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}
	pkg.Enabled = want
	m.Packages[name] = pkg
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
	if err := r.Commit(fmt.Sprintf("glix %s %s: %s", verb, *host, name)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Printf("%sd %s on host %s\n", verb, name, *host)
	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	return nil
}
