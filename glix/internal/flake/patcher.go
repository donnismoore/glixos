// Package flake patches anchored regions inside flake.nix.
//
// glix never parses arbitrary Nix. It only mutates regions delimited by
// matching markers it emitted itself:
//
//	# >>> glix-managed <REGION> >>>
//	...
//	# <<< glix-managed <REGION> <<<
//
// After every mutating command, the entire region is regenerated from the
// in-memory manifest so on-disk state stays a pure projection of glix.toml.
package flake

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glixos/glix/internal/manifest"
)

// Region names known to glix.
const (
	RegionInputs = "inputs"
	RegionHosts  = "hosts"
)

// markers returns the start and end anchor lines for a region.
func markers(region string) (start, end string) {
	return fmt.Sprintf("# >>> glix-managed %s >>>", region),
		fmt.Sprintf("# <<< glix-managed %s <<<", region)
}

// ReplaceRegion replaces the content between start/end markers with body.
// Indentation of the start marker line is preserved on every emitted line.
// Returns an error if either marker is missing or appears more than once.
func ReplaceRegion(src, region, body string) (string, error) {
	start, end := markers(region)
	si := strings.Index(src, start)
	if si < 0 {
		return "", fmt.Errorf("region %q: start marker not found", region)
	}
	if strings.Index(src[si+len(start):], start) >= 0 {
		return "", fmt.Errorf("region %q: start marker appears multiple times", region)
	}
	ei := strings.Index(src, end)
	if ei < 0 {
		return "", fmt.Errorf("region %q: end marker not found", region)
	}
	if strings.Index(src[ei+len(end):], end) >= 0 {
		return "", fmt.Errorf("region %q: end marker appears multiple times", region)
	}
	if ei < si {
		return "", fmt.Errorf("region %q: end marker precedes start marker", region)
	}

	// Indentation = leading whitespace on the start-marker line.
	lineStart := strings.LastIndex(src[:si], "\n") + 1
	indent := src[lineStart:si]

	// End of the start-marker line, beginning of the end-marker line.
	afterStartLine := si + len(start)
	if i := strings.Index(src[afterStartLine:], "\n"); i >= 0 {
		afterStartLine += i + 1
	}
	endLineStart := strings.LastIndex(src[:ei], "\n") + 1

	var out bytes.Buffer
	out.WriteString(src[:afterStartLine])
	if body != "" {
		for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
			if line == "" {
				out.WriteByte('\n')
				continue
			}
			out.WriteString(indent)
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	out.WriteString(src[endLineStart:])
	return out.String(), nil
}

// RenderInputs produces the body of the glix-managed inputs region, derived
// from the manifest. Inputs are sorted alphabetically by package name.
func RenderInputs(m *manifest.Manifest) string {
	if m == nil || len(m.Packages) == 0 {
		return ""
	}
	names := make([]string, 0, len(m.Packages))
	for k := range m.Packages {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		p := m.Packages[name]
		if i > 0 {
			b.WriteByte('\n')
		}
		// Nix attribute names are constrained; we already sanitize package
		// names elsewhere, but quote defensively.
		fmt.Fprintf(&b, "%s = {\n", nixAttr(name))
		fmt.Fprintf(&b, "  url = %q;\n", p.Flake)
		b.WriteString("};\n")
	}
	return b.String()
}

// RenderHosts produces the body of the glix-managed hosts region. M3 only
// supports one host per glix init; future commands can extend this set.
func RenderHosts(hosts []HostEntry) string {
	if len(hosts) == 0 {
		return ""
	}
	sorted := append([]HostEntry(nil), hosts...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var b strings.Builder
	for _, h := range sorted {
		fmt.Fprintf(&b, "%s = mkHost %q %q;\n", nixAttr(h.Name), h.Name, h.System)
	}
	return b.String()
}

// HostEntry describes one nixosConfigurations row.
type HostEntry struct {
	Name   string
	System string
}

// nixAttr returns name as a valid Nix attribute. Bare if it matches
// [A-Za-z_][A-Za-z0-9_-]*, otherwise quoted.
func nixAttr(name string) string {
	if name == "" {
		return `""`
	}
	bare := true
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r == '_':
		case (r >= '0' && r <= '9') || r == '-':
			if i == 0 {
				bare = false
			}
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

// PatchFile atomically rewrites the named flake.nix with the given regions.
func PatchFile(path string, regions map[string]string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(b)
	for region, body := range regions {
		next, err := ReplaceRegion(src, region, body)
		if err != nil {
			return err
		}
		src = next
	}
	return atomicWrite(path, []byte(src), 0o644)
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
