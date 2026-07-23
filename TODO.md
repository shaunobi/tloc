# TODO — tloc

Post-v1.0.0 fixes from the 2026-07-22 adversarial review
(`reviews/2026-07-22-adversarial-review-v1.0.0.md` — file:line references, repros, and
fix directions live there). Findings are pending independent confirmation: for each
item, first reproduce the issue (or refute it — if refuted, record why in DONE.md and
drop the fix), then fix with a regression test. Move completed items to `DONE.md`.

Completed v1 work remains recorded in `DONE.md`.

## v1.0.1 blockers

- [ ] C1 (critical): walker error handler is inverted (`internal/analyze/analyze.go:254`
  returns false; gocodewalker contract and scc return true). One unreadable directory
  silently drops files — at depth ≥ 2 it also aborts readable sibling directories —
  with exit 0 and no warning. Fix: return true, collect walk errors, surface them on
  stderr. Regression test with an unreadable directory at depth 1 and depth ≥ 2
  (skip on platforms where ACL denial isn't testable).
- [ ] M1 (major): one unreadable/locked file mid-scan suppresses the entire report
  (`analyze.go:186-191` + `internal/app/app.go:56-59` discard all records on any error).
  Fix: count what can be counted, warn on stderr per failed file, still render the
  report. Regression test with a share-locked file on Windows and a mid-scan delete.
- [ ] Unified error policy for C1+M1: decide exit-code semantics when warnings occurred
  (recommend exit 0 + stderr warnings) and add a strict-mode flag for fail-hard;
  document in README and --help.
- [ ] M2 (medium): `-o` pointed directly at an existing source file (including Windows
  case-variants) silently overwrites it with the report. Add a refusal/force guard or
  precisely document the semantics; regression test the case-variant path.

## v1.0.1 or fast-follow

- [ ] M3 (medium, docs): calibration accuracy claims are in-sample. Lead with
  leave-one-out numbers (SQL LOO is 10.52%, above the ~10% spec target), add the
  unrepresented-language caveat to README and calibration.md, and reconcile DONE.md's
  "worst 9.30%" framing.
- [ ] M4 (medium): by-file tabular truncates path tails so long paths render as
  identical rows; trim from the head (keep the filename) like scc, or widen/uncap.

## Queued (minor — see review doc §Minor for details)

- [ ] Folder-view width cap counts indentation; depth ≳ 22 rows degrade to bare "…"
  (review Minor 1).
- [ ] `(root files)` flat-path collision in JSON/CSV with a real directory of that
  name; make path identity unambiguous or document (path, synthetic) keying
  (Minor 2).
- [ ] File passed as input renders as a self-containing folder under `--by-folder`
  (Minor 3).
- [ ] Document overlapping-input double-counting in README (Minor 4).
- [ ] Document (or revisit) never counting `.gitignore`/`.ignore`/`.sccignore` even
  under `--no-ignore` — deliberate deviation from scc (Minor 5).
- [ ] Calibration contract test: fail on empty generation sample sets (NaN
  comparison passes vacuously); re-hash corpus files and the embedded ranks asset
  against calibration.json (Minor 6).
- [ ] npm publish.mjs: fix `npm.cmd` fallback (EINVAL on modern Node without
  `shell: true`) (Minor 7).
- [ ] release.yml: don't map transient `gh release view` failures to "not published"
  (duplicate-draft hazard) (Minor 8).
- [ ] platform.js missing-package error: mention the cross-platform lockfile cause
  and fix (Minor 10).
- [ ] ci.yml: stop double-running same-repo PRs (push + pull_request overlap)
  (Minor 11).
- [ ] README: fix examples referencing nonexistent `cmd/` directory (Minor 12).
- [ ] README: document folder-CSV cumulative rows / no totals row — do not SUM the
  column (Minor 13).
- [ ] Validate `-o` path is writable before scanning, not after (Minor 16).
- [ ] tools/calibrate: cap honored `retry-after` (Minor 15).

## Noted, no action planned (record here if that changes)

- Rebuild-divergence window on partial npm re-publish — strongly mitigated by pinned
  toolchain/trimpath (Minor 9).
- Invalid UTF-8 counted as U+FFFD-replaced text — corner case (Minor 14).
- Real `(root files)` dir always displays trailing `/` — cosmetic, deliberate
  (Minor 17).
