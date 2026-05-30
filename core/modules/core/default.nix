inputs:
{ config, lib, pkgs, ... }:
let
  cfg = config.glixos.glix;
in
{
  imports = [
    ./nix.nix
    ./boot.nix
    ./networking.nix
    ./users.nix
    ./locale.nix
    ../branding
  ];

  options.glixos.glix = {
    enable = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = ''
        Install the glix CLI system-wide. Disable on hosts that only
        consume the rendered configuration and never run `glix`
        themselves (CI builders, minimal appliances).
      '';
    };
    package = lib.mkPackageOption pkgs "glix" { };
  };

  config = {
    # Expose `pkgs.glix` to every host that imports nixosModules.glixos
    # so the `cfg.package` default below resolves without forcing each
    # host to wire the overlay by hand.
    nixpkgs.overlays = [ inputs.self.overlays.default ];

    environment.systemPackages = lib.mkIf cfg.enable [ cfg.package ];

    # NixOS release this stateVersion is pinned to. Host can override.
    # https://nixos.org/manual/nixos/stable/options#opt-system.stateVersion
    system.stateVersion = lib.mkDefault "24.11";
  };
}
