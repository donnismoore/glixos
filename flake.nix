{
  description = "glixos: modular flake-driven NixOS";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      # Public library used by the user-packages flake.
      lib = {
        importManifest =
          import ./core/lib/importManifest.nix { inherit (nixpkgs) lib; };
      };

      # NixOS modules consumed by user-packages flakes. The aggregate
      # `glixos` module is a function of `inputs` so it can apply the
      # overlay that exposes `pkgs.glix` to every importing host.
      nixosModules = {
        glixos = import ./core/modules/core inputs;
        desktop = import ./core/modules/desktop;
        default = self.nixosModules.glixos;
      };

      # Overlay that adds `pkgs.glix`. Applied automatically by the
      # glixos module; downstream consumers needing `pkgs.glix` outside
      # a glixos host can apply it directly.
      overlays.default = final: prev: {
        glix = self.packages.${prev.system}.glix;
      };

      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        {
          glix = pkgs.buildGoModule {
            pname = "glix";
            version = self.shortRev or self.dirtyShortRev or "dev";
            src = ./glix;
            vendorHash = "sha256-pbA/AlBz3cQYRTMnQ/qBPcinYOKokrBLNhkbRTq54gE=";
            ldflags = [
              "-s"
              "-w"
              "-X main.version=${self.shortRev or self.dirtyShortRev or "dev"}"
              "-X main.commit=${self.rev or self.dirtyRev or "dirty"}"
            ];
            meta = with pkgs.lib; {
              description = "glixos manifest-driven package CLI";
              homepage = "https://github.com/powerreddude/glixos";
              license = licenses.gpl3Only;
              mainProgram = "glix";
              platforms = platforms.linux;
            };
          };
          default = self.packages.${system}.glix;
        }
        // nixpkgs.lib.optionalAttrs (system == "x86_64-linux") {
          vm = self.nixosConfigurations.vm.config.system.build.vm;
        });

      # Internal smoke-test host. Not the user-facing pattern; the
      # user-packages flake defines its own hosts.
      nixosConfigurations.vm = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        specialArgs = { inherit inputs; };
        modules = [
          self.nixosModules.glixos
          ./core/hosts/vm
        ];
      };

      checks = forAllSystems (system:
        {
          glix = self.packages.${system}.glix;
        }
        // nixpkgs.lib.optionalAttrs (system == "x86_64-linux") {
          vm-toplevel = self.nixosConfigurations.vm.config.system.build.toplevel;
        });

      formatter = forAllSystems (system:
        nixpkgs.legacyPackages.${system}.nixpkgs-fmt);
    };
}
