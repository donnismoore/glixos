package flake

import "testing"

func TestLocalPath(t *testing.T) {
	cases := []struct {
		ref      string
		wantPath string
		wantOK   bool
	}{
		{"", "", false},
		{"github:o/r", "", false},
		{"https://example.com/x.tar.gz", "", false},
		{"git+https://example.com/r.git", "", false},
		{"path:/home/me/x", "/home/me/x", true},
		{"path:./modules/foo", "./modules/foo", true},
		{"path:../sibling", "../sibling", true},
		{"path:/home/me/x?dir=core", "/home/me/x", true},
		{"file:///home/me/x", "/home/me/x", true},
		{"file:/home/me/x", "/home/me/x", true},
		{"/home/me/x", "/home/me/x", true},
		{"./relative", "./relative", true},
		{"../up", "../up", true},
	}
	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			got, ok := LocalPath(c.ref)
			if ok != c.wantOK || got != c.wantPath {
				t.Errorf("LocalPath(%q) = (%q, %v); want (%q, %v)", c.ref, got, ok, c.wantPath, c.wantOK)
			}
		})
	}
}

func TestIsAbsoluteLocalRef(t *testing.T) {
	cases := []struct {
		ref  string
		want bool
	}{
		{"github:o/r", false},
		{"https://x/y", false},
		{"path:/home/me/x", true},
		{"path:./modules/foo", false},
		{"path:../up", false},
		{"file:///home/me/x", true},
		{"/home/me/x", true},
		{"./rel", false},
	}
	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			if got := IsAbsoluteLocalRef(c.ref); got != c.want {
				t.Errorf("IsAbsoluteLocalRef(%q) = %v; want %v", c.ref, got, c.want)
			}
		})
	}
}

func TestEscapesRoot(t *testing.T) {
	root := "/repo"
	cases := []struct {
		name          string
		ref, flakeDir string
		wantEscapes   bool
		wantIsLocal   bool
	}{
		{"non-local github", "github:o/r", "/repo", false, false},
		{"non-local https", "https://x/y", "/repo", false, false},
		{"abs inside", "path:/repo/sub", "/repo", false, true},
		{"abs outside", "path:/home/me/x", "/repo", true, true},
		{"rel inside", "path:./modules/foo", "/repo", false, true},
		{"rel inside nested flake dir", "path:./foo", "/repo/sub", false, true},
		{"rel escapes", "path:../other", "/repo", true, true},
		{"file:// outside", "file:///etc/passwd", "/repo", true, true},
		{"with query inside", "path:/repo/sub?dir=x", "/repo", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			escapes, isLocal := EscapesRoot(c.ref, root, c.flakeDir)
			if escapes != c.wantEscapes || isLocal != c.wantIsLocal {
				t.Errorf("EscapesRoot(%q, %q, %q) = (escapes=%v, isLocal=%v); want (escapes=%v, isLocal=%v)",
					c.ref, root, c.flakeDir, escapes, isLocal, c.wantEscapes, c.wantIsLocal)
			}
		})
	}
}

func TestScanInputURLs(t *testing.T) {
	src := `{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    glixos-core = {
      url = "path:/home/wonnis/glixos/core";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    home-manager = {
      url   =   "github:nix-community/home-manager"  ;
    };
  };
}`
	got := ScanInputURLs(src)
	want := []string{
		"github:NixOS/nixpkgs/nixos-unstable",
		"path:/home/wonnis/glixos/core",
		"github:nix-community/home-manager",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d urls, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("url[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}
