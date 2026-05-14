# Glixos reference package: ships a homeModules.default.
#
# Used to verify the module path in importManifest: when a flake exposes
# homeModules.default, importManifest imports it directly into the user's
# home-manager configuration.
{
  description = "glixos example: home module installing cowsay and a greeting file";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    # glixConfig is injected by importManifest from [packages.<name>.config]
    # in the host's glix.toml. It is a flat attrset of strings; this module
    # treats every key as optional and supplies a default.
    homeModules.default = { pkgs, glixConfig ? { }, ... }: {
      home.packages = [ pkgs.cowsay ];
      home.file.".glixos-greeting".text =
        (glixConfig.message or "Hello from a glixos package home module.") + "\n";
    };
  };
}
