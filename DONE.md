# DONE — tloc v1

Completed items moved from `TODO.md`. Notes capture what was implemented and
any deliberate deviation from `SPEC.md`.

## v1.0.1 adversarial-review fixes

- [x] Continue safely after unreadable files and directories (C1 + M1). — The
  concurrent walker callback now continues and collects deterministic warnings;
  inspect/read failures use the same recoverable channel. tloc renders every
  readable result, warns per skipped entry on stderr, and exits nonzero. JSON
  emits `complete` plus structured `skipped` metadata; CSV emits status columns
  and skipped-entry rows. Depth-one/deep ACL regressions and injected partial
  report tests cover the complete policy.
- [x] Make output replacement explicit and race-safe (M2 + Minor 16). — Existing
  destinations are refused unless `--force` is present. Writability is checked
  before scanning without truncation; new outputs use exclusive creation, and a
  forced existing output retains its verified handle through the scan, checks
  identity before and after writing, and truncates only that handle. Case
  variants, aliases, repeat runs, race-created paths, and retarget attempts are
  covered; `--force` never bypasses source-alias protection.
- [x] Preserve useful tabular labels (M4 + Minor 1). — By-file paths are
  grapheme- and display-width-safe trimmed from the head so filenames survive;
  folder indentation no longer consumes the folder-name width budget.
- [x] Represent direct file inputs honestly in folder view (Minor 3). — A direct
  file now produces exactly one depth-zero synthetic `(root files)` bucket and
  contributes its metrics once, with deterministic lexical parent paths.
- [x] Clarify machine identity and aggregation semantics (Minors 2, 4, 12, 13).
  — README examples now run in this repository; overlapping inputs are explicitly
  double-counted; folder identity is documented as
  `(input_id, folder, synthetic)`; cumulative folder CSV rows have no totals row
  and must not be summed.
- [x] Add true held-out Claude calibration evidence (M3). — Added 20 disjoint
  authored files across C, HTML, Kotlin, and Swift and measured them without
  using them to fit or select factors. Production-factor MAPE is 8.25% overall
  for current Claude and 4.03% for legacy; current-Claude HTML's 17.27% exception
  is disclosed. Existing factors and all prior 80-file measurements remain
  unchanged, and current/Fable ground truth matches across all 100 files.
- [x] Strengthen calibration reproducibility (Minor 6 + Minor 15). — The contract
  rejects empty generation sets, recollects and SHA-256 rehashes all fitting and
  held-out files, checks disjointness and recomputed summaries, and pins the
  held-out exception. Numeric, date-form, and computed retry delays are capped at
  eight seconds.
- [x] Correct stale specification and CLI accuracy text. — SPEC records the
  permanent scoped npm name and no-dispute decision. README, SPEC, and `--help`
  narrow the roughly-10% target to represented/validated workloads and disclose
  the unvalidated global fallback.
- [x] Harden Windows npm publication and installation guidance (Minors 7 + 10).
  — The `npm.cmd` fallback resolves and launches `npm-cli.js` with Node without
  shell parsing, including hostile-argument coverage and a real npm 11.16.0
  fallback run. Missing platform packages now explain cross-platform lockfile and
  `node_modules` remediation.
- [x] Make release/CI state handling fail closed (Minors 8 + 11). — Release
  inspection treats only an explicit GitHub HTTP 404 as absent and propagates
  transient/API failures. Feature-branch pushes no longer duplicate same-repo PR
  CI while main, pull-request, and reusable workflow coverage remain.
- [x] Verify staged binaries independently of GoReleaser labels. — npm staging
  validates PE, ELF, and Mach-O magic plus amd64/arm64 headers for all six targets
  before replacing staged output; cross-wired artifacts fail without destroying
  the prior stage.

## Scaffolding

- [x] Init Go module `github.com/shaunobi/tloc` (Go 1.25+), MIT LICENSE file, and the requested repository layout. — Added `main.go`, internal packages, calibration tools, fixtures, workflows, and npm packaging.
- [x] Add and verify gocodewalker, scc processor, tiktoken-go, and CLI dependencies on Windows. — All imports build and the full local suite passes with Go 1.25.12.
- [x] Confirm the required scc processor APIs and exported language detection. — Reused `ProcessConstants`, `FileJob`, `CountStats`, and `DetectLanguage`; no generator was needed.

## Core pipeline

- [x] Directory walker with ignore semantics and scc-style filters. — Implemented multiple inputs, independent ignore-disable flags, nested negation, compound extensions, special filenames, directory filters, symlink parity, and ignore/filter regression coverage.
- [x] Single-read file processing with size and binary skips plus scc metrics. — The same bounded byte slice feeds `CountStats` and tokenization; growth races cannot trigger unbounded reads.
- [x] Exact offline `o200k` tokenizer via tiktoken-go and `go:embed`. — Embedded the verified `o200k_base` ranks asset, with no runtime network access.
- [x] Worker pool sized to `GOMAXPROCS` with deterministic fan-in. — Added concurrent-use and forced completion-order tests; race mode itself is unavailable locally because no C compiler is installed.
- [x] Claude estimator infrastructure for two generations. — Current/legacy modes, compile-time factors, rounding, reports, and the `codex` alias are wired; readiness gating prevents neutral placeholders from being presented as calibrated.

## Aggregation and views

