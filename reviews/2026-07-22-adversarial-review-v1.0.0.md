# Adversarial review — tloc v1.0.0

Date: 2026-07-22. Scope: full codebase, release artifacts, and DONE.md claims, as of
tag `v1.0.0` (working tree at commit `d10a529`). Method: six independent adversarial
reviewers (core pipeline, aggregation/sorting, output renderers, tokenizer/calibration,
release engineering, docs/spec-compliance), each building the binary and proving
findings end-to-end against hostile fixtures where possible. Critical findings were
independently re-reproduced before inclusion. Findings below are pending independent
confirmation before fixes land.

## Verdict

Quality is genuinely high and DONE.md is honest: every externally checkable claim
verified true (public repo; v1.0.0 release with six archives + checksums; seven npm
packages at `latest=1.0.0` published via OIDC trusted publishing with no token secrets;
every GitHub Action SHA-pin resolving to its commented tag; embedded o200k ranks file
byte-identical to OpenAI's official pin, SHA-256
`446a9538cb6c348e3516120d7c08b09f57c36495e2acfffe59a5bf8b0cfb1a2d`; all 80 calibration
corpus counts and every headline accuracy figure reproducing exactly from the
checked-in JSON). The suite passes clean: gofmt, vet, all Go packages, all 14 npm
packaging tests, all 9 golden files genuinely byte-compared.

However, v1.0.0 shipped with one critical and one major defect, both in scan error
policy, plus one medium data-loss footgun. These merit a v1.0.1.

---

## Critical

### C1. Inverted walker error handler — silently wrong counts, exit 0

`internal/analyze/analyze.go:254` — `walker.SetErrorHandler(func(error) bool { return false })`.

gocodewalker's contract is "return true to continue where possible"; scc itself prints
the error and returns true. Returning false makes the library abort silently:

- Unreadable directory at depth 1 relative to an input root: that subtree's files
  silently vanish from the count.
- Unreadable directory at depth ≥ 2: the walk of **unrelated readable sibling
  directories** in that subtree is also aborted.

Either way `walker.Start()` returns nil and tloc exits 0 with no warning.

Reproduced twice, independently (Windows, `icacls` deny-read on one directory):

- Depth 1: 3-file fixture reports **2 files**, exit 0, no stderr.
- Depth 2 (`top/mid/a_denied`, `top/mid/z_after`): 3-file fixture reports **1 file** —
  the readable `z_after/y.go` also vanished. Exit 0, no stderr.

Permission-denied directories are routine (junction points, OneDrive placeholders,
service dirs). A counting tool producing silently wrong counts is the worst failure
mode. This also falsifies DONE.md's "scc ignore parity" claim for any tree containing
a walk error — parity holds only on error-free trees.

**Fix direction:** return true from the handler (scc parity); collect and surface
warnings on stderr; consider a strict mode flag for fail-hard behavior.

## Major

### M1. One unreadable file mid-scan suppresses the entire report

`internal/analyze/analyze.go:186-191` (Run returns records and joined error) plus
`internal/app/app.go:56-59` (Main discards records whenever err != nil).

A single file that fails open/read between discovery and processing — deleted
mid-scan, or share-locked, which is routine on Windows (editors, AV, log writers) —
prints one error and exits 1 with **no report at all**, discarding every successfully
counted file. Reproduced with a file held open under `FileShare.None`.

Combined with C1 the error policy is incoherent in both directions: walk errors are
swallowed (wrong counts, exit 0); read errors are fatal (all output destroyed, exit 1).

**Fix direction:** same policy as C1 — count what can be counted, warn on stderr per
failed file, render the report; strict mode to opt into fail-hard.

## Medium

### M2. `-o` pointed directly at a source file silently destroys it

`internal/analyze/analyze.go` output-exclusion logic (~455-473) + `internal/app/app.go:87`.

The alias guards correctly reject the output path passed as an input file and
hardlink aliases of scanned files. But `tloc -o proj/hello.rs proj` (or a Windows
case-variant, `-o proj/HELLO.RS`) excludes the file from the scan and then
**overwrites the source file with the report**, exit 0. README's "rejects an output
path that aliases a scanned source file" reads as broader protection than exists.

**Fix direction:** refuse to overwrite an existing file whose extension/content
identifies it as a scannable source file unless forced, or at minimum document the
exact semantics precisely.

### M3. Claude-accuracy claims are in-sample; SPEC's ~10% target not met out-of-sample for SQL

