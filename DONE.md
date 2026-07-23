# DONE — tloc v1

Completed items moved from `TODO.md`. Notes capture what was implemented and
any deliberate deviation from `SPEC.md`.

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
