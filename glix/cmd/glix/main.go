// Command glix is the CLI front end of glixos. See `glix help`.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version and commit are stamped at build time via -ldflags '-X main.version=… -X main.commit=…'.
// Defaults below are what you get from `go build` directly (no Nix wrapper).
var (
	version = "0.1.0-m7"
	commit  = "unknown"
)

type command struct {
	name    string
	summary string
	run     func(args []string) error
}

var commands = []command{
	{"init", "Bootstrap a user-packages flake repo for a host", cmdInit},
	{"add", "Add a package flake to the manifest", cmdAdd},
	{"remove", "Remove a package from the manifest", cmdRemove},
	{"enable", "Re-enable a previously disabled package", cmdEnable},
	{"disable", "Disable a package without removing it", cmdDisable},
	{"set", "Mutate fields on an existing package", cmdSet},
	{"show", "Show details for a single package", cmdShow},
	{"list", "List packages in the manifest", cmdList},
	{"search", "Search the glixos registry", cmdSearch},
	{"update", "Update flake inputs (all or selected)", cmdUpdate},
	{"rebuild", "Run nixos-rebuild against the user-packages flake", cmdRebuild},
	{"rollback", "Revert the last manifest commit and relock", cmdRollback},
	{"gc", "Run nix-collect-garbage (optionally with -d)", cmdGC},
	{"info", "Show repo summary, hosts, and resolved flake inputs", cmdInfo},
	{"doctor", "Run environment and repo health checks", cmdDoctor},
	{"version", "Print the glix version", cmdVersion},
	{"help", "Show this help", nil},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	name := os.Args[1]
	args := os.Args[2:]
	if name == "help" || name == "-h" || name == "--help" {
		usage()
		return
	}
	if name == "--version" || name == "-v" {
		_ = cmdVersion(nil)
		return
	}
	for _, c := range commands {
		if c.name == name {
			if c.run == nil {
				usage()
				return
			}
			if err := c.run(args); err != nil {
				fmt.Fprintf(os.Stderr, "glix %s: %v\n", c.name, err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "glix: unknown command %q\n\n", name)
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintf(os.Stderr, "glix %s — modular flake-driven NixOS\n\nUsage:\n  glix <command> [flags] [args]\n\nCommands:\n", version)
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-9s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(os.Stderr, "\nRun `glix <command> -h` for command-specific flags.")
}

func cmdVersion(_ []string) error {
	fmt.Printf("glix %s (%s)\n", version, commit)
	return nil
}

// newFlagSet returns a FlagSet that prints usage to stderr.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet("glix "+name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}
