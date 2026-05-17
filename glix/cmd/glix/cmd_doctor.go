package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glixos/glix/internal/flake"
	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

// cmdDoctor runs a series of cheap, read-only checks against the local
// environment and the discovered user-packages repo. Each check prints
// "OK" / "WARN" / "FAIL" and a short message. Returns a non-zero exit
// status iff at least one FAIL is recorded.
func cmdDoctor(args []string) error {
	fs := newFlagSet("doctor")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var fails int

	check := func(label string, status, msg string) {
		fmt.Printf("  [%s] %s — %s\n", status, label, msg)
		if status == "FAIL" {
			fails++
		}
	}

	// 1. Nix toolchain.
	fmt.Println("toolchain:")
	if v, err := nix.Version(); err != nil {
		check("nix", "FAIL", err.Error())
	} else {
		check("nix", "OK", v)
	}
	if nix.HasFlakes() {
		check("flakes", "OK", "nix-command + flakes enabled")
	} else {
		check("flakes", "FAIL", "nix flake command unavailable")
	}

	// 2. Repo discovery.
	fmt.Println("repo:")
	r, err := resolveRepo(*dir)
	if err != nil {
		check("discovery", "FAIL", err.Error())
		if fails > 0 {
			fmt.Printf("\n%d check(s) failed\n", fails)
			return fmt.Errorf("doctor reported %d failure(s)", fails)
		}
		return nil
	}
	check("discovery", "OK", "root="+r.Root)

	if r.HasGit() {
		clean, err := r.IsClean()
		switch {
		case err != nil:
			check("git", "WARN", "could not query git: "+err.Error())
		case clean:
			check("git", "OK", "working tree clean")
		default:
			check("git", "WARN", "working tree has uncommitted changes")
		}
	} else {
		check("git", "WARN", "no .git directory (run `git init`)")
	}

	// 3. flake.nix and flake.lock.
	fmt.Println("flake:")
	if _, err := os.Stat(r.FlakePath()); err != nil {
		check("flake.nix", "FAIL", err.Error())
	} else {
		check("flake.nix", "OK", r.FlakePath())
		if b, err := os.ReadFile(r.FlakePath()); err == nil {
			var bad []string
			for _, u := range flake.ScanInputURLs(string(b)) {
				escapes, isLocal := flake.EscapesRoot(u, r.Root, r.Root)
				if !isLocal {
					continue
				}
				if escapes || flake.IsAbsoluteLocalRef(u) {
					bad = append(bad, u)
				}
			}
			switch len(bad) {
			case 0:
				check("inputs", "OK", "no non-portable local-path inputs")
			default:
				check("inputs", "WARN",
					"non-portable path input(s) — these won't resolve on other machines: "+strings.Join(bad, ", "))
			}
		}
	}
	lockPath := filepath.Join(r.Root, "flake.lock")
	if _, err := os.Stat(lockPath); err != nil {
		check("flake.lock", "WARN", "missing (run `glix add` or `nix flake lock`)")
	} else {
		check("flake.lock", "OK", lockPath)
	}

	// 4. Per-host manifests.
	hosts, err := r.ListHosts()
	if err != nil {
		check("hosts/", "FAIL", err.Error())
	} else if len(hosts) == 0 {
		check("hosts/", "WARN", "no hosts defined (run `glix init --host=<name>`)")
	} else {
		fmt.Println("hosts:")
		for _, h := range hosts {
			mp := r.ManifestPath(h)
			if _, err := os.Stat(mp); err != nil {
				check(h, "FAIL", "manifest missing: "+err.Error())
				continue
			}
			m, err := manifest.Load(mp)
			if err != nil {
				check(h, "FAIL", err.Error())
				continue
			}
			check(h, "OK", fmt.Sprintf("%d package(s)", len(m.Packages)))
		}
	}

	if fails > 0 {
		fmt.Printf("\n%d check(s) failed\n", fails)
		return fmt.Errorf("doctor reported %d failure(s)", fails)
	}
	fmt.Println("\nall checks passed")
	return nil
}
