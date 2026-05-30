# registry/

The default glixos registry. A single JSON file mapping short package
names to flake URIs so users can write:

```bash
glix add helix
```

instead of:

```bash
glix add github:helix-editor/helix
```

## Schema

```json
{
  "schema": 1,
  "packages": {
    "<name>": {
      "flake": "<flake-uri>",
      "description": "<one-line description>"
    }
  }
}
```

- `schema` — integer. Bumped only on a breaking shape change.
- `packages.<name>.flake` — any URI that `nix flake metadata` accepts
  (`github:`, `gitlab:`, `sourcehut:`, `path:`, `git+https:`…).
- `packages.<name>.description` — free-form, surfaced by `glix info`
  and `glix search`.

`<name>` must be a valid identifier: `[A-Za-z][A-Za-z0-9_-]*`. It's the
key glix uses everywhere: in `glix.toml`, in command-line arguments, and
as the `inputs.<name>` key in the generated `flake.nix`.

## Resolution order

`glix add <name>` resolves `<name>` in this order (first hit wins):

1. **URI passthrough** — if `<name>` parses as a flake URI (contains
   `:`, starts with `./` or `/`, etc.), it's used as-is and not looked
   up in any registry.
2. **glixos registry** — fetched from `registry.url` in
   `~/.config/glixos/glix.toml`, cached under
   `$XDG_CACHE_HOME/glixos/registry/`.
3. **`nix registry list`** — falls back to the user's Nix flake
   registry.

See [Resolver chain](https://donnismoore.github.io/glixos/contributor/resolver)
for the full algorithm.

## Default URL

`glix init` writes this into the host's `glix.toml`:

```toml
[registry]
url = "https://raw.githubusercontent.com/donnismoore/glixos/main/registry/registry.json"
ttl_hours = 24
```

Override per-host by editing `[registry]` directly, or globally by
setting `GLIX_REGISTRY_URL`.

## Self-hosting

Any HTTP(S) URL that serves the JSON above will work. To run your own:

1. Fork this repo or just publish a single `registry.json`.
2. Point each host's `glix.toml` at it:

   ```toml
   [registry]
   url = "https://example.com/my-registry.json"
   ttl_hours = 24
   ```

3. `glix update --registry` to refresh the on-disk cache immediately.

The cache key is the URL, so multiple registries coexist cleanly.

## Adding a package to this registry

1. Open `registry.json`.
2. Add an entry under `packages`, preserving alphabetical order by key.
3. Pick a `flake` URI that the upstream maintainer is comfortable
   pinning to. Don't include a `?rev=`; pinning is per-host.
4. Open a PR. CI validates the JSON against the schema.

There is no review of the *content* of registered flakes — the registry
is a naming convenience, not an endorsement. Users opt into a flake by
running `glix add`.

## See also

- [Registry (user docs)](https://donnismoore.github.io/glixos/user/registry)
- [Resolver chain (contributor docs)](https://donnismoore.github.io/glixos/contributor/resolver)
- [ADR-004 — Registry-style resolution](https://donnismoore.github.io/glixos/adr/ADR-004-registry-resolution)
