# TODO — tloc v1

Work items to take tloc from spec to v1 release, per SPEC.md. Roughly ordered; items within a phase can interleave. When an item is finished, move it to DONE.md with a one-line note on what was done and any deviations from the spec.

## Scaffolding

- [ ] Init Go module `github.com/shaunobi/tloc` (Go 1.25+), MIT LICENSE file, basic repo layout (`main.go` or `cmd/`, internal packages, `tools/`, `testdata/`).
- [ ] Add dependencies: `github.com/boyter/gocodewalker`, `github.com/boyter/scc/v3/processor`, `github.com/pkoukk/tiktoken-go`; verify each imports and builds on Windows.
- [ ] Confirm scc's processor exposes what we need: `ProcessConstants()`, `FileJob`, `CountStats`, and language detection from file extension. If language detection isn't exported, write the build-time generator that replicates the extension map from languages.json (per SPEC Architecture).

## Core pipeline

- [ ] Directory walker: gocodewalker over one or more input paths with gitignore/.ignore/.sccignore semantics; `--no-ignore`/`--no-gitignore` disable handling; `--exclude-dir`, `--include-ext`, `--exclude-ext` filters mirroring scc.
- [ ] File processing: read each file once; skip binaries (walker/scc detection) and files over `--max-file-bytes` (default 1,000,000); build FileJob and run CountStats for lines/code/comments/blanks/complexity.
- [ ] Tokenizer interface `Count(content []byte) (int64, error)` with the o200k implementation via tiktoken-go; embed the o200k_base ranks file with go:embed (no runtime downloads).
- [ ] Claude estimator: o200k count × compile-time calibration factor; two generations (current default, `claude-legacy` for pre-Opus-4.7). Wire `--tokenizer claude|claude-legacy|o200k` with `codex` as alias for o200k.
- [ ] Worker pool sized to GOMAXPROCS, one file per job, channel fan-in to per-language aggregates; ensure deterministic output ordering regardless of completion order.

## Aggregation and views

- [ ] Per-language aggregation (default view) and per-file records (`--by-file`).
- [ ] Folder-tree aggregation (`--by-folder`): full recursive tree, one row per folder containing counted files, cumulative rollups, synthetic `(root files)` row, per-input-path subtrees, sibling ordering by `--sort` key. Mutually exclusive with `--by-file` (error if both passed).
- [ ] Sorting: `--sort tokens|code|lines|files|name` (default tokens desc; name asc), applied across all three views.

## Output formats

- [ ] Tabular renderer: Language/Files/Lines/Code/Tokens/Tok:Line columns, totals row, width-capped names, tokenizer footer with estimate caveat for claude modes; `--by-folder` variant with Folder column and two-space indentation.
- [ ] JSON renderer: one object with per-language array, optional per-file array, optional per-folder array (slash-separated paths, cumulative counts), totals, and metadata (version, tokenizer, calibration factor). Full record incl. comments/blanks/complexity/bytes.
- [ ] CSV renderer: summary by default; per-file rows under `--by-file`; per-folder rows under `--by-folder` with flat slash-separated path column.
- [ ] `-o/--output` file flag (default stdout) and `--version`.

## Calibration (needs Anthropic API key + small spend — may be human-blocked)

- [ ] Calibration script in `tools/`: sample representative source files across languages against the count_tokens API for both tokenizer generations; derive the compile-time factors; document re-derivation steps. Decide from the data whether per-language factors are justified or one factor per generation suffices.
- [ ] Verify claude estimates land within ~10% of count_tokens ground truth across common languages; record results.

## Testing

- [ ] Unit: tokenizer counts against fixed vectors with known o200k values; calibration math; sort behavior; folder-tree aggregation (cumulative rollups, sibling ordering, root files).
- [ ] Golden-file tests for tabular, JSON, and CSV output in all three views.
- [ ] Integration: fixtures tree with several languages; assert language, folder, and total rows; verify ignore-file handling matches scc output on the same tree.

## Docs

- [ ] README: install (go install, release binaries, npm), usage examples for all views and formats, claude-estimate accuracy caveat. Standalone tone — no comparisons to other counters.
- [ ] `--help` text covering all flags, with the estimate caveat on claude tokenizer modes.

## Release

- [ ] GitHub Actions CI: gofmt check, go vet, tests on linux/macos/windows.
- [ ] goreleaser config: darwin/linux/windows × amd64/arm64 archives.
- [ ] npm packaging: `@shaunobi/tloc` wrapper + per-platform binary packages via optionalDependencies (esbuild-style), bin name `tloc`, no postinstall download; publish wired into the goreleaser flow.
- [ ] Create public GitHub repo under shaunobi, push, tag v1.0.0, run release end-to-end; verify `go install`, a release binary, and `npx @shaunobi/tloc` all work.

## Post-v1 follow-ups (not release blockers)

- [ ] Homebrew tap (after first tagged release).
- [ ] File npm support dispute for the abandoned unscoped `tloc` name; if granted, move package to unscoped.
