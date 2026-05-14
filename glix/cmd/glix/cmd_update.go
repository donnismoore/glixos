package main

import (
	"fmt"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

func cmdUpdate(args []string) error {
	fs := newFlagSet("update")
	host := fs.String("host", currentHostname(), "host whose manifest to consult for validation")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after locking")
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

	// If specific inputs are requested, validate them against the manifest's
	// package list and the well-known top-level inputs.
	inputs := fs.Args()
	if len(inputs) > 0 {
		m, err := manifest.Load(r.ManifestPath(*host))
		if err != nil {
			return err
		}
		known := map[string]struct{}{
			"nixpkgs":      {},
			"glixos-core":  {},
			"home-manager": {},
		}
		for name := range m.Packages {
			known[name] = struct{}{}
		}
		for _, in := range inputs {
			if _, ok := known[in]; !ok {
				return fmt.Errorf("unknown input %q (not a package and not a top-level input)", in)
			}
		}
	}

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}
	if err := nix.FlakeUpdate(r.Root, inputs...); err != nil {
		if rerr := snap.Restore(); rerr != nil {
			return fmt.Errorf("nix flake update failed (%w) and rollback also failed: %v", err, rerr)
		}
		return fmt.Errorf("nix flake update failed; rolled back: %w", err)
	}

	msg := "glix update " + *host + ": all inputs"
	if len(inputs) > 0 {
		msg = fmt.Sprintf("glix update %s: %v", *host, inputs)
	}
	if err := r.Commit(msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if len(inputs) == 0 {
		fmt.Println("updated all flake inputs")
	} else {
		fmt.Printf("updated inputs: %v\n", inputs)
	}

	m, err := manifest.Load(r.ManifestPath(*host))
	if err != nil {
		return err
	}
	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	return nil
}
