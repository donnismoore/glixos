package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

// cmdSet supports mutating individual fields on an existing package:
//
//	glix set <name> scope=home
//	glix set <name> flake=github:owner/repo enabled=false
//
// Multiple key=value pairs are applied atomically.
func cmdSet(args []string) error {
	fs := newFlagSet("set")
	host := fs.String("host", currentHostname(), "host whose manifest to mutate")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after staging")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("usage: glix set <name> <key=value>...")
	}
	name := fs.Arg(0)
	assignments := fs.Args()[1:]

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

	for _, kv := range assignments {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return fmt.Errorf("malformed assignment %q (want key=value)", kv)
		}
		switch k {
		case "flake":
			if v == "" {
				return fmt.Errorf("flake must be non-empty")
			}
			pkg.Flake = v
		case "scope":
			s := manifest.Scope(v)
			if !s.Valid() {
				return fmt.Errorf("invalid scope %q (want system or home)", v)
			}
			pkg.Scope = s
		case "enabled":
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("enabled must be true or false, got %q", v)
			}
			pkg.Enabled = b
		case "pin":
			pkg.Pin = v
		case "user":
			if v != "" {
				if err := requireValidIdent("user", v); err != nil {
					return err
				}
			}
			pkg.User = v
		default:
			return fmt.Errorf("unknown field %q (want flake, scope, enabled, pin, user)", k)
		}
	}

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}
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
	if err := r.Commit(fmt.Sprintf("glix set %s: %s (%s)", *host, name, strings.Join(assignments, " "))); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Printf("updated %s on host %s: %s\n", name, *host, strings.Join(assignments, " "))
	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	return nil
}
