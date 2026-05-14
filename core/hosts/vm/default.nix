# Smoketest host. Built via `nix build .#vm` from the core flake.
# Not a template for user hosts — user hosts live in the user-packages flake.
{ lib, modulesPath, ... }:
{
  imports = [
    (modulesPath + "/profiles/qemu-guest.nix")
  ];

  boot.loader.grub = {
    enable = true;
    device = "/dev/vda";
    useOSProber = false;
  };

  fileSystems."/" = {
    device = "/dev/vda1";
    fsType = "ext4";
  };

  # NetworkManager is overkill in a VM; use plain DHCP.
  networking.networkmanager.enable = lib.mkForce false;
  networking.useDHCP = lib.mkForce true;
  networking.hostName = "glixos-vm";

  users.users.glixos = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    initialPassword = "glixos";
  };

  services.openssh = {
    enable = true;
    settings.PermitRootLogin = "no";
  };

  system.stateVersion = "24.11";
}
