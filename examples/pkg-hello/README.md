# pkg-hello

Reference flake that exposes only `packages.${system}.default`. Used to
verify the **fallback path** in `importManifest`: when a flake exposes
no `nixosModules.default` or `homeModules.default`, glix wraps
`packages.default` into `environment.systemPackages` (for scope
`system`) or `home.packages` (for scope `home`).

## Use it

```bash
glix add path:/abs/path/to/examples/pkg-hello
# or, if you've published this repo:
glix add github:donnismoore/glixos?dir=examples/pkg-hello
```

By default it lands in `environment.systemPackages` (scope `system`).

## Source

```nix
{
  description = "glixos example: bare hello package";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system: {
        default = nixpkgs.legacyPackages.${system}.hello;
      });
    };
}
```

That's the whole flake. Any flake of this shape is a valid glixos
package.

## See also

- [Flake contract](https://donnismoore.github.io/glixos/contributor/flake-contract)
- [pkg-greeting](../pkg-greeting/) — the `homeModules.default` path.
