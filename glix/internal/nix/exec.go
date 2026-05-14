// Package nix wraps the subset of `nix` / `nixos-rebuild` invocations that
// glix delegates to. Thin layer: no parsing of nix output.
package nix

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// FlakeUpdate runs `nix flake update [inputs...]` in dir. With no inputs,
// updates every input in flake.lock.
func FlakeUpdate(dir string, inputs ...string) error {
	args := []string{
		"--extra-experimental-features", "nix-command flakes",
		"flake", "update",
	}
	args = append(args, inputs...)
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

// Version returns the trimmed first line of `nix --version`.
// Returns an error if nix is not installed or not on PATH.
func Version() (string, error) {
	cmd := exec.Command("nix", "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("nix --version: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(strings.SplitN(stdout.String(), "\n", 2)[0]), nil
}

// HasFlakes reports whether `nix flake --help` succeeds, i.e. the flakes
// experimental feature is available to this user.
func HasFlakes() bool {
	cmd := exec.Command("nix",
		"--extra-experimental-features", "nix-command flakes",
		"flake", "--help",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
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
