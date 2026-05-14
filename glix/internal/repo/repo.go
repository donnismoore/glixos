// Package repo manages the user-packages git repo on disk: bootstrap,
// path discovery, and per-mutation auto-commit (ADR-008).
package repo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Repo is a handle to a user-packages flake repo.
type Repo struct {
	Root string
}

// DefaultRoot returns the default location for the user-packages repo:
// $XDG_CONFIG_HOME/glixos (falling back to ~/.config/glixos).
func DefaultRoot() (string, error) {
	if v := os.Getenv("GLIX_ROOT"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "glixos"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "glixos"), nil
}

// Discover walks up from start looking for a directory that contains both a
// flake.nix and a hosts/ subdirectory. If none is found, falls back to
// DefaultRoot if that directory exists.
func Discover(start string) (*Repo, error) {
	cur := start
	for {
		if isRepoRoot(cur) {
			return &Repo{Root: cur}, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	def, err := DefaultRoot()
	if err != nil {
		return nil, err
	}
	if isRepoRoot(def) {
		return &Repo{Root: def}, nil
	}
	return nil, fmt.Errorf("no glixos user-packages repo found (looked upward from %s and at %s)", start, def)
}

func isRepoRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "flake.nix")); err != nil {
		return false
	}
	if st, err := os.Stat(filepath.Join(dir, "hosts")); err != nil || !st.IsDir() {
		return false
	}
	return true
}

// FlakePath returns the absolute path to flake.nix.
func (r *Repo) FlakePath() string { return filepath.Join(r.Root, "flake.nix") }

// HostDir returns the absolute path to hosts/<host>/.
func (r *Repo) HostDir(host string) string { return filepath.Join(r.Root, "hosts", host) }

// ManifestPath returns the absolute path to hosts/<host>/glix.toml.
func (r *Repo) ManifestPath(host string) string {
	return filepath.Join(r.HostDir(host), "glix.toml")
}

// HostExists reports whether a host directory is present.
func (r *Repo) HostExists(host string) bool {
	st, err := os.Stat(r.HostDir(host))
	return err == nil && st.IsDir()
}

// ListHosts returns the names of all host subdirectories under hosts/.
func (r *Repo) ListHosts() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(r.Root, "hosts"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// GitInit runs `git init -b main` in the repo root if no .git directory
// already exists.
func (r *Repo) GitInit() error {
	if _, err := os.Stat(filepath.Join(r.Root, ".git")); err == nil {
		return nil
	}
	return r.git("init", "-b", "main")
}

// Commit stages all changes in the repo and creates a commit. If there are
// no changes, returns nil.
func (r *Repo) Commit(message string) error {
	if err := r.git("add", "-A"); err != nil {
		return err
	}
	// Check whether anything is staged.
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = r.Root
	if err := cmd.Run(); err == nil {
		// exit 0 => no staged changes.
		return nil
	}
	return r.git("commit", "-m", message)
}

func (r *Repo) git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
