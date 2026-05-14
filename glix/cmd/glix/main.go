// Command glix is the CLI front end of glixos. See `glix help`.
package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0-m3"

type command struct {
	name    string
	summary string
	run     func(args []string) error
}

var commands = []command{
	{"init", "Bootstrap a user-packages flake repo for a host", cmdInit},
	{"add", "Add a package flake to the manifest", cmdAdd},
	{"remove", "Remove a package from the manifest", cmdRemove},
	{"list", "List packages in the manifest", cmdList},
	{"rebuild", "Run nixos-rebuild against the user-packages flake", cmdRebuild},
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
		fmt.Fprintf(os.Stderr, "  %-9s  %s\n", c.name, c.summary)
	}
	fmt.Fprintln(os.Stderr, "\nRun `glix <command> -h` for command-specific flags.")
}

func cmdVersion(_ []string) error {
	fmt.Println(version)
	return nil
}

// newFlagSet returns a FlagSet that prints usage to stderr.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet("glix "+name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}
