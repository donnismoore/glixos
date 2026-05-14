package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glixos/glix/internal/flake"
	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/repo"
	"github.com/glixos/glix/internal/templates"
)

const defaultCoreURL = "github:glixos/glixos?dir=core"

func cmdInit(args []string) error {
	fs := newFlagSet("init")
	host := fs.String("host", currentHostname(), "host name (used for hosts/<host>/)")
	user := fs.String("user", currentUser(), "primary user account name")
	system := fs.String("system", "x86_64-linux", "Nix system tuple")
	dir := fs.String("dir", "", "target directory (default: $XDG_CONFIG_HOME/glixos)")
	coreURL := fs.String("core", defaultCoreURL, "flake URL for glixos-core")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireValidIdent("hostname", *host); err != nil {
		return err
	}
	if err := requireValidIdent("user", *user); err != nil {
		return err
	}

	root := *dir
	if root == "" {
		def, err := repo.DefaultRoot()
		if err != nil {
			return err
		}
		root = def
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	r := &repo.Repo{Root: root}

	// Refuse to clobber an existing repo with a different layout.
	if existing, err := os.ReadDir(root); err == nil && len(existing) > 0 {
		if _, err := os.Stat(r.FlakePath()); err == nil {
			if r.HostExists(*host) {
				return fmt.Errorf("host %q already exists at %s", *host, r.HostDir(*host))
			}
			// Existing repo, new host: allow it.
		} else {
			// Non-empty directory without a flake.nix is unsafe.
			return fmt.Errorf("target directory %s is not empty and is not a glixos repo", root)
		}
	}

	// Write flake.nix only if absent (existing-repo init re-uses the file).
	if _, err := os.Stat(r.FlakePath()); errors.Is(err, os.ErrNotExist) {
		if err := writeFile(r.FlakePath(), []byte(templates.FlakeNix(*coreURL)), 0o644); err != nil {
			return err
		}
	}

	// Host directory + per-host files.
	if err := os.MkdirAll(r.HostDir(*host), 0o755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(r.HostDir(*host), "default.nix"), []byte(templates.HostNix(*host, *user)), 0o644); err != nil {
		return err
	}
	if err := writeFile(r.ManifestPath(*host), []byte(templates.GlixToml()), 0o644); err != nil {
		return err
	}

	// Register the host in the glix-managed hosts region of flake.nix.
	if err := regenerateFlake(r); err != nil {
		return err
	}

	// .gitignore.
	gitignore := "flake.lock.bak\nresult\nresult-*\n.direnv/\n"
	gp := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(gp); errors.Is(err, os.ErrNotExist) {
		if err := writeFile(gp, []byte(gitignore), 0o644); err != nil {
			return err
		}
	}

	if err := r.GitInit(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	if err := r.Commit(fmt.Sprintf("glix init: bootstrap host %s", *host)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Printf("initialized glixos user-packages repo at %s\n", root)
	fmt.Printf("  host:   %s\n", *host)
	fmt.Printf("  user:   %s\n", *user)
	fmt.Printf("  system: %s\n", *system)
	fmt.Printf("  core:   %s\n", *coreURL)
	return nil
}

func writeFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		return err
	}
	return nil
}

// regenerateFlake rewrites the glix-managed regions of flake.nix from the
// current state of the repo (manifests across all hosts + host directory
// listing).
func regenerateFlake(r *repo.Repo) error {
	hosts, err := r.ListHosts()
	if err != nil {
		return err
	}
	entries := make([]flake.HostEntry, 0, len(hosts))
	for _, h := range hosts {
		entries = append(entries, flake.HostEntry{Name: h, System: "x86_64-linux"})
	}

	// Inputs are the union of all hosts' manifests. M3 only ever has one
	// host per init, but unioning here is forward-compatible.
	merged := manifest.New()
	for _, h := range hosts {
		mp := r.ManifestPath(h)
		if _, err := os.Stat(mp); err != nil {
			continue
		}
		m, err := manifest.Load(mp)
		if err != nil {
			return fmt.Errorf("load %s: %w", mp, err)
		}
		for name, pkg := range m.Packages {
			merged.Packages[name] = pkg
		}
	}

	return flake.PatchFile(r.FlakePath(), map[string]string{
		flake.RegionInputs: flake.RenderInputs(merged),
		flake.RegionHosts:  flake.RenderHosts(entries),
	})
}
