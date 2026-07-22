# Claude tokenizer calibration

This tool derives the global fallback and selective per-language factors used by
tloc's `claude` and `claude-legacy` estimators. It sends representative source
text to Anthropic's `POST /v1/messages/count_tokens` endpoint and compares the
returned content count with the embedded `o200k_base` count.

The tool does not have default model IDs. Choose one model that uses the current
Claude tokenizer generation and one that uses the legacy generation, then record
the exact IDs in the generated report. This avoids silently reusing a stale
model choice when Anthropic's catalog changes.

## Privacy and cost

Source samples are sent to Anthropic. Use code you are authorized to disclose.
The tool requires `--confirm-spend` before it can call the API, never writes the
API key to a report, and runs sequentially with conservative sample limits. A
dry run makes no API calls.

## Re-derive the factors

1. Start with the checked-in authored corpus, which contains five varied files
   in each of 16 languages. Add other code trees when validating a workload not
   represented there. Avoid generated, vendored, or tiny files.
2. From the repository root, inspect the deterministic sample plan:

   ```powershell
   go run ./tools/calibrate --dry-run `
     --samples tools/calibrate/testdata/corpus `
     --max-per-language 5 `
     --max-samples 100
   ```

3. Check the selected languages, file count, byte cap, and total o200k tokens.
   Tune `--max-per-language`, `--max-samples`, and `--sample-bytes` if needed.
4. Set the API key only in the environment and run the calibration with exact
   model IDs:

   ```powershell
   $env:ANTHROPIC_API_KEY = "..."
   go run ./tools/calibrate `
     --current-model CURRENT_MODEL_ID `
     --legacy-model LEGACY_MODEL_ID `
     --spot-model fable5=SPOT_CHECK_MODEL_ID `
     --samples tools/calibrate/testdata/corpus `
     --samples C:\path\to\corpus-a `
     --samples C:\path\to\corpus-b `
     --max-per-language 5 `
     --max-samples 100 `
     --confirm-spend
   Remove-Item Env:ANTHROPIC_API_KEY
   ```

   `--spot-model label=model-id` is repeatable. Spot models appear in the report
   but do not map to a compile-time factor automatically.

5. Review `tools/calibrate/results/calibration.json` (machine-readable) and
   `calibration.md` (human-readable). The reported factor is:

   ```text
   sum(Claude content tokens) / sum(o200k_base tokens)
   ```

   Claude content tokens are the API's message count minus a separately measured
   framing baseline for the same model. Because the API rejects empty user
   messages, the tool measures `count_tokens("x") - 1`: the one-character probe
   contributes one content token and the remainder is message framing. The JSON
   records this method and also contains per-file ratios, fitted per-language
   factors, content hashes, and mean absolute percentage error (MAPE). Each
   language row separately shows error from the model-global factor and
   leave-one-out error from fitting that language's other samples, so a low
   fitted-factor MAPE cannot hide a poor global estimate.
6. Prefer one global factor per generation. Add a language override only when
   the global-factor language error is materially above the roughly 10% target,
   leave-one-out fitting improves it by several points, and the result is not
   driven by one tiny file. Languages without an override use the global factor.
7. Update `internal/tokenizer/factors.go`, set
   `ClaudeCurrentCalibrationReady` and `ClaudeLegacyCalibrationReady` to `true`,
   rerun tokenizer and calibration-contract tests, and commit the factor change
   together with the two generated result files and exact model IDs. The Claude
   CLI modes intentionally remain disabled while their readiness constants are
   false.

## Sampling behavior

The checked-in `testdata/corpus` is a balanced, authored cross-language baseline:
five substantive files in each of 16 languages. Combine it with larger
real-world trees when evaluating a specialized workload. The collector
recognizes common programming-language extensions and special filenames such as
`Dockerfile` and `Makefile`. It skips binary/non-UTF-8 files, lockfiles,
symlinks, caches, generated results, and common dependency directories. Files
are selected deterministically and round-robin across languages so a global cap
does not crowd out later-sorting languages. Only the configured prefix of each
file is sent; truncation is recorded in the result.

The token count endpoint may itself be an estimate, and server-side behavior can
change. Always retain the date, exact model IDs, per-file measurements, and
content hashes emitted by the tool so later runs can be compared meaningfully.
