# Boot-related defaults that aren't bootloader-specific.
# Hosts must choose their own bootloader (grub / systemd-boot / etc.).
{ ... }:
{
  boot.tmp.cleanOnBoot = true;
  boot.loader.timeout = 3;
}
