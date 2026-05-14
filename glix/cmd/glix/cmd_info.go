package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/glixos/glix/internal/manifest"
)

// cmdInfo prints a high-level summary of the user-packages repo: discovered
// root, git status and HEAD, per-host manifest stats, and the resolved
// flake.lock input nodes. Read-only.
func cmdInfo(args []string) error {
	fs := newFlagSet("info")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	fmt.Printf("root    %s\n", r.Root)

	if r.HasGit() {
		clean, cerr := r.IsClean()
		switch {
		case cerr != nil:
			fmt.Printf("git     error: %v\n", cerr)
		case clean:
			fmt.Printf("git     clean\n")
		default:
			fmt.Printf("git     dirty\n")
		}
		if subj, herr := r.HeadSubject(); herr == nil && subj != "" {
			fmt.Printf("head    %s\n", subj)
		}
	} else {
		fmt.Printf("git     (no .git)\n")
	}

	hosts, err := r.ListHosts()
	if err != nil {
		return err
	}
	sort.Strings(hosts)
	fmt.Println("hosts:")
	if len(hosts) == 0 {
		fmt.Println("  (none)")
	}
	for _, h := range hosts {
		mp := r.ManifestPath(h)
		m, lerr := manifest.Load(mp)
		if lerr != nil {
			fmt.Printf("  %-12s [error: %v]\n", h, lerr)
			continue
		}
		var sys, primary string
		sys = m.Settings.System
		if sys == "" {
			sys = "(default)"
		}
		primary = m.Settings.PrimaryUser
		if primary == "" {
			primary = "(default)"
		}
		var enabled, total int
		total = len(m.Packages)
		for _, p := range m.Packages {
			if p.Enabled {
				enabled++
			}
		}
		fmt.Printf("  %-12s system=%s primary_user=%s packages=%d/%d enabled\n",
			h, sys, primary, enabled, total)
	}

	fmt.Println("flake.lock:")
	lockPath := filepath.Join(r.Root, "flake.lock")
	b, err := os.ReadFile(lockPath)
	if err != nil {
		fmt.Printf("  (missing or unreadable: %v)\n", err)
		return nil
	}
	var lock struct {
		Nodes map[string]struct {
			Locked map[string]any `json:"locked"`
			Inputs any            `json:"inputs"`
		} `json:"nodes"`
		Root string `json:"root"`
	}
	if err := json.Unmarshal(b, &lock); err != nil {
		fmt.Printf("  (parse error: %v)\n", err)
		return nil
	}
	names := make([]string, 0, len(lock.Nodes))
	for k := range lock.Nodes {
		if k == lock.Root {
			continue
		}
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		node := lock.Nodes[n]
		ident := lockIdent(node.Locked)
		fmt.Printf("  %-20s %s\n", n, ident)
	}
	return nil
}

// lockIdent returns a one-line identifier for a flake.lock "locked" entry.
// Falls back to the type/url if no revision is present.
func lockIdent(locked map[string]any) string {
	if locked == nil {
		return "(unresolved)"
	}
	typ, _ := locked["type"].(string)
	rev, _ := locked["rev"].(string)
	narHash, _ := locked["narHash"].(string)
	switch typ {
	case "github", "gitlab", "sourcehut":
		owner, _ := locked["owner"].(string)
		repo, _ := locked["repo"].(string)
		if rev != "" {
			return fmt.Sprintf("%s:%s/%s @ %s", typ, owner, repo, shortRev(rev))
		}
		return fmt.Sprintf("%s:%s/%s", typ, owner, repo)
	case "path":
		p, _ := locked["path"].(string)
		return "path:" + p
	case "tarball", "file":
		url, _ := locked["url"].(string)
		if rev != "" {
			return fmt.Sprintf("%s @ %s", url, shortRev(rev))
		}
		return url
	case "git":
		url, _ := locked["url"].(string)
		if rev != "" {
			return fmt.Sprintf("git:%s @ %s", url, shortRev(rev))
		}
		return "git:" + url
	}
	if rev != "" {
		return shortRev(rev)
	}
	if narHash != "" {
		return narHash
	}
	return "(unknown)"
}

func shortRev(rev string) string {
	if len(rev) > 12 {
		return rev[:12]
	}
	return rev
}
