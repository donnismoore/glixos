{
  description = "glixos-core: the immutable OS layer of glixos";

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
          import ./lib/importManifest.nix { inherit (nixpkgs) lib; };
      };

      # Modules consumed by user-packages flakes.
      nixosModules = {
        glixos = import ./modules/core;
        desktop = import ./modules/desktop;
        default = self.nixosModules.glixos;
      };

      # Internal smoke-test host. NOT the user-facing pattern;
      # the user-packages flake defines its own hosts.
      nixosConfigurations.vm = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        specialArgs = { inherit inputs; };
        modules = [
          self.nixosModules.glixos
          ./hosts/vm
        ];
      };

      # `nix build .#vm` produces a runnable QEMU VM.
      packages = forAllSystems (system: nixpkgs.lib.optionalAttrs (system == "x86_64-linux") {
        vm = self.nixosConfigurations.vm.config.system.build.vm;
        default = self.nixosConfigurations.vm.config.system.build.vm;
      });

      # `nix flake check` will build this.
      checks = forAllSystems (system: nixpkgs.lib.optionalAttrs (system == "x86_64-linux") {
        vm-toplevel = self.nixosConfigurations.vm.config.system.build.toplevel;
      });

      formatter = forAllSystems (system:
        nixpkgs.legacyPackages.${system}.nixpkgs-fmt);
    };
}
