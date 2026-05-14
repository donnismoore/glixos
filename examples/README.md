# examples/

Reference flakes that exercise the glixos package contract.

| Path             | Demonstrates                                           |
|------------------|--------------------------------------------------------|
| `pkg-hello/`     | `packages.${system}.default` fallback path (no module).|
| `pkg-greeting/`  | `homeModules.default` with `glixConfig` opt-in.        |
| `user-packages/` | A hand-written user-packages flake (pre-`glix init`).  |

These exist for two purposes:

1. **Documentation.** They're the simplest possible example of each
   contract path; copy them as a starting point for your own flake.
2. **Smoke testing.** `glix add path:./examples/pkg-hello` exercises
   the URI passthrough and fallback resolution end-to-end without
   touching the network.

## Trying them

From a freshly-`glix init`-ed repo:

```bash
glix add path:/abs/path/to/glixos/examples/pkg-hello
glix add --scope=home path:/abs/path/to/glixos/examples/pkg-greeting
glix set greeting config.message="Hello, world!"
glix rebuild switch
```

## See also

- [Packages (user docs)](https://powerreddude.github.io/glixos/user/packages)
- [Flake contract (contributor docs)](https://powerreddude.github.io/glixos/contributor/flake-contract)
