---
title: 'ADR-004: Resolver chain'
sidebar_label: 'ADR-004 Resolver chain'
---

# ADR-004 — Resolver chain: URI → glixos registry → nix registry

**Status:** Accepted

## Context

Users want short names (`glix add firefox`) but full flake URIs must remain
authoritative. A curated registry adds value without locking out arbitrary
flakes.

## Decision

Resolution order for `glix add <ref>`:

1. Anything that looks like a flake URI (`github:`, `gitlab:`, `git+`,
   `path:`, `http(s)://`, contains `#`) is used verbatim.
2. The glixos registry (cached JSON, configurable URL).
3. The user's `nix registry list` entries (unless `--no-nix-registry`).
4. Error with suggestions.

## Consequences

- Curated names get a fast path; everything else still works.
- The registry is just a JSON blob hosted anywhere; there is no server to
  operate.
- `--registry-url` and `--refresh` give users an escape hatch for testing
  alternate registries.
- The chain is symmetric for `glix search`, so the same names resolve in
  both commands.

## Trust model

Because the registry maps short names like `glix add firefox` to flake
URIs that ultimately reach `nix flake metadata` and `nix flake lock`,
**users trust whoever serves the registry as much as they'd trust them
to install a package on the machine.** A registry host that becomes
hostile — or an attacker who can MITM the connection — can silently
redirect any short-name to any flake URI on the public internet.

A few practical consequences of this position:

- HTTPS to the registry endpoint protects against passive MITM but
  does **not** authenticate the contents themselves. A compromised
  origin or an attacker with a valid certificate for the registry host
  can serve a malicious `registry.json` and the client has no way to
  notice.
- Short-names are not stable identities. The same name can point at
  different flake URIs across versions of the registry. Treat the
  short-name as a convenience; the flake URI in your committed
  `flake.lock` is the authoritative reference.
- `--registry-url` and the `GLIX_REGISTRY_URL` env var let users
  point at an alternate registry. There is no built-in pinning or
  signature check; if you self-host, the trust boundary moves to your
  infrastructure.
- `file://` registry URLs are gated behind `--allow-file-registry`.
  See the `glix add` / `glix search` help output. The opt-in exists so
  a glix invocation in a privileged or shared-tenancy context cannot
  be tricked into reading arbitrary local files through a `file://`
  URL chosen by a less-trusted actor.

### Future work

Signing the registry document and verifying it client-side against a
hardcoded public key would close the MITM and origin-compromise gaps
above. Candidates considered: **minisign** (small dependency, simple
key model), **age signatures**, and a **Nix-store-import-style
detached signature** (most idiomatic for the ecosystem). None of these
are implemented for pre-1.0; the right time to choose is when there is
a single canonical registry endpoint with operators committed to key
rotation. Until then this ADR documents the trust boundary so users can
reason about it explicitly.
