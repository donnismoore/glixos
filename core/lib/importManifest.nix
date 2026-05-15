# Pure function that turns a glix.toml manifest + a set of flake inputs into
# module lists ready to be wired into a NixOS configuration.
#
# Called from the user-packages flake, e.g.:
#
#   let
#     m = glixos-core.lib.importManifest {
#       manifestPath = ./hosts/${hostname}/glix.toml;
#       inherit inputs;
#       defaultUser = "donnis";
#     };
#   in {
#     nixosConfigurations.${hostname} = nixpkgs.lib.nixosSystem {
#       modules = [ glixos-core.nixosModules.glixos ] ++ m.systemModules;
#       ...
#     };
#   }
#
# Output attributes:
#   manifest          — the raw decoded TOML
#   systemModules     — list of nixos modules (one per enabled scope=system pkg)
#   homeModules       — flat list of home-manager modules (legacy single-user path)
#   homeModulesByUser — { <user> = [ module ... ]; ... } keyed by target user
#                       (Package.user, falling back to settings.primary_user,
#                       falling back to defaultUser).
#
# This file is *static*. glix never writes here.
{ lib }:
{ manifestPath
, inputs ? { }
, defaultUser ? "user"
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

  settings = manifest.settings or { };
  primaryUser = settings.primary_user or defaultUser;

  enabled = lib.filterAttrs (_: p: (p.enabled or true)) (manifest.packages or { });

  byScope = scope:
    lib.filterAttrs (_: p: (p.scope or "system") == scope) enabled;

  systemPkgs = byScope "system";
  homePkgs = byScope "home";

  mkSystemPkgModule = input: { pkgs, ... }: {
    environment.systemPackages = [ input.packages.${pkgs.system}.default ];
  };

  mkHomePkgModule = input: { pkgs, ... }: {
    home.packages = [ input.packages.${pkgs.system}.default ];
  };

  requireInput = name:
    inputs.${name} or (throw
      "glix: manifest references package '${name}' but no matching flake input was found. Did you run `glix add` without rebuilding?");

  # withGlixConfig wraps an arbitrary module so that, when evaluated, the
  # package's `[packages.<name>.config]` table is exposed as the
  # `glixConfig` module argument.
  #
  # Per-module scoping: we cannot use `_module.args.glixConfig = cfg` because
  # that arg is a single user-wide (or system-wide) value — multiple
  # sibling modules (e.g. several home-scope packages for one user) would
  # each define it, producing "defined multiple times" collisions on the
  # `raw`-typed `_module.args` attribute.
  #
  # Instead, when mod is a function we build a wrapper function whose
  # functionArgs are the standard NixOS module args (config/lib/pkgs/options)
  # — these are always populated by the module evaluator — and we forward
  # to mod with `glixConfig = cfg` merged in. The loader never tries to
  # resolve `glixConfig` from `_module.args`, so siblings cannot collide.
  withGlixConfig = cfg: mod:
    if builtins.isFunction mod
    then { config, lib, pkgs, options, ... }@args:
      mod (args // { glixConfig = cfg; })
    else mod;

  resolveSystem = name: p:
    let
      i = requireInput name;
      base = i.nixosModules.default or (mkSystemPkgModule i);
    in
    withGlixConfig (p.config or { }) base;

  resolveHome = name: p:
    let
      i = requireInput name;
      base = i.homeModules.default or (mkHomePkgModule i);
    in
    withGlixConfig (p.config or { }) base;

  # Per-package home module, tagged with its target user.
  homeEntries = lib.mapAttrsToList
    (name: p: {
      user = p.user or primaryUser;
      module = resolveHome name p;
    })
    homePkgs;

  homeModulesByUser = lib.foldl'
    (acc: e: acc // { ${e.user} = (acc.${e.user} or [ ]) ++ [ e.module ]; })
    { }
    homeEntries;
in
{
  inherit manifest;
  systemModules = lib.mapAttrsToList resolveSystem systemPkgs;
  homeModules = map (e: e.module) homeEntries;
  inherit homeModulesByUser;
}
