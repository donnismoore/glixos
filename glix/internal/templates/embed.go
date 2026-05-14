// Package templates embeds the file templates that `glix init` drops into a
// freshly created user-packages flake repo.
package templates

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed flake.nix.tmpl
var flakeNix string

//go:embed host.nix.tmpl
var hostNix string

//go:embed glix.toml.tmpl
var glixToml string

// FlakeNix returns the rendered top-level flake.nix.
// Placeholders: @CORE_URL@.
func FlakeNix(coreURL string) string {
	return strings.ReplaceAll(flakeNix, "@CORE_URL@", coreURL)
}

// HostNix returns the rendered hosts/<host>/default.nix.
// Placeholders: @HOSTNAME@, @USER@.
func HostNix(hostname, user string) string {
	s := strings.ReplaceAll(hostNix, "@HOSTNAME@", hostname)
	s = strings.ReplaceAll(s, "@USER@", user)
	return s
}

// GlixToml returns the rendered hosts/<host>/glix.toml seed.
func GlixToml() string {
	return glixToml
}

// MustRender is a small helper for callers that prefer panicking on missing
// placeholders. Not currently exercised by the CLI but kept for future use.
func MustRender(name, content string, vars map[string]string) string {
	out := content
	for k, v := range vars {
		marker := fmt.Sprintf("@%s@", k)
		if !strings.Contains(out, marker) {
			panic(fmt.Sprintf("template %s: missing placeholder %s", name, marker))
		}
		out = strings.ReplaceAll(out, marker, v)
	}
	return out
}
