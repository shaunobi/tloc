# tloc

`tloc` counts lines of code and LLM tokens in one pass. It reports both metrics
side by side by language, file, or folder so you can see how much source code a
project contains and how much model context it consumes.

## Install

### Homebrew

```sh
brew install shaunobi/tap/tloc
```

### Go

Go 1.25.2 or newer is required:

```sh
go install github.com/shaunobi/tloc@latest
```

### Release binaries

Archives for macOS, Linux, and Windows on amd64 and arm64 are published on the
[GitHub releases page](https://github.com/shaunobi/tloc/releases). Download the
archive for your platform, extract it, and place `tloc` (or `tloc.exe`) on your
`PATH`.

Release checksums are available as `checksums.txt` alongside the archives.

### npm

The npm package installs the matching prebuilt binary as an optional platform
dependency. It does not download an executable during installation.

```sh
npm install --global @shaunobi/tloc
tloc .
```

You can also run it without a global install:

```sh
npx --yes @shaunobi/tloc .
```

## Usage

```text
tloc [flags] [paths...]
```

With no path, `tloc` scans the current directory. Multiple paths are combined
into one report. Inputs are independent: if they overlap, a file reachable
through more than one input is counted once for each input. Pass non-overlapping
paths when each physical file should contribute only once.

```sh
# Summarize the current project by language.
tloc

# Scan several inputs together.
tloc internal tools README.md

# Order languages by code lines instead of tokens.
tloc --sort code .
```

The default table contains `Language`, `Files`, `Lines`, `Code`, `Tokens`, and
`Tok/Line` columns, followed by a totals row. `Tok/Line` is tokens divided by
code lines.

### File view

Use `--by-file` to report one row per file:

```sh
tloc --by-file .
tloc --by-file --sort name internal
```

### Folder view

Use `--by-folder` for a cumulative directory tree. A folder includes every
counted file beneath it; files directly in an input root appear under
`(root files)`. A file passed directly as an input produces one depth-zero
synthetic `(root files)` bucket, rather than pretending the file is a folder.

```sh
tloc --by-folder .
tloc --by-folder --sort code internal tools
```

`--by-file` and `--by-folder` are mutually exclusive.

### JSON and CSV

```sh
tloc --format json .
tloc --format json --by-file --output report.json .
tloc --format csv --by-folder -o folders.csv .
# Existing output files require explicit replacement.
tloc --format csv --by-folder -o folders.csv --force .
```

JSON contains language records, the selected optional view, totals, and
metadata. JSON and CSV include comments, blanks, complexity, byte counts, and
token counts in addition to the columns shown in the table. Folder records also
include the input ID, depth, and synthetic-root marker so overlapping or
repeated input paths remain distinguishable in machine-readable output. Treat
`(input_id, folder, synthetic)` as the folder identity; `folder` alone can
collide with the synthetic `(root files)` row. Folder CSV rows are cumulative,
contain no separate totals row, and must not be summed together: a parent's
metrics already include its descendants.

JSON metadata always includes `complete` and a `skipped` array. CSV appends
`record_type`, `complete`, and `skipped_*` columns; recoverable scan failures
appear as `skipped` rows. If a file or directory becomes unreadable during a
scan, tloc still renders the readable portion, warns on standard error, and
exits nonzero. Check the structured completeness fields before consuming a
JSON or CSV report.

### Tokenizers

The default tokenizer is `o200k`:

```sh
tloc --tokenizer o200k .
tloc --tokenizer codex .
tloc --tokenizer claude .
tloc --tokenizer claude-legacy .
```

`codex` is an alias for `o200k`. The o200k count is exact for that encoding and
works entirely offline. `claude` estimates the current Claude tokenizer from
the o200k count; `claude-legacy` targets models before the current tokenizer
generation. Anthropic does not publish an exact local tokenizer for these
models, so Claude results are estimates and may differ from the `count_tokens`
API. Both modes use an offline global fallback plus only the language overrides
justified by leave-one-out validation on a balanced 80-file corpus. The exact
models, factors, per-language errors, and content hashes are retained in the
[calibration report](tools/calibrate/results/calibration.md). On that corpus,
the production factors measured 5.85% overall MAPE for current Claude and 4.08%
for legacy, with every represented language below 10% in-sample MAPE. Those
figures describe the represented fitting corpus, not every programming
language. Leave-one-out language error reached 10.52%. On a separate 20-file
holdout covering C, HTML, Kotlin, and Swift, production-factor MAPE was 8.25%
overall for current Claude and 4.03% for legacy. Current-Claude HTML was the
exception at 17.27%; other held-out language/generation combinations were below
10%. Languages absent from both sets use the global factor without direct
validation.

### Filtering and ignore files

By default, repository ignore files are honored. The scanning controls are:

| Flag | Purpose |
| --- | --- |
| `--include-ext` | Count only the listed file extensions |
| `--exclude-ext` | Exclude the listed file extensions |
| `--exclude-dir` | Exclude directory names |
| `--no-ignore` | Disable `.ignore` and `.sccignore` handling |
| `--no-gitignore` | Disable `.gitignore` handling |
| `--max-file-bytes` | Skip larger files (default `1000000`) |

Extension and directory lists use the same comma-separated form as the CLI
help, for example:

```sh
tloc --include-ext go,ts --exclude-dir vendor,node_modules .
```

Binary files and files larger than `--max-file-bytes` are skipped.

Before scanning, tloc checks that an output destination is writable without
truncating it. Existing output paths are refused by default; pass `--force` to
replace one deliberately. When an output points inside a scanned directory,
that exact file is excluded from the scan. Source/output aliases are rejected
even with `--force`.

### Other flags

| Flag | Values | Default |
| --- | --- | --- |
| `-f`, `--format` | `tabular`, `json`, `csv` | `tabular` |
| `--sort` | `tokens`, `code`, `lines`, `files`, `name` | `tokens` |
| `-o`, `--output` | output file path | standard output |
| `--force` | replace an existing output file | off |
| `--version` | print the tloc version | |
| `-h`, `--help` | show complete CLI help | |

Numeric sorts are descending. Name sort is ascending. Tabular output ends with
the tokenizer used and explicitly labels Claude counts as estimates.

## License

MIT
