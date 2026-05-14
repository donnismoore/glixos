# Pure function that turns a glix.toml manifest + a set of flake inputs into
# two lists of modules: one for NixOS, one for home-manager.
#
# Called from the user-packages flake, e.g.:
#
#   let
#     m = glixos-core.lib.importManifest {
#       manifestPath = ./hosts/${hostname}/glix.toml;
#       inherit inputs;
#     };
#   in {
#     nixosConfigurations.${hostname} = nixpkgs.lib.nixosSystem {
#       modules = [ glixos-core.nixosModules.glixos ] ++ m.systemModules;
#       ...
#     };
#   }
#
# This file is *static*. glix never writes here.
{ lib }:
{ manifestPath
, inputs ? { }
}:
let
  raw =
    if builtins.pathExists manifestPath
    then builtins.fromTOML (builtins.readFile manifestPath)
    else { schema = 1; packages = { }; settings = { }; };

  manifest =
    let v = raw.schema or 0;
    in
    if v == 1 then raw
    else throw "glix: unsupported manifest schema ${toString v} (expected 1) at ${toString manifestPath}";

  enabled = lib.filterAttrs (_: p: (p.enabled or true)) (manifest.packages or { });

  byScope = scope:
    lib.filterAttrs (_: p: (p.scope or "system") == scope) enabled;

  systemPkgs = byScope "system";
  homePkgs = byScope "home";

  # Fallback wrapper when a flake doesn't ship its own module.
  mkSystemPkgModule = input: { pkgs, ... }: {
    environment.systemPackages = [ input.packages.${pkgs.system}.default ];
  };

  mkHomePkgModule = input: { pkgs, ... }: {
    home.packages = [ input.packages.${pkgs.system}.default ];
  };

  requireInput = name:
    inputs.${name} or (throw
      "glix: manifest references package '${name}' but no matching flake input was found. Did you run `glix add` without rebuilding?");

  resolveSystem = name: _:
    let i = requireInput name;
    in i.nixosModules.default or (mkSystemPkgModule i);

  resolveHome = name: _:
    let i = requireInput name;
    in i.homeModules.default or (mkHomePkgModule i);
in
{
  inherit manifest;
  systemModules = lib.mapAttrsToList resolveSystem systemPkgs;
  homeModules = lib.mapAttrsToList resolveHome homePkgs;
}
