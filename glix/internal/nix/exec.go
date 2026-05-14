// Package nix wraps the subset of `nix` / `nixos-rebuild` invocations that
// glix delegates to. Thin layer: no parsing of nix output.
package nix

import (
	"fmt"
	"os"
	"os/exec"
)

// FlakeLock runs `nix flake lock` in dir, optionally updating named inputs.
func FlakeLock(dir string, inputs ...string) error {
	args := []string{
		"--extra-experimental-features", "nix-command flakes",
		"flake", "lock",
	}
	for _, in := range inputs {
		args = append(args, "--update-input", in)
	}
	return run(dir, "nix", args...)
}

// Rebuild runs `nixos-rebuild <action> --flake .#<host>` in dir.
// action is one of: switch, test, boot, build.
func Rebuild(dir, host, action string) error {
	target := "."
	if host != "" {
		target = "." + "#" + host
	}
	return run(dir, "nixos-rebuild", action, "--flake", target)
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
