package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/registry"
)

// resolveRegistryURL determines which URL to read the registry from.
// Precedence: explicit flag > $GLIX_REGISTRY_URL > manifest settings.
func resolveRegistryURL(dir, host, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if v := os.Getenv("GLIX_REGISTRY_URL"); v != "" {
		return v, nil
	}
	r, err := resolveRepo(dir)
	if err != nil {
		// No repo discovered → no manifest settings to read.
		return "", nil
	}
	if !r.HostExists(host) {
		return "", nil
	}
	m, err := manifest.Load(r.ManifestPath(host))
	if err != nil {
		return "", err
	}
	return m.Settings.RegistryURL, nil
}

// loadRegistry constructs a Loader and runs it. Returns an empty registry
// (with no error) when URL is empty, so callers can transparently skip the
// glixos-registry step in resolution.
func loadRegistry(url string, refresh bool) (*registry.Registry, error) {
	if url == "" {
		return registry.Empty(), nil
	}
	cache, err := registry.DefaultCachePath()
	if err != nil {
		return nil, err
	}
	loader := &registry.Loader{
		URL:       url,
		CachePath: cache,
		Refresh:   refresh,
		Warn:      func(s string) { fmt.Fprintln(os.Stderr, "glix:", s) },
	}
	reg, err := loader.Load()
	if err != nil {
		// Soft-fail: a missing/unreachable registry should not block adding by URI.
		if errors.Is(err, os.ErrNotExist) {
			return registry.Empty(), nil
		}
		fmt.Fprintln(os.Stderr, "glix: registry load failed:", err)
		return registry.Empty(), nil
	}
	return reg, nil
}
