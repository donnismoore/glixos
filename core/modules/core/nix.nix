# Nix daemon configuration. Enforces flakes-only (ADR-007).
{ lib, ... }:
{
  nix = {
    settings = {
      experimental-features = [ "nix-command" "flakes" ];
      auto-optimise-store = true;
      trusted-users = [ "root" "@wheel" ];
      warn-dirty = false;
    };

    # ADR-007: glixos is flakes-only.
    channel.enable = false;

    gc = {
      automatic = true;
      dates = "weekly";
      options = "--delete-older-than 30d";
    };

    optimise.automatic = true;
  };

  nixpkgs.config.allowUnfree = lib.mkDefault true;
}
