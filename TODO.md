# TODO — tloc

Post-v1.0.0 fixes from the 2026-07-22 adversarial review, as adjudicated by the
independent confirmation pass (both recorded in
`reviews/2026-07-22-adversarial-review-v1.0.0.md`, including the adjudication
section — file:line references, repros, and fix directions live there). Every
primary finding below has been independently reproduced. Fix with a regression
test; move completed items to `DONE.md`.

Completed v1 work remains recorded in `DONE.md`.

## v1.0.1 blockers

- [ ] C1 (critical, confirmed): walker error handler is inverted
  (`internal/analyze/analyze.go:254` returns false; gocodewalker treats true as
  continue). One unreadable directory silently drops files — at depth ≥ 2 it also
  aborts readable sibling directories — with exit 0 and empty stderr.
  Implementation notes from adjudication: flipping the boolean is not sufficient —
  the callback must collect warnings concurrency-safely (it is invoked from walker
  goroutines), and collected walk warnings must NOT flow through the current fatal
  error path (or they re-create M1). Note gocodewalker discards first-level
  traversal errors internally regardless of handler return. Regression tests for
  unreadable directories at depth 1 and depth ≥ 2 (skip where ACL denial isn't
  testable).
- [ ] M2 (medium, confirmed data loss): `-o` pointed at an existing source file
  (including Windows case-variants) silently truncates it with the report.
  Adjudicated fix direction: default no-clobber on existing files + explicit
  `--force`, rather than detecting "source-like" files (reports can themselves be
  JSON/CSV/Markdown). DECISION NEEDED before implementing: default no-clobber
  breaks the currently-verified repeat-run workflow (`tloc -o report.csv .` twice,
  byte-stable) — choose between always requiring `--force` on existing files, or a
  carve-out for files tloc previously wrote. Regression test the case-variant path.
- [ ] M4 (medium, confirmed): by-file tabular truncates path tails
  (`internal/output/tabular.go:67`), so long paths render as identical rows and
  absolute-path scans collapse many rows to one prefix. Trim from the head (keep
  the filename) as scc does. JSON/CSV unaffected.

## Blocked on product decision — M1 (medium; behavior confirmed)

One unreadable/locked file mid-scan currently suppresses the entire report
(`analyze.go:186-191` + `internal/app/app.go:56-59`), exit 1, no output. This is
fail-closed (not silently wrong like C1) — downgraded from major in adjudication.
Adjudicated design (recommended): render the partial report, warn per-file on
stderr, exit NONZERO, and add structured completeness metadata to JSON/CSV
(`complete: false` + skipped entries) so machine consumers can't mistake a partial
report for a full one. Note this deliberately breaks scc parity (scc prints and
continues with exit 0) in favor of safety — confirm exit-code semantics with Shaun
before implementing, then coordinate with C1's warning channel.

## Fast-follow (docs / calibration)

- [ ] M3 (downgraded to docs+coverage in adjudication; numbers verified honest,
  LOO already disclosed in DONE.md): narrow the "across common languages" claim,
  document that unrepresented languages ride the global factor unvalidated, and —
  stronger follow-up — add held-out samples (and ideally new languages: C, HTML,
  Kotlin, Swift, …) to the calibration corpus rather than merely rewording.
- [ ] SPEC.md still contains stale npm-dispute language (calls the unscoped `tloc`
  package abandoned and anticipates a name dispute) contradicting the recorded
  decision in DONE.md to retain `@shaunobi/tloc` and file no claim. Update SPEC.
- [ ] `--help` states Claude counts may differ but omits the "roughly 10%" accuracy
  target that SPEC.md explicitly requires --help to state. Add it.

## Queued (minor — see review doc §Minor and §Adjudication for details)

- [ ] Folder-view width cap counts indentation; deep rows degrade to bare "…"
  (Minor 1).
- [ ] README: document that folder identity in JSON/CSV is the composite key
  (input_id, folder, synthetic) — the flat `folder` path alone can collide with a
  real `(root files)` directory (Minor 2, narrowed to docs in adjudication).
- [ ] File passed as input renders as a self-containing folder under `--by-folder`,
  conflicting with "rows are folders only" (Minor 3).
- [ ] README: document overlapping-input double-counting (intentional, scc/spec
  consistent) (Minor 4).
- [ ] Calibration contract test, narrowed scope (Minor 6): fail on empty
  generation sample sets (0/0 → NaN passes vacuously) and re-hash corpus files
  against calibration.json. (Ranks asset already independently SHA-pinned;
  "post-hoc cutoff" charge withdrawn as speculation.)
- [ ] npm publish.mjs: `npm.cmd` fallback throws EINVAL on modern Node without
  `shell: true` — reproduced on Node 24/Windows (Minor 7).
- [ ] release.yml: transient `gh release view` failures read as "not published",
  which can mint a duplicate draft for an already-published tag (Minor 8).
- [ ] platform.js missing-package error: mention regenerating cross-platform
  lockfiles/node_modules as a cause and fix (Minor 10).
- [ ] ci.yml: same-repo PR branches run CI twice (push + pull_request) (Minor 11).
- [ ] README: `cmd` examples fail verbatim in this repo; make examples runnable
  (Minor 12).
- [ ] README: state explicitly that folder CSV has no totals row and cumulative
  rows must not be summed (Minor 13, narrowed — cumulative semantics are already
  documented).
- [ ] tools/calibrate: bound numeric and date-form `Retry-After` to the same cap as
  computed backoff (Minor 15).
- [ ] Preflight `-o` destination writability before scanning WITHOUT truncating an
  existing destination (adjudication suggests same-directory temp file + rename)
  (Minor 16; coordinate with M2's no-clobber semantics).
- [ ] Hardening (from adjudication): npm stage.mjs trusts GoReleaser goos/goarch
  metadata labels for artifact→package mapping; consider verifying binary headers
  (PE/ELF/Mach-O magic + arch) at staging time.

## Refuted / no action planned (record here if that changes)

- Minor 5 REFUTED in adjudication and verified against scc v3.7.0 source
  (`processor/file.go:132`): scc also skips `.gitignore`/`.ignore` files unless its
  separate `--count-ignore` flag is set; tloc's default IS scc parity and the spec
  never requested `--count-ignore`. No change.
- Minor 9: rebuild-divergence window on partial npm re-publish — theoretical,
  strongly mitigated by pinned toolchain/trimpath.
- Minor 14: invalid UTF-8 counted as U+FFFD-replaced text — corner case.
- Minor 17: real `(root files)` dir always displays trailing `/` — cosmetic,
  deliberate.
