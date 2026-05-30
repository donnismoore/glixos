# OS branding: identify the system as glixos at every OS boundary.
#
# Without this, /etc/os-release, hostnamectl, and (more importantly) the
# BUG_REPORT_URL surfaced by GUI crash handlers and `xdg-open`-style tools
# all point at NixOS/nixpkgs. That misroutes bug reports and erases the
# project's identity.
#
# Everything here is lib.mkDefault / lib.mkForce-overridable so a downstream
# host can re-claim plain NixOS branding if it wants to.
{ lib, config, ... }:
let
  homeUrl    = "https://donnismoore.github.io/glixos/";
  supportUrl = "https://donnismoore.github.io/glixos/user/getting-started";
  docsUrl    = "https://donnismoore.github.io/glixos/";
  bugUrl     = "https://github.com/donnismoore/glixos/issues";
in
{
  options.glixos.meta = {
    cliVersion = lib.mkOption {
      type = lib.types.str;
      default = "unknown";
      description = ''
        Version of the glix CLI that produced this user-packages repo. Set
        from the user-packages flake (e.g. via specialArgs) so it appears
        in /etc/glixos-release for bug reports.
      '';
    };
    schema = lib.mkOption {
      type = lib.types.int;
      default = 1;
      description = "glix.toml schema version this host was built against.";
    };
    commit = lib.mkOption {
      type = lib.types.str;
      default = "unknown";
      description = ''
        Commit hash of the user-packages repo at build time. Set from the
        user-packages flake via `self.rev or self.dirtyRev`.
      '';
    };
  };

  config = {
    system.nixos = {
      distroId   = lib.mkDefault "glixos";
      distroName = lib.mkDefault "glixos";
      variant_id = lib.mkDefault "manifest";
    };

    # nixpkgs' version.nix hardcodes BUG_REPORT_URL/SUPPORT_URL/DOCUMENTATION_URL
    # to upstream NixOS. mkForce the whole file so our URLs win. We still pull
    # the version/codename from system.nixos so the file stays accurate across
    # nixpkgs bumps.
    environment.etc."os-release".text = lib.mkForce ''
      NAME="glixos"
      PRETTY_NAME="glixos ${config.system.nixos.release} (${config.system.nixos.codeName})"
      ID=glixos
      ID_LIKE=nixos
      BUILD_ID="${config.system.nixos.version}"
      VERSION="${config.system.nixos.release} (${config.system.nixos.codeName})"
      VERSION_CODENAME=${lib.toLower config.system.nixos.codeName}
      VERSION_ID="${config.system.nixos.release}"
      VARIANT_ID=manifest
      HOME_URL="${homeUrl}"
      SUPPORT_URL="${supportUrl}"
      DOCUMENTATION_URL="${docsUrl}"
      BUG_REPORT_URL="${bugUrl}"
      ANSI_COLOR="1;34"
    '';

    # Bug-report companion file: distro-specific identifiers that issue
    # templates can ask for directly.
    environment.etc."glixos-release".text = ''
      GLIXOS_CLI_VERSION=${config.glixos.meta.cliVersion}
      GLIXOS_SCHEMA=${toString config.glixos.meta.schema}
      GLIXOS_COMMIT=${config.glixos.meta.commit}
      NIXOS_RELEASE=${config.system.nixos.release}
      NIXOS_VERSION=${config.system.nixos.version}
    '';
  };
}
