package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/glixos/glix/internal/repo"
)

// resolveRepo returns the repo we should operate on, honoring --dir if set.
func resolveRepo(dir string) (*repo.Repo, error) {
	if dir != "" {
		return &repo.Repo{Root: dir}, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return repo.Discover(cwd)
}

// currentHostname returns the local machine hostname, lowercased.
func currentHostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "glixos"
	}
	return h
}

// currentUser returns the value of $USER, falling back to "glixos".
func currentUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return "glixos"
}

var validIdent = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// requireValidIdent ensures name is usable both as a TOML bare key and a Nix
// attribute name. Returns a descriptive error otherwise.
func requireValidIdent(kind, name string) error {
	if !validIdent.MatchString(name) {
		return fmt.Errorf("%s %q must match [A-Za-z][A-Za-z0-9_-]* ", kind, name)
	}
	return nil
}
