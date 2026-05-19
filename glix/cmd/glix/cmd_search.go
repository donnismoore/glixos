package main

import (
	"fmt"
	"os"

	"github.com/glixos/glix/internal/registry"
	"github.com/glixos/glix/internal/ui"
)

func cmdSearch(args []string) error {
	fs := newFlagSet("search")
	host := fs.String("host", currentHostname(), "host whose settings to use (for registry_url)")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	regURL := fs.String("registry-url", "", "override registry URL")
	refresh := fs.Bool("refresh", false, "force refetch of the registry")
	allowFileReg := fs.Bool("allow-file-registry", false, "permit file:// registry URLs (off by default to avoid arbitrary local-file reads in shared-tenancy contexts)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	query := ""
	if fs.NArg() > 0 {
		query = fs.Arg(0)
	}

	url, err := resolveRegistryURL(*dir, *host, *regURL)
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("no registry URL configured (set settings.registry_url in glix.toml, pass --registry-url, or set $GLIX_REGISTRY_URL)")
	}

	cache, err := registry.DefaultCachePath()
	if err != nil {
		return err
	}
	loader := &registry.Loader{
		URL:           url,
		CachePath:     cache,
		Refresh:       *refresh,
		AllowFileURLs: *allowFileReg,
		Warn:          func(s string) { fmt.Fprintln(os.Stderr, "glix search:", s) },
	}
	reg, err := loader.Load()
	if err != nil {
		return err
	}

	matches := reg.Search(query)
	if len(matches) == 0 {
		if query == "" {
			fmt.Println("(empty registry)")
		} else {
			fmt.Printf("no matches for %q\n", query)
		}
		return nil
	}
	rows := make([][]string, 0, len(matches))
	for _, m := range matches {
		rows = append(rows, []string{m.Name, m.Entry.Flake, m.Entry.Description})
	}
	ui.Table(os.Stdout, []string{"NAME", "FLAKE", "DESCRIPTION"}, rows)
	return nil
}
