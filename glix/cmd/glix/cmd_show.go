package main

import (
	"fmt"
	"sort"

	"github.com/glixos/glix/internal/manifest"
)

func cmdShow(args []string) error {
	fs := newFlagSet("show")
	host := fs.String("host", currentHostname(), "host whose manifest to read")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: glix show <name>")
	}
	name := fs.Arg(0)

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
	p, ok := m.Packages[name]
	if !ok {
		return fmt.Errorf("package %q not in manifest", name)
	}

	state := "enabled"
	if !p.Enabled {
		state = "disabled"
	}
	fmt.Printf("name    %s\n", name)
	fmt.Printf("flake   %s\n", p.Flake)
	fmt.Printf("scope   %s\n", p.Scope)
	fmt.Printf("state   %s\n", state)
	if p.Pin != "" {
		fmt.Printf("pin     %s\n", p.Pin)
	}
	if p.Scope == manifest.ScopeHome {
		user := p.User
		if user == "" {
			user = m.Settings.PrimaryUser
			if user == "" {
				user = "(default)"
			}
		}
		fmt.Printf("user    %s\n", user)
	}
	fmt.Printf("host    %s\n", *host)
	if len(p.Config) > 0 {
		fmt.Println("config")
		keys := make([]string, 0, len(p.Config))
		for k := range p.Config {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s = %s\n", k, p.Config[k])
		}
	}
	return nil
}
