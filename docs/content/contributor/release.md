---
title: Release process
sidebar_position: 13
---

# Release process

How a milestone gets cut.

## Checklist

1. **Bump version.** `glix/cmd/glix/main.go` `const version = "0.1.0-mN"`.
2. **Run the pre-commit checks** (see [Build & test](./build-test)).
3. **Update the milestone log.** Add an entry to
   [Milestones](./milestones) describing what shipped.
4. **Smoke test end-to-end.** A fresh `glix init` in `/tmp` plus
   exercising every new command path.
5. **Commit.** Message: `MN: <one-line summary>`, no co-authors or
   non-glixos signatures.
6. **Tag.** `git tag v0.1.0-mN -m "MN: <summary>"`.
7. **Push.** `git push && git push --tags`.

## What `main` looks like after a milestone

```
git log --oneline -10
37845d9 M7: per-package config and repo introspection
52963d9 M6: multi-host, multi-user, pinning, gc
6cd8c61 M5: lifecycle commands (update, enable/disable, set, show, doctor, rollback)
ec00db6 M4: registry, resolver, and transactional rollback
0616888 M3: glix CLI MVP
7920ad4 M2: prove manifest contract end-to-end
...
```

Each commit is the **complete** milestone. There are no per-feature
sub-commits within a milestone — a milestone is the atomic unit on
`main`.

## Docs deployment

Pushing to `main` triggers `.github/workflows/docs.yml`, which builds
the Docusaurus site under `docs/` and deploys it to GitHub Pages at
<https://powerreddude.github.io/glixos>.

If the workflow fails, the site stays on the last successful build.
Check the Actions tab for the build log.

## Binary releases (future)

A future release workflow will:

1. Build `glix` for `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`,
   `aarch64-darwin`.
2. Compute checksums.
3. Attach binaries + checksums to the GitHub release for the tag.

Until that exists, users install via `go install` or `nix run`.

## Breaking changes

A breaking manifest change requires:

1. A new ADR explaining the change and its migration story.
2. Bumping `schema = N` in `manifest.go`.
3. Code in `manifest.Load` that detects the old schema and either
   migrates in place or returns a clear "run `glix migrate`" error.
4. A milestone dedicated to the migration (no other feature work in the
   same milestone).

Breaking changes to the CLI surface (renamed commands, removed flags)
follow the same principle: ADR + deprecation path + dedicated milestone.
