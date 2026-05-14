# Aggregate "glixos core" NixOS module.
#
# Imported by every glixos host. Composes the focused submodules in this
# directory. Do not put substantive logic here — add it to the right submodule
# (or create a new one) so each concern stays isolated.
#
# TODO(M3): once the glix CLI is packaged, include it in
# environment.systemPackages here.
{ lib, ... }:
{
  imports = [
    ./nix.nix
    ./boot.nix
    ./networking.nix
    ./users.nix
    ./locale.nix
  ];

  # Host can override; setting a default avoids eval errors on first build.
  # NixOS release this stateVersion is pinned to. See:
  # https://nixos.org/manual/nixos/stable/options#opt-system.stateVersion
  system.stateVersion = lib.mkDefault "24.11";
}
