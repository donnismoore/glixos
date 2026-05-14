# User-management defaults. No users are declared here; hosts must declare
# at least one normal user. mutableUsers is off so the system stays
# reproducible: passwords go through hashedPassword / agenix / sops-nix.
{ lib, ... }:
{
  users.mutableUsers = lib.mkDefault false;

  # Convenience for development; tighten in production hosts.
  security.sudo.wheelNeedsPassword = lib.mkDefault false;
}
