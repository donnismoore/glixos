package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sampleJSON = `{
  "schema": 1,
  "packages": {
    "hello":    {"flake": "github:owner/hello",    "description": "Greeter"},
    "helix":    {"flake": "github:helix-editor/helix", "description": "Postmodern modal text editor"},
    "firefox":  {"flake": "github:glixos/pkg-firefox"}
  }
}`

func TestParse(t *testing.T) {
	r, err := Parse([]byte(sampleJSON))
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Packages["helix"].Flake; got != "github:helix-editor/helix" {
		t.Fatalf("helix flake = %q", got)
	}
	if got := r.Packages["firefox"].Description; got != "" {
		t.Fatalf("firefox description should be empty, got %q", got)
	}
}

func TestParse_WrongSchema(t *testing.T) {
	if _, err := Parse([]byte(`{"schema": 99}`)); err == nil {
		t.Fatal("expected error for unsupported schema")
	}
}

func TestSearch(t *testing.T) {
	r, _ := Parse([]byte(sampleJSON))
	if got := r.Search("hel"); len(got) != 2 {
		t.Fatalf("expected 2 hits for 'hel', got %d: %+v", len(got), got)
	}
	// Prefix match wins ordering.
	got := r.Search("hel")
	if got[0].Name != "helix" && got[0].Name != "hello" {
		t.Fatalf("unexpected first hit: %s", got[0].Name)
	}
}

func TestLoad_FromFileURL(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(regPath, []byte(sampleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	l := &Loader{
		URL:           "file://" + regPath,
		CachePath:     filepath.Join(dir, "cache.json"),
		AllowFileURLs: true,
	}
	r, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(r.Packages))
	}
	if _, err := os.Stat(l.CachePath); err != nil {
		t.Errorf("cache not written: %v", err)
	}
}

func TestLoad_FileURL_DeniedByDefault(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(regPath, []byte(sampleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	l := &Loader{
		URL:       "file://" + regPath,
		CachePath: filepath.Join(dir, "cache.json"),
	}
	if _, err := l.Load(); err == nil {
		t.Fatal("expected file:// URL to be refused without AllowFileURLs")
	}
}

func TestLoad_HTTPBodySizeCap(t *testing.T) {
	// Serve a body larger than MaxRegistrySize and confirm we reject it
	// rather than buffering the whole thing.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		big := make([]byte, MaxRegistrySize+1024)
		for i := range big {
			big[i] = 'a'
		}
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := &Loader{URL: srv.URL, CachePath: filepath.Join(dir, "cache.json"), TTL: time.Hour}
	if _, err := l.Load(); err == nil {
		t.Fatal("expected oversize registry response to be rejected")
	}
}

func TestLoad_HTTPAndCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleJSON))
	}))
	defer srv.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "cache.json")
	l := &Loader{URL: srv.URL, CachePath: cache, TTL: time.Hour}

	r1, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(r1.Packages) != 3 {
		t.Fatalf("first load: expected 3 packages, got %d", len(r1.Packages))
	}
	// Stop the server; cache should still work.
	srv.Close()
	r2, err := l.Load()
	if err != nil {
		t.Fatalf("second load (cache): %v", err)
	}
	if len(r2.Packages) != 3 {
		t.Fatalf("second load: expected 3 packages, got %d", len(r2.Packages))
	}
}

func TestLoad_OfflineFallback(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache.json")
	if err := os.WriteFile(cache, []byte(sampleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make cache stale.
	old := time.Now().Add(-48 * time.Hour)
	_ = os.Chtimes(cache, old, old)

	warns := []string{}
	l := &Loader{
		URL:       "http://127.0.0.1:1", // unreachable
		CachePath: cache,
		TTL:       time.Hour,
		Warn:      func(s string) { warns = append(warns, s) },
	}
	r, err := l.Load()
	if err != nil {
		t.Fatalf("expected fallback to cache, got error: %v", err)
	}
	if len(r.Packages) != 3 {
		t.Fatalf("offline fallback: expected 3 packages, got %d", len(r.Packages))
	}
	if len(warns) == 0 {
		t.Error("expected at least one warning on offline fallback")
	}
}

func TestSuggest(t *testing.T) {
	r, _ := Parse([]byte(sampleJSON))
	got := r.Suggest("fox", 5)
	if len(got) != 1 || got[0] != "firefox" {
		t.Fatalf("expected [firefox], got %+v", got)
	}
}

func TestEmpty_Roundtrip(t *testing.T) {
	e := Empty()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Packages) != 0 {
		t.Fatal("expected empty packages map")
	}
}
