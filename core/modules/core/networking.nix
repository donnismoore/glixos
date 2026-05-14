# Default networking stack: NetworkManager + firewall on. Hosts override.
{ lib, ... }:
{
  networking.networkmanager.enable = lib.mkDefault true;
  networking.firewall.enable = lib.mkDefault true;

  # NetworkManager handles DHCP; the global useDHCP option is legacy.
  networking.useDHCP = lib.mkDefault false;
}
