# TODO — tloc

Post-v1.0.0 findings from the 2026-07-22 adversarial review and independent
confirmation pass are recorded in
`reviews/2026-07-22-adversarial-review-v1.0.0.md`.

## Open items

None. Every confirmed or partially confirmed item from the adjudicated review
has been implemented and moved to `DONE.md`.

## Refuted / no action planned

- Minor 5 was refuted against scc v3.7.0 source
  (`processor/file.go:132`): scc also skips `.gitignore`/`.ignore` files unless
  its separate `--count-ignore` flag is set. tloc's default is scc parity and
  the spec never requested `--count-ignore`.
- Minor 9, a rebuild-divergence window on partial npm re-publish, remains
  theoretical and strongly mitigated by the pinned toolchain and trimpath.
- Minor 14, invalid UTF-8 being counted as replacement-character text, remains
  a low-impact corner case.
- Minor 17, a real `(root files)` directory displaying a trailing slash, is
  cosmetic and deliberate.
