# Glixos reference package: ships a homeModules.default.
#
# Used to verify the module path in importManifest: when a flake exposes
# homeModules.default, importManifest imports it directly into the user's
# home-manager configuration.
{
  description = "glixos example: home module installing cowsay and a greeting file";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    homeModules.default = { pkgs, ... }: {
      home.packages = [ pkgs.cowsay ];
      home.file.".glixos-greeting".text =
        "Hello from a glixos package home module.\n";
    };
  };
}
