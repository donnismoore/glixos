package resolver

import (
	"strings"
	"testing"

	"github.com/glixos/glix/internal/registry"
)

func TestIsFlakeURI(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"github:o/r", true},
		{"gitlab:o/r", true},
		{"github:o/r/branch", true},
		{"github:o/r?dir=sub", true},
		{"path:./x", true},
		{"./x", true},
		{"../x", true},
		{"/abs/path", true},
		{"https://example.com/foo.tar.gz", true},
		{"file:///x/y", true},
		{"git+ssh://x", true},
		{"flake:nixpkgs", true},
		{"firefox", false},
		{"helix", false},
		{"my-package", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsFlakeURI(c.in); got != c.want {
			t.Errorf("IsFlakeURI(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestResolve_URI(t *testing.T) {
	r, err := Resolve("github:owner/repo", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != SourceURI || r.Ref != "github:owner/repo" {
		t.Fatalf("got %+v", r)
	}
}

func TestResolve_GlixosRegistry(t *testing.T) {
	reg, _ := registry.Parse([]byte(`{
		"schema": 1,
		"packages": {"helix": {"flake": "github:helix-editor/helix", "description": "editor"}}
	}`))
	r, err := Resolve("helix", Options{Registry: reg})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != SourceGlixosRegistry {
		t.Fatalf("source = %s", r.Source)
	}
	if r.Ref != "github:helix-editor/helix" || r.ShortName != "helix" || r.Description != "editor" {
		t.Fatalf("got %+v", r)
	}
}

type stubNix struct{ ref string }

func (s stubNix) Lookup(name string) (string, error) {
	if name == "nixpkgs" {
		return s.ref, nil
	}
	return "", ErrNotFound
}

func TestResolve_NixRegistryFallback(t *testing.T) {
	r, err := Resolve("nixpkgs", Options{
		Registry:    registry.Empty(),
		NixRegistry: stubNix{ref: "github:NixOS/nixpkgs/nixos-unstable"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != SourceNixRegistry {
		t.Fatalf("expected nix-registry source, got %s", r.Source)
	}
	if r.Ref != "github:NixOS/nixpkgs/nixos-unstable" {
		t.Fatalf("ref = %s", r.Ref)
	}
}

func TestResolve_NotFoundSuggests(t *testing.T) {
	reg, _ := registry.Parse([]byte(`{
		"schema": 1,
		"packages": {"firefox": {"flake": "f"}, "fish": {"flake": "g"}}
	}`))
	_, err := Resolve("fir", Options{Registry: reg, NixRegistry: NoNixRegistry{}})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "firefox") {
		t.Errorf("expected suggestion containing firefox, got: %v", err)
	}
}

func TestParseNixRegistry(t *testing.T) {
	out := `
user   flake:nixpkgs           github:NixOS/nixpkgs/nixos-unstable
global flake:home-manager      github:nix-community/home-manager
global flake:helix             github:helix-editor/helix
`
	got, err := parseNixRegistry(out, "helix")
	if err != nil || got != "github:helix-editor/helix" {
		t.Fatalf("got (%q, %v)", got, err)
	}
	if _, err := parseNixRegistry(out, "nonexistent"); err == nil {
		t.Error("expected not-found error")
	}
}