`tools/calibrate/results/calibration.md`, DONE.md (calibration section), README accuracy note.

The headline "worst represented-language MAPE 9.30%" evaluates per-language overrides
on the same 5 files they were fitted on. Leave-one-out (the honest unseen-code
estimate) puts SQL at **10.52%**, above the spec's ~10% target. Languages absent from
the 16-language corpus (C, HTML, CSS, Kotlin, Swift, Dockerfile, …) ride the global
factor with zero evidence, and measured per-language deviations (C# +15%, SQL +21% vs
global) show unrepresented languages can plausibly exceed 10%. All arithmetic
reproduces exactly from the checked-in JSON — the issue is framing, not fabrication.

**Fix direction:** docs change — lead with LOO numbers, state the unrepresented-
language caveat in README and calibration.md.

### M4. By-file tabular truncation drops the filename

`internal/output/tabular.go:16,67,177-185` — paths over 48 chars are truncated from
the tail (`runewidth.Truncate`), so files sharing a long directory prefix render as
byte-identical rows; with absolute input paths, every row can collapse to the same
truncated prefix. scc trims from the front precisely so the filename survives.

**Fix direction:** head-trimming (keep the tail) for path columns, or widen/uncap.

## Minor

1. **Deep-tree `…` rows** — `internal/output/tabular.go:80-81`: indentation counts
   against the 48-char folder width cap, so at depth ≳ 22 rows render as
   indistinguishable bare `…`. (Found independently by two reviewers.)
2. **`(root files)` path collision in JSON/CSV** — `internal/aggregate/aggregate.go:150`:
   a real depth-1 directory named `(root files)` emits a byte-identical `folder` path
   to the synthetic row; only the `synthetic` column disambiguates. DONE.md's
   "collision-safe" holds only for consumers keying on (path, synthetic).
3. **File inputs under `--by-folder`** — `aggregate.go:126-159`: a file passed
   directly as an input renders as a depth-0 "folder" named after the file with a
   synthetic `(root files)` child duplicating its metrics.
4. **Overlapping inputs double-count** — `tloc . ./src` counts shared files once per
   input in language/total rows, silently. Matches scc and arguably SPEC's
   "multiple paths aggregate"; deserves a README note.
5. **Ignore files never counted** — `analyze.go:249,314-317`: `.gitignore`/`.ignore`/
   `.sccignore` are excluded unconditionally, even under `--no-ignore --no-gitignore`;
   scc counts them as languages. Deliberate deviation; undocumented.
6. **Calibration contract-test gaps** — `tools/calibrate/calibration_contract_test.go`:
   (a) an empty legacy sample set passes vacuously (0/0 → NaN, `NaN > 5.0` false);
   (b) the test never re-hashes the corpus files or the embedded ranks asset against
   calibration.json, so silent drift stays green; (c) the pinned "selection rule"
   constants (`LOO ≤ 11`) sit just above SQL's 10.52% — post-hoc, documented only
   qualitatively.
7. **npm `runNpm` Windows fallback broken** — `npm/scripts/publish.mjs:47-49`:
   `spawnSync("npm.cmd", …)` without `shell: true` throws EINVAL on every supported
   Node (CVE-2024-27980 hardening). Unreachable in Linux CI; affects local Windows
   runs with unusual npm layouts.
8. **Release re-run can mint a duplicate draft** — `.github/workflows/release.yml:42-52`:
   `gh release view … || true` maps transient API failures to "not published", taking
   the fresh-draft path for an already-published tag. Manual cleanup only; no data loss.
9. **Rebuild-divergence window on partial npm re-publish** — `release.yml:54-79` +
   `npm/scripts/publish.mjs:134-137`: a re-run rebuilds GitHub archives while
   already-published npm platform packages keep first-run bytes. Strongly mitigated
   (pinned toolchain go1.25.12, GoReleaser v2.17.0, `-trimpath`, `mod_timestamp`,
   CGO off); theoretical.
