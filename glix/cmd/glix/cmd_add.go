package main

import (
	"fmt"
	"os"

	"github.com/glixos/glix/internal/flake"
	"github.com/glixos/glix/internal/manifest"
	"github.com/glixos/glix/internal/nix"
	"github.com/glixos/glix/internal/resolver"
)

func cmdAdd(args []string) error {
	fs := newFlagSet("add")
	host := fs.String("host", currentHostname(), "host whose manifest to mutate")
	dir := fs.String("dir", "", "user-packages repo root (default: discovered)")
	scopeFlag := fs.String("scope", "", "package scope: system or home (default from manifest)")
	nameFlag := fs.String("name", "", "override package name (default: inferred)")
	userFlag := fs.String("user", "", "target user for scope=home (default: settings.primary_user)")
	pinFlag := fs.String("pin", "", "lock to this revision via ?rev=<pin> (github/gitlab/sourcehut)")
	apply := fs.Bool("apply", false, "run `nixos-rebuild switch` after staging")
	dryRun := fs.Bool("dry-run", false, "print planned changes; do not write")
	regURL := fs.String("registry-url", "", "override registry URL")
	refresh := fs.Bool("refresh", false, "force refetch of the registry before resolving")
	noNix := fs.Bool("no-nix-registry", false, "skip the `nix registry list` fallback")
	allowLocal := fs.Bool("allow-local-path", false, "permit an absolute local-path flake ref (path:/file:/absolute); for testing against a local checkout only")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: glix add <ref-or-name> [flags]")
	}
	input := fs.Arg(0)

	r, err := resolveRepo(*dir)
	if err != nil {
		return err
	}
	if !r.HostExists(*host) {
		return fmt.Errorf("host %q not found in %s (run `glix init --host=%s` first)", *host, r.Root, *host)
	}

	// Resolve short name → canonical URI via the chain. URIs pass through unchanged.
	url, err := resolveRegistryURL(*dir, *host, *regURL)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(url, *refresh)
	if err != nil {
		return err
	}
	opts := resolver.Options{Registry: reg}
	if *noNix {
		opts.NixRegistry = resolver.NoNixRegistry{}
	} else {
		opts.NixRegistry = resolver.NixCLI{}
	}
	res, err := resolver.Resolve(input, opts)
	if err != nil {
		return err
	}
	if !*allowLocal && flake.IsAbsoluteLocalRef(res.Ref) {
		return fmt.Errorf(
			"refusing to add %q: it resolves to an absolute local-path flake ref, which bakes a machine-specific path into your committed flake.nix and breaks every downstream clone. Pass --allow-local-path if you really intend this (e.g. for local development).",
			res.Ref)
	}

	m, err := manifest.Load(r.ManifestPath(*host))
	if err != nil {
		return err
	}

	name := *nameFlag
	if name == "" {
		switch res.Source {
		case resolver.SourceURI:
			n, ok := manifest.PackageNameFromRef(res.Ref)
			if !ok {
				return fmt.Errorf("could not infer package name from %q; pass --name", res.Ref)
			}
			name = n
		default:
			name = res.ShortName
		}
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

	if *userFlag != "" && scope != manifest.ScopeHome {
		return fmt.Errorf("--user only applies to scope=home (got scope=%s)", scope)
	}
	if *userFlag != "" {
		if err := requireValidIdent("user", *userFlag); err != nil {
			return err
		}
	}

	pkg := manifest.Package{
		Flake:   res.Ref,
		Scope:   scope,
		Enabled: true,
		Pin:     *pinFlag,
		User:    *userFlag,
	}

	if *dryRun {
		fmt.Fprintf(os.Stdout, "would add package %q -> %s (source=%s, scope=%s)\n", name, res.Ref, res.Source, scope)
		return nil
	}

	snap, err := r.TakeSnapshot(*host)
	if err != nil {
		return err
	}

	m.Packages[name] = pkg
	if err := manifest.Save(r.ManifestPath(*host), m); err != nil {
		_ = snap.Restore()
		return err
	}
	if err := regenerateFlake(r); err != nil {
		_ = snap.Restore()
		return err
	}
	if err := nix.FlakeLock(r.Root); err != nil {
		if rerr := snap.Restore(); rerr != nil {
			return fmt.Errorf("nix flake lock failed (%w) and rollback also failed: %v", err, rerr)
		}
		return fmt.Errorf("nix flake lock failed; rolled back: %w", err)
	}
	if err := r.Commit(fmt.Sprintf("glix add %s: %s (%s, %s, via %s)", *host, name, res.Ref, scope, res.Source)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if res.Source == resolver.SourceURI {
		fmt.Printf("added %s -> %s (scope=%s) to host %s\n", name, res.Ref, scope, *host)
	} else {
		fmt.Printf("added %s -> %s (scope=%s, via %s) to host %s\n", name, res.Ref, scope, res.Source, *host)
	}

	if *apply || m.Settings.AutoApply {
		fmt.Println("applying with nixos-rebuild switch...")
		return nix.Rebuild(r.Root, *host, "switch")
	}
	fmt.Println("staged. run `glix rebuild` to apply.")
	return nil
}
