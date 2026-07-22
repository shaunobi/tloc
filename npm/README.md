# npm release packages

This directory contains the source templates and release tooling for the seven
packages that provide `tloc` through npm:

- `@shaunobi/tloc`, the JavaScript command shim.
- Six OS/architecture packages containing the Go release binaries.

The package templates deliberately use version `0.0.0`, contain no binaries,
and set `private: true`, so npm refuses to publish them directly.
`scripts/stage.mjs` reads GoReleaser's `dist/artifacts.json`, copies the six
binaries into `dist/npm`, applies the tag version to every package, and removes
the private marker only from those staged copies.

No public package defines an install lifecycle script. npm selects the matching
binary package through `os`, `cpu`, and `optionalDependencies`; installation
does not fetch an executable from a separate server.

## Local checks

```sh
npm test --prefix npm
```

After producing GoReleaser artifacts, package contents and publish-time policy
can be checked without publishing:

```sh
node npm/scripts/stage.mjs --version 1.0.0
node npm/scripts/publish.mjs --root dist/npm --version 1.0.0 --dry-run
```

## Publishing authentication

The tag workflow uses npm trusted publishing and needs `id-token: write`; it
does not need a long-lived npm token after bootstrap. Trusted publishing also
adds npm provenance automatically, so the publish script deliberately does not
force `--provenance`; this keeps token-authenticated local bootstrap viable.

An npm package must exist before its trusted publisher can be configured. To
create all seven packages for the first time:

1. From a clean checkout of the commit that will become `v1.0.0`, create a
   local-only bootstrap tag and build its exact artifacts without publishing a
   GitHub release:

   ```sh
   git tag v0.0.0-bootstrap.0
   goreleaser release --clean --skip=publish
   node npm/scripts/stage.mjs --version 0.0.0-bootstrap.0
   node npm/scripts/publish.mjs \
     --root dist/npm \
     --version 0.0.0-bootstrap.0 \
     --dry-run
   ```

2. Authenticate to npm with an account protected by 2FA, or a temporary
   granular token with bypass-2FA permission, then run the same publish command
   without `--dry-run`. Platform packages are created before the wrapper, and
   the bootstrap prerelease receives the non-default `next` dist-tag so it
   cannot satisfy a normal install before `v1.0.0` exists.
3. Configure each of the seven packages with npm CLI 11.15.0 or newer:

   ```sh
   npm trust github <package> \
     --repo shaunobi/tloc \
     --file release.yml \
     --allow-publish \
     --yes
   npm trust list <package> --json
   ```

   The workflow filename is case-sensitive and is not a path. Do not set an
   environment unless the workflow is updated to use the exact same environment
   name.
4. After one OIDC release succeeds, disallow token publishing for the packages
   and revoke any temporary bootstrap token.

The publish script checks the registry before each publish, skips package and
version pairs that already exist, publishes all platform packages first, and
publishes the wrapper last. Stable versions use the `latest` dist-tag;
prereleases use `next` so they cannot replace `latest`.

The release workflow runs the full CI suite, including a real GoReleaser
snapshot-to-npm preflight. GoReleaser then creates or reuses a GitHub draft and
hands its exact `dist` directory to a separate OIDC-only npm job. The draft is
published only after all npm packages succeed. If npm fails partway through,
use **Re-run failed jobs**: the npm job reuses the preserved artifacts and skips
versions already present in the registry.
