// Package resolver maps user-supplied references to canonical flake URIs.
//
// Resolution order (ADR-004):
//  1. Anything that looks like a flake URI is used verbatim.
//  2. Glixos registry (cached JSON).
//  3. `nix registry list` output.
//  4. Error with suggestions.
package resolver

import (
	"fmt"
	"strings"

	"github.com/glixos/glix/internal/registry"
)

// Source describes where a Result came from.
type Source string

const (
	SourceURI            Source = "uri"
	SourceGlixosRegistry Source = "glixos-registry"
	SourceNixRegistry    Source = "nix-registry"
)

// Result is what Resolve returns on success.
type Result struct {
	// Input is the original string the caller passed.
	Input string
	// Ref is the canonical flake URI (suitable for use in a flake input).
	Ref string
	// Source records which path of the chain matched.
	Source Source
	// Description is optional human-readable text (only set for registry hits).
	Description string
	// ShortName is a recommended package name. For registry hits this is
	// the registry key itself; for URI passthrough it's the caller's choice.
	ShortName string
}

// Options configures a single resolution.
type Options struct {
	// Registry, if non-nil, is consulted in step 2. Pass an empty registry
	// to skip the glixos registry without disabling the nix registry step.
	Registry *registry.Registry
	// NixRegistry, if non-nil, is consulted in step 3. Pass NoNixRegistry{}
	// to skip the step entirely.
	NixRegistry NixRegistryLooker
}

// NixRegistryLooker can resolve a flake:<name> alias via `nix registry list`.
// The default implementation is NixCLI{}.
type NixRegistryLooker interface {
	Lookup(name string) (string, error)
}

// NoNixRegistry disables the nix-registry fallback.
type NoNixRegistry struct{}

// Lookup always returns ErrNotFound.
func (NoNixRegistry) Lookup(string) (string, error) { return "", ErrNotFound }

// ErrNotFound is returned when no resolver path matches.
var ErrNotFound = fmt.Errorf("not found")

// Resolve runs the full resolution chain.
func Resolve(input string, opts Options) (*Result, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty reference")
	}

	if IsFlakeURI(input) {
		return &Result{Input: input, Ref: input, Source: SourceURI}, nil
	}

	if opts.Registry != nil {
		if entry, ok := opts.Registry.Packages[input]; ok {
			return &Result{
				Input:       input,
				Ref:         entry.Flake,
				Source:      SourceGlixosRegistry,
				Description: entry.Description,
				ShortName:   input,
			}, nil
		}
	}

	if opts.NixRegistry != nil {
		if ref, err := opts.NixRegistry.Lookup(input); err == nil {
			return &Result{
				Input:     input,
				Ref:       ref,
				Source:    SourceNixRegistry,
				ShortName: input,
			}, nil
		}
	}

	// Build a helpful error.
	var suggestions []string
	if opts.Registry != nil {
		suggestions = opts.Registry.Suggest(input, 5)
	}
	msg := fmt.Sprintf("could not resolve %q (not a flake URI, not in glixos registry, not in nix registry)", input)
	if len(suggestions) > 0 {
		msg += "\n\nDid you mean:\n  " + strings.Join(suggestions, "\n  ")
	}
	return nil, fmt.Errorf("%s", msg)
}

// IsFlakeURI reports whether s parses as a flake reference rather than a
// short name. Conservative: returns true for any of the known schemes and
// for bare paths starting with `.` or `/`.
func IsFlakeURI(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '/' || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	schemes := []string{
		"github:", "gitlab:", "sourcehut:", "srht:",
		"git+", "tarball+",
		"file:", "path:",
		"http:", "https:",
		"flake:",
	}
	for _, p := range schemes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
