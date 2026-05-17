package flake

import (
	"path/filepath"
	"regexp"
	"strings"
)

// LocalPath extracts the filesystem path embedded in a flake reference, if
// the ref uses a local-path form (path:, file:, file://, or a bare path
// starting with /, ./, or ../). Any query string or fragment is stripped.
// Returns "", false for non-local refs (github:, https:, git+, …).
func LocalPath(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	rest := ref
	if i := strings.IndexAny(rest, "?#"); i >= 0 {
		rest = rest[:i]
	}
	switch {
	case strings.HasPrefix(rest, "path://"):
		return strings.TrimPrefix(rest, "path://"), true
	case strings.HasPrefix(rest, "path:"):
		return strings.TrimPrefix(rest, "path:"), true
	case strings.HasPrefix(rest, "file://"):
		return strings.TrimPrefix(rest, "file://"), true
	case strings.HasPrefix(rest, "file:"):
		return strings.TrimPrefix(rest, "file:"), true
	case strings.HasPrefix(rest, "/"),
		strings.HasPrefix(rest, "./"),
		strings.HasPrefix(rest, "../"):
		return rest, true
	}
	return "", false
}

// IsAbsoluteLocalRef reports whether ref is a local-path flake reference
// rooted at an absolute filesystem path (path:/foo, file:///foo, /foo).
// Such refs are non-portable: committing one bakes the maintainer's
// machine layout into the flake, breaking every downstream clone.
func IsAbsoluteLocalRef(ref string) bool {
	p, ok := LocalPath(ref)
	if !ok {
		return false
	}
	return filepath.IsAbs(p)
}

// EscapesRoot reports whether the local-path flake ref resolves to a
// location outside root. flakeDir is the directory containing the flake
// that declares the input; relative path: refs are resolved against it.
// Non-local refs return (false, false).
func EscapesRoot(ref, root, flakeDir string) (escapes, isLocal bool) {
	p, ok := LocalPath(ref)
	if !ok {
		return false, false
	}
	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(flakeDir, abs)
	}
	abs = filepath.Clean(abs)
	rootClean := filepath.Clean(root)
	rel, err := filepath.Rel(rootClean, abs)
	if err != nil {
		return true, true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true, true
	}
	return false, true
}

var urlAttrPattern = regexp.MustCompile(`url\s*=\s*"([^"]+)"`)

// ScanInputURLs returns every flake input URL declared in src, identified
// by lines of the form `url = "…";`. Best-effort: glix never fully parses
// Nix, so this misses URLs constructed by interpolation or held in let
// bindings — which is fine for the doctor's portability check.
func ScanInputURLs(src string) []string {
	matches := urlAttrPattern.FindAllStringSubmatch(src, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}
