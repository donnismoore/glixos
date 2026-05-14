package main

import (
	"fmt"

	"github.com/glixos/glix/internal/nix"
)

func cmdRebuild(args []string) error {
	fs := newFlagSet("rebuild")
	host := fs.String("host", currentHostname(), "host to build")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	action := "switch"
	if fs.NArg() > 0 {
		action = fs.Arg(0)
	}
	switch action {
	case "switch", "boot", "test", "build", "dry-build", "dry-activate":
	default:
		return fmt.Errorf("unknown rebuild action %q (want switch, boot, test, build, dry-build, dry-activate)", action)
	}

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	if !r.HostExists(*host) {
		return fmt.Errorf("host %q not found in %s", *host, r.Root)
	}
	return nix.Rebuild(r.Root, *host, action)
}