- [x] Per-language and per-file aggregation. — Reports retain complete metrics and deterministic total ordering for every sort key.
- [x] Recursive cumulative folder trees. — Added per-input roots, synthetic direct-file rows, collision-safe real `(root files)` directories, overlapping-input identity, and stable hierarchy.
- [x] Sorting by tokens, code, lines, files, or name across all views. — Numeric sorts descend, names ascend, and deterministic tie-breakers are fully tested.

## Output formats

- [x] Tabular renderer for language, file, and folder views. — Includes totals, token density, display-cell-aware width caps, folder indentation, and honest tokenizer footers.
- [x] JSON renderer with optional detail arrays, totals, and metadata. — Full metrics are emitted; folder rows also include input ID, depth, and synthetic identity so repeated inputs remain distinguishable.
- [x] CSV renderer for all three views. — Full metrics use flat slash-separated paths; cumulative folder rows include their machine identity and intentionally omit a misleading totals row.
- [x] Output-file and version flags. — Output files inside a scan are excluded for repeatability, and source/output path aliases are rejected to prevent destructive overwrites.

## Calibration tooling

- [x] Build the offline calibration workflow and representative corpus. — Added deterministic sampling, Anthropic `count_tokens` clients, framing-baseline subtraction, retries, spend confirmation, per-language/global analysis, JSON/Markdown reports, documentation, and a 16-language authored corpus; no source has been sent without explicit approval.
- [x] Run an approved, reproducible Claude calibration and retain the reports. — Expanded the authored corpus to five substantive files in each of 16 languages, excluded caches/generated reports, measured 80 files with `claude-opus-4-7`, `claude-opus-4-6`, and `claude-fable-5`, and recorded exact model IDs, hashes, global/fitted factors, global-factor error, and leave-one-out error in JSON and Markdown. The framing baseline uses the API-compatible `count_tokens("x") - 1` probe.
- [x] Calibrate and enable both Claude estimator generations. — Enabled global fallbacks `1.654377` (current) and `1.250226` (legacy); added only the data-justified current overrides C# `1.907865`, JSON `1.481303`, SQL `1.997848`, YAML `1.459644`, and legacy overrides C# `1.404494`, Markdown `1.119048`. Runtime counting uses exact canonical scc language labels with an immutable global fallback, while `codex` remains the exact `o200k` alias.
- [x] Verify the roughly 10% Claude accuracy target across common languages. — Production-factor MAPE is 5.85% overall for current Claude and 4.08% for legacy; the worst represented-language MAPE is 9.30% and 6.67%, respectively. Selected overrides improve leave-one-out MAPE to 1.67–10.52%, and Opus 4.7 and Fable 5 produced identical content counts on all 80 files. A calibration-contract test pins the factors, selection rule, coverage, spot-model equivalence, and per-language ceiling.

## Testing

- [x] Unit coverage for tokenizer vectors, calibration math, sorting, concurrency, filters, safety, and folder rollups. — Added regressions for all adversarial findings from three independent audits.
- [x] Golden tests for tabular, JSON, and CSV in all three views. — Nine checked-in goldens cover the complete renderer matrix.
- [x] Integration fixtures and scc ignore parity. — The counted file oracle matches pinned scc output; CLI output, stable in-tree reports, aliases, and independent ignore modes are covered.

## Docs

- [x] Standalone README with Go, release-binary, and npm installation plus all views and formats. — Documents filtering, machine folder identity, output safety, and the Claude accuracy/readiness caveat without competitor comparisons.
- [x] Complete `--help` text. — Covers every flag, writes successful help to stdout, and derives Claude availability from compile-time readiness.

## Release engineering

- [x] GitHub Actions CI. — Pinned action SHAs run formatting, vet, tests on Linux/macOS/Windows, npm tests, and a real GoReleaser snapshot-to-npm preflight; the workflow is reusable by releases.
- [x] GoReleaser configuration. — Builds macOS/Linux/Windows archives for amd64/arm64, embeds versions, publishes checksums, and creates resumable draft releases.
- [x] esbuild-style npm packaging. — Added the scoped wrapper and six private source templates, exact optional dependencies, no install-time downloads, safe staging, stable/prerelease dist-tags, platform-first publication, trusted-publisher support, artifact handoff, and 14 passing packaging tests.
- [x] Public repository and npm trusted-publishing bootstrap. — Authenticated GitHub and npm, made `shaunobi/tloc` public, pushed the completed implementation, published the wrapper and all six platform packages at `0.0.0-bootstrap.0`, and independently verified seven GitHub trusted-publisher configurations bound to `shaunobi/tloc` and `release.yml` with publish permission.
- [x] Homebrew tap. — Published `shaunobi/homebrew-tap` with a source-built `tloc` formula for v1.0.0, checksum-pinned the release source, documented `brew install shaunobi/tap/tloc`, and passed Homebrew test-bot on Intel macOS, Apple-silicon macOS, and Linux before merging the formula.
- [x] v1.0.0 public release and installation verification. — Tagged v1.0.0, completed the draft → trusted OIDC npm publish → finalized GitHub release workflow, published six checksum-listed platform archives and all seven npm packages at `latest=1.0.0`, and independently verified a checksum-matched downloaded binary, clean global npm and `npx` executions, and a fresh Go proxy/checksum-backed `go install`; every installed path reported `tloc 1.0.0` and completed a real scan.
- [x] Scoped npm naming decision. — Retained `@shaunobi/tloc` and closed the proposed unscoped-name dispute after reviewing npm's current policy and the existing functional `tloc` package; no support claim will be filed.
