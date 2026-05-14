package manifest

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	m := New()
	m.Packages["zlib-thing"] = Package{
		Flake:   "github:owner/zlib-thing",
		Scope:   ScopeSystem,
		Enabled: true,
	}
	m.Packages["alpha"] = Package{
		Flake:   "path:./pkg-alpha",
		Scope:   ScopeHome,
		Enabled: false,
		Pin:     "sha256-abc",
	}

	var buf bytes.Buffer
	if err := Encode(&buf, m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "glix.toml")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !reflect.DeepEqual(m, got) {
		t.Fatalf("round-trip mismatch:\n  want %#v\n  got  %#v", m, got)
	}
}

func TestRoundTrip_M6Fields(t *testing.T) {
	m := New()
	m.Settings.System = "aarch64-linux"
	m.Settings.PrimaryUser = "alice"
	m.Packages["foo"] = Package{
		Flake:   "github:o/r",
		Scope:   ScopeHome,
		Enabled: true,
		Pin:     "deadbeef",
		User:    "bob",
	}

	var buf bytes.Buffer
	if err := Encode(&buf, m); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "glix.toml")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, got) {
		t.Fatalf("M6-field round-trip mismatch:\n  want %#v\n  got  %#v", m, got)
	}
}

func TestRoundTrip_Config(t *testing.T) {
	m := New()
	m.Packages["foo"] = Package{
		Flake:   "github:o/r",
		Scope:   ScopeSystem,
		Enabled: true,
		Config: map[string]string{
			"message": "hi there",
			"level":   "info",
		},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, m); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "glix.toml")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, got) {
		t.Fatalf("config round-trip mismatch:\n  want %#v\n  got  %#v", m, got)
	}
	// Deterministic key order: level before message alphabetically.
	out := buf.String()
	iL, iM := indexOf(out, "level"), indexOf(out, "message")
	if !(iL < iM) {
		t.Fatalf("config keys not sorted:\n%s", out)
	}
}

func TestDeterministicOrder(t *testing.T) {
	m := New()
	for _, n := range []string{"charlie", "alpha", "bravo"} {
		m.Packages[n] = Package{Flake: "x", Scope: ScopeSystem, Enabled: true}
	}
	var b1, b2 bytes.Buffer
	if err := Encode(&b1, m); err != nil {
		t.Fatal(err)
	}
	if err := Encode(&b2, m); err != nil {
		t.Fatal(err)
	}
	if b1.String() != b2.String() {
		t.Fatal("encode output is not deterministic")
	}
	// Alphabetical: alpha appears before bravo before charlie.
	out := b1.String()
	if iA, iB, iC := indexOf(out, "[packages.alpha]"), indexOf(out, "[packages.bravo]"), indexOf(out, "[packages.charlie]"); !(iA < iB && iB < iC) {
		t.Fatalf("packages not sorted alphabetically:\n%s", out)
	}
}

func TestPackageNameFromRef(t *testing.T) {
	cases := []struct {
		ref  string
		want string
		ok   bool
	}{
		{"github:owner/repo", "repo", true},
		{"github:owner/repo/branch", "repo", true},
		{"github:owner/repo?dir=subpkg", "subpkg", true},
		{"path:./examples/pkg-hello", "pkg-hello", true},
		{"path:../pkg-foo", "pkg-foo", true},
		{"https://example.com/x/y/foo.git", "foo", true},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := PackageNameFromRef(c.ref)
		if ok != c.ok || got != c.want {
			t.Errorf("PackageNameFromRef(%q) = (%q, %v), want (%q, %v)", c.ref, got, ok, c.want, c.ok)
		}
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
