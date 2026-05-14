package flake

import (
	"strings"
	"testing"

	"github.com/glixos/glix/internal/manifest"
)

const sample = `{
  description = "x";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # >>> glix-managed inputs >>>
    # <<< glix-managed inputs <<<
  };
}
`

func TestReplaceRegion_PreservesIndent(t *testing.T) {
	body := "foo = {\n  url = \"github:o/r\";\n};\n"
	out, err := ReplaceRegion(sample, RegionInputs, body)
	if err != nil {
		t.Fatal(err)
	}
	wantSubs := []string{
		`    # >>> glix-managed inputs >>>`,
		`    foo = {`,
		`      url = "github:o/r";`,
		`    };`,
		`    # <<< glix-managed inputs <<<`,
	}
	for _, s := range wantSubs {
		if !strings.Contains(out, s) {
			t.Errorf("missing line %q in output:\n%s", s, out)
		}
	}
}

func TestReplaceRegion_Idempotent(t *testing.T) {
	m := manifest.New()
	m.Packages["foo"] = manifest.Package{Flake: "github:o/r", Scope: manifest.ScopeSystem, Enabled: true}
	body := RenderInputs(m)
	once, err := ReplaceRegion(sample, RegionInputs, body)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := ReplaceRegion(once, RegionInputs, body)
	if err != nil {
		t.Fatal(err)
	}
	if once != twice {
		t.Fatal("ReplaceRegion is not idempotent")
	}
}

func TestReplaceRegion_RoundTripToEmpty(t *testing.T) {
	m := manifest.New()
	m.Packages["foo"] = manifest.Package{Flake: "github:o/r", Scope: manifest.ScopeSystem, Enabled: true}
	withFoo, err := ReplaceRegion(sample, RegionInputs, RenderInputs(m))
	if err != nil {
		t.Fatal(err)
	}
	empty, err := ReplaceRegion(withFoo, RegionInputs, "")
	if err != nil {
		t.Fatal(err)
	}
	if empty != sample {
		t.Fatalf("expected to recover original after clearing region.\n--- got ---\n%s\n--- want ---\n%s", empty, sample)
	}
}

func TestReplaceRegion_MissingMarker(t *testing.T) {
	_, err := ReplaceRegion("no markers here", RegionInputs, "x")
	if err == nil {
		t.Fatal("expected error for missing marker")
	}
}

func TestRenderInputs_AlphabeticalOrder(t *testing.T) {
	m := manifest.New()
	m.Packages["zeta"] = manifest.Package{Flake: "z", Scope: manifest.ScopeSystem, Enabled: true}
	m.Packages["alpha"] = manifest.Package{Flake: "a", Scope: manifest.ScopeSystem, Enabled: true}
	got := RenderInputs(m)
	if strings.Index(got, "alpha") > strings.Index(got, "zeta") {
		t.Fatalf("inputs not alphabetical:\n%s", got)
	}
}
