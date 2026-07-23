# tloc — SPEC

Count lines of code and LLM tokens for a codebase in one pass, reported side by side per language and per file. Standalone Go CLI in the spirit of scc; not a fork of it.

## Why

LOC counters (scc, tokei, cloc) measure size but not what a codebase costs in context window. tloc reports both metric families from a single read of each file: classic line counts plus token counts, with a tokens-per-line density column that shows which parts of a codebase actually eat context.

## Architecture

Single binary, Go 1.25+. Directory walking uses github.com/boyter/gocodewalker, which provides the same gitignore, .ignore, and .sccignore semantics as scc. Line counting uses github.com/boyter/scc/v3/processor as a library: call `processor.ProcessConstants()` once at startup, then per file build a `FileJob{Filename, Language, Content, Bytes}` and call `processor.CountStats(job)` to obtain lines, code, comments, blanks, and complexity. Each file is read exactly once; the same in-memory bytes feed CountStats and the tokenizer. Language detection reuses scc's languages.json mapping via the processor package; if the needed detection function is not exported, replicate the extension map at build time from languages.json.

Concurrency: worker pool sized to GOMAXPROCS, one file per job. Tokenization dominates runtime (BPE is far slower than scc's byte state machine), so parallelism per file matters. Aggregate per-language totals via channel fan-in. Skip binary files using the walker/scc detection and skip files over a size cap (default 1,000,000 bytes, configurable).

## Tokenizers

Interface: `Count(content []byte) (int64, error)`. Two implementations.

**o200k (default)** — exact BPE counting via github.com/pkoukk/tiktoken-go using the o200k_base encoding used by current OpenAI/codex models. Embed the BPE ranks file with go:embed so the binary never downloads anything at runtime.

**claude** — Anthropic ships no local tokenizer for Claude 3+; the only exact source is their count_tokens API, which v1 deliberately does not call. Estimate instead: run the o200k count and multiply by a calibration factor. Two tokenizer generations matter: models before Opus 4.7, and Opus 4.7 / Fable 5 onward, whose tokenizer produces roughly 30% more tokens for the same text. Default to the current generation. Factors are compile-time constants derived offline by sampling representative source files across languages against the count_tokens API; keep the derivation script in tools/ so factors can be re-derived as models change. Use per-language factors only if sampling shows the variance justifies it; otherwise one factor per generation.

Selection: `--tokenizer claude|claude-legacy|o200k` (default o200k), with `codex` accepted as an alias for o200k.

## CLI

`tloc [flags] [paths...]`, default path is the current directory. Multiple paths aggregate.

Flags: `--tokenizer` as above, `--by-file` for per-file rows, `--by-folder` for per-folder rows (mutually exclusive with `--by-file`; passing both is an error), `--format tabular|json|csv` (default tabular), `-o/--output` file (default stdout), `--force` to replace an existing output file, `--sort tokens|code|lines|files|name` (default tokens, descending; name ascending), `--include-ext`/`--exclude-ext`/`--exclude-dir` mirroring scc semantics, `--max-file-bytes` (default 1000000), `--no-ignore`/`--no-gitignore` to disable ignore-file handling, `--version`.

`--by-folder` reports the full directory tree: one row per folder that contains at least one counted file (directly or transitively), with counts cumulative — a folder's row includes everything beneath it. Rows are folders only; language detail is not broken out in this view (use the default view for that). Files sitting directly in a scanned root appear under a synthetic `(root files)` row. A direct file input is represented once as a depth-zero synthetic `(root files)` bucket. Sibling folders are ordered by the `--sort` key; children always appear under their parent. With multiple input paths, each path forms its own top-level subtree keyed by the path as given.

Existing output destinations are never overwritten unless `--force` is present. Output writability is checked before scanning without truncating the destination; source/output aliases remain errors even under `--force`.

## Output

Tabular: columns Language, Files, Lines, Code, Tokens, Tok/Line, plus a totals row. Tok/Line is tokens divided by code lines, one decimal place. Language names width-capped like scc. A one-line footer names the tokenizer used and, for claude modes, notes that counts are estimates.

Tabular with --by-folder: first column is Folder instead of Language; nesting is shown by indenting each level two spaces under its parent. Other columns and the totals row are unchanged.

JSON and CSV carry the full record: language, file path (when --by-file), files, lines, code, comments, blanks, complexity, bytes, tokens. JSON is one object containing a per-language array, an optional per-file array, an optional per-folder array (when --by-folder: slash-separated relative path, cumulative counts), totals, and metadata (tool version, tokenizer name, calibration factor, completeness, and skipped entries). CSV emits a summary by default, per-file rows under --by-file, and per-folder rows under --by-folder with the folder path as a flat slash-separated column — indentation exists only in the tabular view. CSV status columns distinguish data from skipped-entry rows and carry report completeness. Complexity and comment/blank detail appear only in JSON/CSV, not the tabular view.

Recoverable per-file and directory traversal errors do not suppress readable results: render the partial report, warn on stderr for every skipped entry, and exit nonzero. JSON and CSV must mark such a report incomplete so machine consumers cannot accept it silently.

## Accuracy

o200k counts are exact for that encoding. claude counts are estimates: target roughly 10% error against count_tokens ground truth for validated workloads. The calibration report must identify represented and held-out languages, disclose exceptions above the target, and make clear that unrepresented languages use an unvalidated global fallback. State this plainly in README and --help.

## Packaging

Public OSS, MIT license. Module path github.com/shaunobi/tloc. README covers install (go install, release binaries, npm), usage examples, and the accuracy caveat for Claude estimates; the tool stands alone — no comparisons to other counters. Release with goreleaser for darwin/linux/windows on amd64/arm64. npm distribution: publish `@shaunobi/tloc` wrapping the release binaries, esbuild-style — per-platform packages carrying the prebuilt binary, selected via optionalDependencies, so `npm install -g @shaunobi/tloc` and `npx @shaunobi/tloc` work without a postinstall download; the bin name is `tloc` regardless. Publishing is wired into the goreleaser release flow. The scoped package name is the permanent choice; no claim or dispute over the unrelated unscoped `tloc` package is planned. Homebrew tap after the first tagged release. CI via GitHub Actions: gofmt check, go vet, tests on all three OSes.

## Testing

Unit: tokenizer counts against fixed vectors with known o200k values, calibration math, sort behavior, folder-tree aggregation (cumulative rollups, sibling ordering, root files), golden-file tests for all three output formats. Integration: run against a fixtures tree containing several languages and assert language, folder, and total rows; verify ignore-file handling matches scc output on the same tree.

## Non-goals for v1

No count_tokens API mode (the --exact flag is deferred), no COCOMO/LOCOMO cost modeling, no git history reports, no HTML output, no ULOC/DRYness metrics.
