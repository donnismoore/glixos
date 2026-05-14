# Example host config. Mirrors core/hosts/vm so the demo is bootable as a
# VM via `nix build .#nixosConfigurations.example.config.system.build.vm`.
{ lib, manifest, modulesPath, ... }:
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

  networking.networkmanager.enable = lib.mkForce false;
  networking.useDHCP = lib.mkForce true;
  networking.hostName = "glixos-example";

  users.users.alice = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    initialPassword = "alice";
  };

  # Wire the home-scope modules from the manifest into home-manager.
  home-manager = {
    useGlobalPkgs = true;
    useUserPackages = true;
    users.alice.imports = manifest.homeModules ++ [
      { home.stateVersion = "24.11"; }
    ];
  };

  services.openssh = {
    enable = true;
    settings.PermitRootLogin = "no";
  };

  system.stateVersion = "24.11";
}
