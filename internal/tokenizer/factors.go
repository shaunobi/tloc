package tokenizer

// Claude calibration factors multiply an o200k_base token count and are
// rounded to the nearest token. Language overrides use exact canonical scc
// labels; all other languages use the corresponding global factor.
const (
	ClaudeCurrentCalibrationFactor = 1.654377
	ClaudeLegacyCalibrationFactor  = 1.250226
	ClaudeCurrentCalibrationReady  = true
	ClaudeLegacyCalibrationReady   = true
)

var claudeCurrentCalibrationOverrides = [...]CalibrationOverride{
	{Language: "C#", Factor: 1.907865},
	{Language: "JSON", Factor: 1.481303},
	{Language: "SQL", Factor: 1.997848},
	{Language: "YAML", Factor: 1.459644},
}

var claudeLegacyCalibrationOverrides = [...]CalibrationOverride{
	{Language: "C#", Factor: 1.404494},
	{Language: "Markdown", Factor: 1.119048},
}
