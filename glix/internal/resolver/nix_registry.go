package resolver

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// NixCLI looks up names via `nix registry list`. Each call reshells the
// command; output is small so this is fine for the resolver hot path.
type NixCLI struct {
	// Bin overrides the binary (default "nix").
	Bin string
}

// Lookup returns the canonical flake URI for `flake:<name>` if the local
// nix registry knows about it.
func (n NixCLI) Lookup(name string) (string, error) {
	bin := n.Bin
	if bin == "" {
		bin = "nix"
	}
	cmd := exec.Command(bin,
		"--extra-experimental-features", "nix-command flakes",
		"registry", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("nix registry list: %w", err)
	}
	return parseNixRegistry(out.String(), name)
}

// parseNixRegistry scans the stdout of `nix registry list` for a line whose
// "from" reference is flake:<name> and returns the "to" reference.
//
// Output format example:
//
//	user   flake:nixpkgs            github:NixOS/nixpkgs/nixos-unstable
//	global flake:home-manager       github:nix-community/home-manager
//
// The first column is the scope; we ignore it and match the second column.
func parseNixRegistry(text, name string) (string, error) {
	target := "flake:" + name
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if fields[1] == target {
			return fields[2], nil
		}
	}
	return "", ErrNotFound
}
