package main

import (
	"fmt"
	"os"

	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
)

func cmdAdd(args []string) error {
	fs := newFlagSet("add")
	host := fs.String("host", currentHostname(), "host whose manifest to mutate")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	scopeFlag := fs.String("scope", "", "package scope: system or home (default from manifest)")
	nameFlag := fs.String("name", "", "override package name (default: inferred from ref)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after staging")
	dryRun := fs.Bool("dry-run", false, "print planned changes; do not write")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: glix add <flake-ref> [flags]")
	}
	ref := fs.Arg(0)

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	if !r.HostExists(*host) {
		return fmt.Errorf("host %q not found in %s (run `glix init --host=%s` first)", *host, r.Root, *host)
	}

	m, err := manifest.Load(r.ManifestPath(*host))
	if err != nil {
		return err
	}

	name := *nameFlag
	if name == "" {
		n, ok := manifest.PackageNameFromRef(ref)
		if !ok {
			return fmt.Errorf("could not infer package name from %q; pass --name", ref)
		}
		name = n
	}
	if err := requireValidIdent("package name", name); err != nil {
		return err
	}
	if _, exists := m.Packages[name]; exists {
		return fmt.Errorf("package %q already present; remove it first or pass --name", name)
	}

	scope := manifest.Scope(*scopeFlag)
	if scope == "" {
		scope = m.Settings.DefaultScope
	}
	if !scope.Valid() {
		return fmt.Errorf("invalid --scope %q (want system or home)", scope)
	}

	pkg := manifest.Package{
		Flake:   ref,
		Scope:   scope,
		Enabled: true,
	}

	if *dryRun {
		fmt.Fprintf(os.Stdout, "would add package %q -> %s (scope=%s) at %s\n", name, ref, scope, r.ManifestPath(*host))
		return nil
	}

	m.Packages[name] = pkg
	if err := manifest.Save(r.ManifestPath(*host), m); err != nil {
		return err
	}
	if err := regenerateFlake(r); err != nil {
		return err
	}
	if err := nix.FlakeLock(r.Root); err != nil {
		return fmt.Errorf("nix flake lock: %w", err)
	}
	if err := r.Commit(fmt.Sprintf("glix add %s: %s (%s, %s)", *host, name, ref, scope)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Printf("added %s -> %s (scope=%s) to host %s\n", name, ref, scope, *host)

	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	fmt.Println("staged. run `glix rebuild` to apply.")
	return nil
}
