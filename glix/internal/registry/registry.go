// Package registry holds the glixos package registry data model and loader.
//
// The registry is a small JSON document mapping short package names to
// canonical flake URIs. It is fetched from a configurable URL (HTTP or
// file:) and cached to disk under $XDG_CACHE_HOME/glix/.
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const CurrentSchema = 1

// Registry is the in-memory form of registry.json.
type Registry struct {
	Schema   int              `json:"schema"`
	Packages map[string]Entry `json:"packages"`
}

// Entry describes one short-name → flake mapping.
type Entry struct {
	Flake       string `json:"flake"`
	Description string `json:"description,omitempty"`
}

// Empty returns a syntactically valid empty registry.
func Empty() *Registry {
	return &Registry{Schema: CurrentSchema, Packages: map[string]Entry{}}
}

// Parse validates and returns a Registry from the given JSON bytes.
func Parse(data []byte) (*Registry, error) {
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("registry: parse json: %w", err)
	}
	if r.Schema != CurrentSchema {
		return nil, fmt.Errorf("registry: unsupported schema %d (expected %d)", r.Schema, CurrentSchema)
	}
	if r.Packages == nil {
		r.Packages = map[string]Entry{}
	}
	return &r, nil
}

// MaxRegistrySize caps a fetched registry response (http or file://).
// Hostile or compromised servers can otherwise return arbitrarily large
// bodies and OOM the glix process.
const MaxRegistrySize = 10 << 20 // 10 MiB

// Loader fetches and caches the registry. Concurrent use is not supported.
type Loader struct {
	// URL is either http(s):// or file://; an empty string disables loading.
	URL string
	// CachePath is the on-disk cache file. Created on first successful fetch.
	CachePath string
	// TTL is how long the cache is considered fresh.
	TTL time.Duration
	// Refresh forces a refetch even if the cache is fresh.
	Refresh bool
	// HTTPTimeout caps each network attempt.
	HTTPTimeout time.Duration
	// AllowFileURLs gates the file:// scheme. Defaults to false because a
	// file:// URL chosen by a less-trusted actor in a shared-tenancy or
	// privileged context turns the registry fetcher into an arbitrary
	// local-file read primitive (a confused-deputy hazard). Callers that
	// genuinely want to load a local registry must opt in.
	AllowFileURLs bool
	// Warn is invoked on non-fatal conditions (offline, stale cache, etc.).
	// May be nil.
	Warn func(string)
}

// Load returns a Registry, honoring TTL/refresh logic and offline fallback.
//
// Resolution order:
//  1. If URL is empty: return Empty(), nil.
//  2. If cache exists and is fresh (mtime within TTL) and !Refresh:
//     load and return cache.
//  3. Try to fetch from URL. On success, write cache and return.
//  4. On fetch failure: if cache exists, load it and warn; else return error.
func (l *Loader) Load() (*Registry, error) {
	if strings.TrimSpace(l.URL) == "" {
		return Empty(), nil
	}
	timeout := l.HTTPTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ttl := l.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	cacheFresh := false
	if !l.Refresh && l.CachePath != "" {
		if st, err := os.Stat(l.CachePath); err == nil {
			if time.Since(st.ModTime()) < ttl {
				cacheFresh = true
			}
		}
	}
	if cacheFresh {
		if r, err := l.loadCache(); err == nil {
			return r, nil
		}
		// Cache unreadable — fall through to refetch.
	}

	data, fetchErr := fetch(l.URL, timeout, l.AllowFileURLs)
	if fetchErr == nil {
		r, parseErr := Parse(data)
		if parseErr != nil {
			return nil, parseErr
		}
		if l.CachePath != "" {
			if err := writeCache(l.CachePath, data); err != nil && l.Warn != nil {
				l.Warn(fmt.Sprintf("could not write registry cache: %v", err))
			}
		}
		return r, nil
	}

	// Fetch failed — try the cache as a last resort.
	if l.CachePath != "" {
		if r, err := l.loadCache(); err == nil {
			if l.Warn != nil {
				l.Warn(fmt.Sprintf("offline: using cached registry (%v)", fetchErr))
			}
			return r, nil
		}
	}
	return nil, fmt.Errorf("registry: fetch %s: %w", l.URL, fetchErr)
}

func (l *Loader) loadCache() (*Registry, error) {
	b, err := os.ReadFile(l.CachePath)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

func fetch(rawURL string, timeout time.Duration, allowFileURLs bool) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "file":
		if !allowFileURLs {
			return nil, fmt.Errorf("file:// registry URLs require --allow-file-registry")
		}
		path := u.Path
		if path == "" {
			path = u.Opaque
		}
		return readCapped(path)
	case "http", "https":
		client := &http.Client{Timeout: timeout}
		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "glix/0.1.0-m4")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, MaxRegistrySize+1))
		if err != nil {
			return nil, err
		}
		if len(data) > MaxRegistrySize {
			return nil, fmt.Errorf("registry response exceeds %d bytes", MaxRegistrySize)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported registry URL scheme %q", u.Scheme)
	}
}

// readCapped reads a local file but caps the read at MaxRegistrySize so
// a hostile or runaway local file cannot OOM the process.
func readCapped(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, MaxRegistrySize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxRegistrySize {
		return nil, fmt.Errorf("registry response exceeds %d bytes", MaxRegistrySize)
	}
	return data, nil
}

func writeCache(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".glix-reg-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// DefaultCachePath returns $XDG_CACHE_HOME/glix/registry.json (or
// ~/.cache/glix/registry.json as a fallback).
func DefaultCachePath() (string, error) {
	if v := os.Getenv("XDG_CACHE_HOME"); v != "" {
		return filepath.Join(v, "glix", "registry.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "glix", "registry.json"), nil
}

// Search returns entries whose name or description contains query
// (case-insensitive). Results are sorted by name.
func (r *Registry) Search(query string) []Match {
	q := strings.ToLower(strings.TrimSpace(query))
	out := []Match{}
	for name, e := range r.Packages {
		if q == "" ||
			strings.Contains(strings.ToLower(name), q) ||
			strings.Contains(strings.ToLower(e.Description), q) {
			out = append(out, Match{Name: name, Entry: e})
		}
	}
	// Sort: exact-prefix matches first, then alphabetical.
	sortMatches(out, q)
	return out
}

// Match is one Search result.
type Match struct {
	Name  string
	Entry Entry
}

// Suggest returns up to n registry names that loosely resemble query
// (substring containment). Used in resolver error messages.
func (r *Registry) Suggest(query string, n int) []string {
	if query == "" {
		return nil
	}
	q := strings.ToLower(query)
	var hits []string
	for name := range r.Packages {
		if strings.Contains(strings.ToLower(name), q) {
			hits = append(hits, name)
		}
	}
	sortStringsAlpha(hits)
	if len(hits) > n {
		hits = hits[:n]
	}
	return hits
}

// IgnoreEntry is a sentinel used for tests; not exported intentionally.
var _ = errors.New
