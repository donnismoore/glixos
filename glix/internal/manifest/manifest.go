// Package manifest reads and writes glix.toml, the single file the glix CLI
// owns. Read is permissive (uses BurntSushi/toml); Write is hand-rolled so
// the on-disk format is deterministic regardless of map iteration order.
//
// Schema version 1.
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const CurrentSchema = 1

// Scope is the scope at which a package is installed.
type Scope string

const (
	ScopeSystem Scope = "system"
	ScopeHome   Scope = "home"
)

// Valid reports whether s is a known scope.
func (s Scope) Valid() bool {
	return s == ScopeSystem || s == ScopeHome
}

// Manifest mirrors the on-disk glix.toml schema.
type Manifest struct {
	Schema   int                `toml:"schema"`
	Settings Settings           `toml:"settings"`
	Packages map[string]Package `toml:"packages"`
}

// Settings holds host-wide manifest settings.
type Settings struct {
	DefaultScope Scope  `toml:"default_scope"`
	AutoApply    bool   `toml:"auto_apply"`
	RegistryURL  string `toml:"registry_url,omitempty"`
}

// Package describes one entry under [packages.<name>].
type Package struct {
	Flake   string `toml:"flake"`
	Scope   Scope  `toml:"scope"`
	Enabled bool   `toml:"enabled"`
	Pin     string `toml:"pin,omitempty"`
	// Config is intentionally omitted in M3. It will be added once
	// `glix set` lands and we settle the serialization shape.
}

// New returns a manifest with default settings and no packages.
func New() *Manifest {
	return &Manifest{
		Schema: CurrentSchema,
		Settings: Settings{
			DefaultScope: ScopeSystem,
			AutoApply:    false,
		},
		Packages: map[string]Package{},
	}
}

// Load reads and validates a glix.toml file.
func Load(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	m := &Manifest{Packages: map[string]Package{}}
	if _, err := toml.Decode(string(b), m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Schema != CurrentSchema {
		return nil, fmt.Errorf("%s: unsupported schema version %d (expected %d)", path, m.Schema, CurrentSchema)
	}
	if m.Settings.DefaultScope == "" {
		m.Settings.DefaultScope = ScopeSystem
	}
	if !m.Settings.DefaultScope.Valid() {
		return nil, fmt.Errorf("%s: invalid default_scope %q", path, m.Settings.DefaultScope)
	}
	for name, p := range m.Packages {
		if !p.Scope.Valid() {
			return nil, fmt.Errorf("%s: package %q has invalid scope %q", path, name, p.Scope)
		}
	}
	return m, nil
}

// Save atomically writes the manifest using a deterministic format.
func Save(path string, m *Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := Encode(&buf, m); err != nil {
		return err
	}
	return atomicWrite(path, buf.Bytes(), 0o644)
}

// Encode writes the manifest in the canonical glix format. Output is stable:
// packages are emitted in alphabetical order, fields in a fixed order.
func Encode(w io.Writer, m *Manifest) error {
	if m == nil {
		return errors.New("nil manifest")
	}
	fmt.Fprintln(w, "# glixos manifest. Managed by glix; hand edits preserved on best-effort.")
	fmt.Fprintf(w, "schema = %d\n\n", m.Schema)

	fmt.Fprintln(w, "[settings]")
	fmt.Fprintf(w, "default_scope = %q\n", string(m.Settings.DefaultScope))
	fmt.Fprintf(w, "auto_apply    = %t\n", m.Settings.AutoApply)
	if m.Settings.RegistryURL != "" {
		fmt.Fprintf(w, "registry_url  = %q\n", m.Settings.RegistryURL)
	}
	fmt.Fprintln(w)

	names := make([]string, 0, len(m.Packages))
	for k := range m.Packages {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		p := m.Packages[name]
		fmt.Fprintf(w, "[packages.%s]\n", tomlBareOrQuoted(name))
		fmt.Fprintf(w, "flake   = %q\n", p.Flake)
		fmt.Fprintf(w, "scope   = %q\n", string(p.Scope))
		fmt.Fprintf(w, "enabled = %t\n", p.Enabled)
		if p.Pin != "" {
			fmt.Fprintf(w, "pin     = %q\n", p.Pin)
		}
		fmt.Fprintln(w)
	}
	return nil
}

// tomlBareOrQuoted returns the TOML key form for name: bare if it matches
// [A-Za-z0-9_-]+, otherwise a quoted string.
func tomlBareOrQuoted(name string) string {
	if name == "" {
		return `""`
	}
	bare := true
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_', r == '-':
		default:
			bare = false
		}
		if !bare {
			break
		}
	}
	if bare {
		return name
	}
	return fmt.Sprintf("%q", name)
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".glix-tmp-*")
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
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// PackageNameFromRef derives a package name from a flake reference.
// Examples:
//
//	github:owner/repo            -> repo
//	github:owner/repo/branch     -> repo
//	github:owner/repo?dir=foo    -> foo
//	gitlab:owner/repo            -> repo
//	path:./examples/pkg-hello    -> pkg-hello
//	https://x/y/foo.git          -> foo
//
// Returns ("", false) if a sensible name cannot be derived.
func PackageNameFromRef(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}

	// Split scheme.
	var scheme, rest string
	if i := strings.Index(ref, ":"); i >= 0 {
		scheme = ref[:i]
		rest = ref[i+1:]
	} else {
		rest = ref
	}

	// Honor an explicit `dir=` query parameter first.
	if i := strings.Index(rest, "?"); i >= 0 {
		query := rest[i+1:]
		rest = rest[:i]
		for _, kv := range strings.Split(query, "&") {
			if strings.HasPrefix(kv, "dir=") {
				if v := strings.TrimPrefix(kv, "dir="); v != "" {
					return filepath.Base(v), true
				}
			}
		}
	}
	rest = strings.TrimSuffix(rest, ".git")

	// Tokenize, dropping empty / `.` / `..` segments.
	var parts []string
	for _, p := range strings.Split(rest, "/") {
		if p == "" || p == "." || p == ".." {
			continue
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return "", false
	}

	switch scheme {
	case "github", "gitlab", "sourcehut", "srht":
		// <scheme>:owner/repo[/ref]
		if len(parts) >= 2 {
			return parts[1], true
		}
		return parts[0], true
	default:
		// path:, https:, git+*, or no scheme — last component wins.
		return parts[len(parts)-1], true
	}
}