10. **Platform-package error message omits lockfile fix** —
    `npm/packages/tloc/lib/platform.js:53-58`: the most common real-world cause
    (cross-platform `package-lock.json`, npm/cli#4828) isn't mentioned in the
    remediation hint.
11. **CI double-runs same-repo PRs** — `.github/workflows/ci.yml:3-8`: `push` on all
    branches plus `pull_request` runs everything twice, including the GoReleaser
    snapshot preflight. Cost/noise only.
12. **README examples reference a nonexistent `cmd/` directory** — README.md:62,89:
    `tloc cmd internal README.md` fails verbatim in this repo (`internal/` exists,
    `cmd/` does not), so the examples look repo-specific but aren't.
13. **Folder-CSV semantics undocumented** — folder rows are cumulative and no CSV view
    emits a totals row (consistent and pinned by test), but README never warns
    consumers not to sum the column; a spreadsheet SUM over folder rows silently
    inflates ~2x+.
14. **Invalid UTF-8 counted as U+FFFD-replaced text** — `internal/tokenizer/o200k.go:98-102`:
    Latin-1-encoded files (which pass the NUL-based binary skip) count replacement
    characters rather than byte-exact o200k semantics. Corner case; real APIs can't
    receive such bytes anyway.
15. **Uncapped `retry-after` in calibration tool** — `tools/calibrate/anthropic.go:110-112`:
    a hostile/buggy `retry-after: 99999` sleeps ~28h where computed backoff caps at 8s.
    Offline tool, context-aware, ctrl-C works.
16. **Bad `-o` path discovered only after full scan** — `internal/app/app.go:87`: the
    output file opens after scanning, so a long scan is wasted on an unwritable path.
17. **Real `(root files)` dir always displays trailing `/`** — `aggregate.go:195-199`:
    even with no synthetic sibling. Cosmetic; deliberate per comment.

## DONE.md claims: overstated (wording, not substance)

- "Collision-safe real `(root files)` directories" — true only with the `synthetic`
  column; flat path identity is ambiguous (Minor 2).
- "Worst represented-language MAPE is 9.30%" — in-sample; LOO worst is 10.52% (M3).
- CI runs "formatting, vet, tests on Linux/macOS/Windows" — vet and tests yes; the
  gofmt check is a separate ubuntu-only job (the right engineering call; the sentence
  overreads).

## Verified clean (highlights)

- **Aggregation:** cumulative rollups exact in every fixture (root rows sum to
  totals); five-key sorting a genuine total order, arrival-order independent;
  Tok/Line guards code==0; mutual exclusion `--by-file`/`--by-folder` exits 2.
- **Rendering:** unicode/CJK alignment correct in all views; JSON int64 end-to-end,
  valid under hostile names; CSV quoting per encoding/csv; goldens byte-compared,
  9-view matrix real; output-file exclusion byte-stable across repeated in-tree runs;
  hardlink/symlink aliases of the output rejected.
- **Pipeline:** size cap exact at boundary, TOCTOU growth bounded (cap+1 read);
  single read per file verified; extension-filter semantics (case, compound `.d.ts`)
  match scc exactly; no deadlocks; usage errors exit 2, runtime errors exit 1.
- **Tokenizer:** embedded ranks authentic o200k_base (hash matches OpenAI's pin);
  all unit vectors match a reference tiktoken build; all 80 corpus files' bytes,
  SHA-256, and o200k counts reproduce with 0 mismatches; `codex` alias exactly o200k;
  estimator rounding/overflow/override lookup correct — override labels verified
  against scc's canonical `languages.json` labels.
- **Calibration math:** global factors 29256/17684=1.654377 and 22109/17684=1.250226,
  production MAPE 5.854%/4.083%, per-language and LOO figures all recomputed from the
  per-file JSON exactly; Opus 4.7 / Fable 5 identical content counts confirmed in data.
- **Release engineering:** all six action SHAs verified against the GitHub API; no
  workflow injection (tag reaches shell only as quoted env var); OIDC publish proven
  by the live 1.0.0; exact optionalDependency pinning verified in the published
  tarballs (including 0o755 bin permissions, which also repairs upload-artifact's
  permission stripping); artifact→package mapping cannot cross-wire platforms;
  wrapper published strictly last; prerelease→`next` dist-tag verified live;
  musl correctly needs no special handling (CGO off, static binaries); dist/ not
  tracked; templates stay private with no lifecycle scripts.
- **Spec compliance:** every flag, default, alias, exit code, output column, footer,
  and caveat in SPEC.md tested against the built binary and confirmed; README
  contains no competitor comparisons; `--help` covers every flag with the estimate
  caveat.

## Recommendation

Ship v1.0.1 with C1 + M1 fixed under one coherent error policy (continue past
per-file/per-directory errors, warn on stderr, optional strict mode) and the M2
output guard. M3 is a docs edit. Everything else can queue behind those.
