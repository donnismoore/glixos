package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/ui"
)

func cmdList(args []string) error {
	fs := newFlagSet("list")
	host := fs.String("host", currentHostname(), "host whose manifest to read")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	showAll := fs.Bool("all", false, "include disabled packages")
	if err := fs.Parse(args); err != nil {
		return err
	}

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

	names := make([]string, 0, len(m.Packages))
	for k, p := range m.Packages {
		if !p.Enabled && !*showAll {
			continue
		}
		names = append(names, k)
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Println("(no packages)")
		return nil
	}

	rows := make([][]string, 0, len(names))
	for _, n := range names {
		p := m.Packages[n]
		state := "enabled"
		if !p.Enabled {
			state = "disabled"
		}
		rows = append(rows, []string{n, string(p.Scope), state, p.Flake})
	}
	ui.Table(os.Stdout, []string{"NAME", "SCOPE", "STATE", "FLAKE"}, rows)
	return nil
}
