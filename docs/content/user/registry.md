---
title: Registry
sidebar_position: 8
---

# Registry

The glixos registry is a JSON file that maps short names to flake refs.
It's the second step in the resolver chain (after URI passthrough, before
`nix registry list`); see [ADR-004](../adr/ADR-004-resolver-chain).

## Shape

```json
{
  "schema": 1,
  "packages": {
    "hello": {
      "flake": "github:donnismoore/glixos?dir=examples/pkg-hello",
      "description": "GNU Hello example"
    },
    "greeting": {
      "flake": "github:donnismoore/glixos?dir=examples/pkg-greeting",
      "description": "Demo home module"
    }
  }
}
```

There is no server. The registry is just a file at an HTTPS URL. The
default lives in this repo under `registry/registry.json`.

## Configuring a registry

Per host, in `glix.toml`:

```toml
[settings]
registry_url = "https://raw.githubusercontent.com/donnismoore/glixos/main/registry/registry.json"
```

Per invocation:

```bash
glix add --registry-url=https://example.com/my-registry.json firefox
glix search --refresh hello
```

## Caching

The fetched JSON is cached under `$XDG_CACHE_HOME/glix/`. Use `--refresh`
to force a refetch:

```bash
glix search --refresh hello
glix add --refresh hello
```

## Self-hosting

Any HTTPS endpoint that serves the JSON schema above works. Common
approaches:

- A file in a public git repo, served via `raw.githubusercontent.com`.
- A static site behind your own domain.
- A file inside an internal-only HTTP server for private deployments.

## Bypassing entirely

If you never want to use the registry, always pass full flake refs:

```bash
glix add github:owner/repo
glix add path:./local/flake
```

`glix add` only consults the registry for short names; URIs go straight
through.
