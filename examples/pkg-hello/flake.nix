# Glixos reference package: bare `packages.default`.
#
# Used to verify the fallback path in importManifest: when a flake exposes
# no nixosModules.default, glix wraps packages.${system}.default into
# environment.systemPackages (or home.packages for home scope).
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
