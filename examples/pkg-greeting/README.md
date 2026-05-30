# pkg-greeting

Reference flake that ships a `homeModules.default`. Used to verify the
**module path** in `importManifest`: when a flake exposes
`homeModules.default`, glix imports it directly into the target user's
home-manager configuration instead of falling back to
`home.packages`/`environment.systemPackages`.

It also demonstrates the **per-package config** mechanism: the module
reads `glixConfig.message`, which is populated from
`[packages.greeting.config]` in `glix.toml`.

## Use it

```bash
glix add --scope=home path:/abs/path/to/examples/pkg-greeting
# or
glix add --scope=home github:donnismoore/glixos?dir=examples/pkg-greeting

# Set the greeting text:
glix set pkg-greeting config.message="Hello, world!"

# Apply:
sudo glix rebuild switch
```

After the rebuild, the target user's home directory will contain
`~/.glixos-greeting` with the configured text, and `cowsay` will be on
their `$PATH`.

## Source

```nix
{
  description = "glixos example: home module installing cowsay and a greeting file";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    # glixConfig is injected by importManifest from [packages.<name>.config]
    # in the host's glix.toml. It is a flat attrset of strings; this module
    # treats every key as optional and supplies a default.
    homeModules.default = { pkgs, glixConfig ? { }, ... }: {
      home.packages = [ pkgs.cowsay ];
      home.file.".glixos-greeting".text =
        (glixConfig.message or "Hello from a glixos package home module.") + "\n";
    };
  };
}
```

The `glixConfig ? { }` default is the contract: the module must work
when no `[packages.<name>.config]` table exists in the manifest.

## See also

- [Flake contract](https://donnismoore.github.io/glixos/contributor/flake-contract)
- [Per-package config](https://donnismoore.github.io/glixos/user/config)
- [pkg-hello](../pkg-hello/) — the fallback (no-module) path.
