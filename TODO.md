# TODO — tloc v1

Only unfinished work remains here. Completed items and implementation notes are
in `DONE.md`.

## Release (remote access and account setup required)

- [ ] Authenticate GitHub and npm, create the public `shaunobi/tloc` repository, push the completed code, bootstrap all seven scoped npm packages, and configure each package's trusted publisher for `.github/workflows/release.yml`.
- [ ] Tag `v1.0.0`, run the draft → npm → finalized GitHub release workflow end to end, and verify `go install`, a downloaded release binary, global npm installation, and `npx @shaunobi/tloc`.

## Post-v1 follow-ups (not release blockers)

- [ ] Homebrew tap (after first tagged release).
- [ ] File npm support dispute for the abandoned unscoped `tloc` name; if granted, move the package to unscoped.
