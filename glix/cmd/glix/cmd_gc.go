package main

import "github.com/glixos/glix/internal/nix"

// cmdGC is a thin wrapper around nix-collect-garbage.
func cmdGC(args []string) error {
	fs := newFlagSet("gc")
	deleteOld := fs.Bool("delete-old", false, "also remove old generations (-d)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return nix.CollectGarbage(*deleteOld)
}
