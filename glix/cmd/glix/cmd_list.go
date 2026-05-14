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
	allHosts := fs.Bool("all-hosts", false, "show packages across every host")
	if err := fs.Parse(args); err != nil {
		return err
	}

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}

	var hosts []string
	if *allHosts {
		hosts, err = r.ListHosts()
		if err != nil {
			return err
		}
	} else {
		if !r.HostExists(*host) {
			return fmt.Errorf("host %q not found in %s", *host, r.Root)
		}
		hosts = []string{*host}
	}

	type row struct {
		host, name, scope, state, user, flake string
	}
	var rows []row
	for _, h := range hosts {
		m, err := manifest.Load(r.ManifestPath(h))
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
		for _, n := range names {
			p := m.Packages[n]
			state := "enabled"
			if !p.Enabled {
				state = "disabled"
			}
			user := p.User
			if user == "" && p.Scope == manifest.ScopeHome {
				user = m.Settings.PrimaryUser
			}
			rows = append(rows, row{h, n, string(p.Scope), state, user, p.Flake})
		}
	}

	if len(rows) == 0 {
		fmt.Println("(no packages)")
		return nil
	}

	if *allHosts {
		header := []string{"HOST", "NAME", "SCOPE", "STATE", "USER", "FLAKE"}
		body := make([][]string, len(rows))
		for i, r := range rows {
			body[i] = []string{r.host, r.name, r.scope, r.state, r.user, r.flake}
		}
		ui.Table(os.Stdout, header, body)
	} else {
		header := []string{"NAME", "SCOPE", "STATE", "USER", "FLAKE"}
		body := make([][]string, len(rows))
		for i, r := range rows {
			body[i] = []string{r.name, r.scope, r.state, r.user, r.flake}
		}
		ui.Table(os.Stdout, header, body)
	}
	return nil
}
