# Optional desktop profile. Stubbed for M1 — real wiring lands in M5+.
{ lib, config, ... }:
{
  options.glixos.desktop = {
    enable = lib.mkEnableOption "glixos desktop profile";
  };

  config = lib.mkIf config.glixos.desktop.enable {
    # TODO(M5): X/Wayland session, audio, display manager, fonts.
  };
}
