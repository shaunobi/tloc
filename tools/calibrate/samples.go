package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

type sourceSample struct {
	Path       string
	Language   string
	Content    []byte
	Bytes      int
	Truncated  bool
	ContentSHA string
	O200K      int64
	selectKey  string
}

var extensionLanguages = map[string]string{
	".c":     "C",
	".cc":    "C++",
	".cpp":   "C++",
	".cs":    "C#",
	".css":   "CSS",
	".dart":  "Dart",
	".ex":    "Elixir",
	".exs":   "Elixir",
	".go":    "Go",
	".h":     "C",
	".hpp":   "C++",
	".html":  "HTML",
	".java":  "Java",
	".js":    "JavaScript",
	".json":  "JSON",
	".jsx":   "JavaScript",
	".kt":    "Kotlin",
	".kts":   "Kotlin",
	".lua":   "Lua",
	".md":    "Markdown",
	".php":   "PHP",
	".py":    "Python",
	".r":     "R",
	".rb":    "Ruby",
	".rs":    "Rust",
	".scala": "Scala",
	".sh":    "Shell",
	".sql":   "SQL",
	".swift": "Swift",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".toml":  "TOML",
	".vue":   "Vue",
	".yaml":  "YAML",
	".yml":   "YAML",
	".zig":   "Zig",
}

var filenameLanguages = map[string]string{
	"dockerfile": "Dockerfile",
	"gemfile":    "Ruby",
	"makefile":   "Makefile",
	"rakefile":   "Ruby",
}

var skippedDirectories = map[string]struct{}{
	".cache":       {},
	".git":         {},
	".hg":          {},
	".svn":         {},
	".venv":        {},
	"build":        {},
	"dist":         {},
	"node_modules": {},
	"results":      {},
	"target":       {},
	"vendor":       {},
	"venv":         {},
}

var skippedFilenames = map[string]struct{}{
	"cargo.lock":        {},
	"package-lock.json": {},
	"pnpm-lock.yaml":    {},
	"yarn.lock":         {},
}

func collectSamples(inputs []string, maxBytes, maxPerLanguage, maxSamples int) ([]sourceSample, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one sample file or directory is required")
	}
	if maxBytes < 1 || maxPerLanguage < 1 || maxSamples < 1 {
		return nil, fmt.Errorf("sample limits must all be positive")
	}

	seen := make(map[string]struct{})
	var candidates []sourceSample
	for _, input := range inputs {
		cleaned := filepath.Clean(input)
		info, err := os.Stat(cleaned)
		if err != nil {
			return nil, fmt.Errorf("stat sample input %q: %w", input, err)
		}
		if info.IsDir() {
			err = filepath.WalkDir(cleaned, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					if path != cleaned {
						if _, skip := skippedDirectories[strings.ToLower(entry.Name())]; skip {
							return filepath.SkipDir
						}
					}
					return nil
				}
				if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
					return nil
				}
				relative, relErr := filepath.Rel(cleaned, path)
				if relErr != nil {
					return relErr
				}
				display := filepath.ToSlash(filepath.Join(filepath.Base(cleaned), relative))
				sample, ok, sampleErr := loadSample(path, display, maxBytes)
				if sampleErr != nil {
					return sampleErr
				}
				if ok {
					canonical, canonicalErr := filepath.Abs(path)
					if canonicalErr != nil {
						return canonicalErr
					}
					canonical = strings.ToLower(filepath.Clean(canonical))
					if _, duplicate := seen[canonical]; !duplicate {
						seen[canonical] = struct{}{}
						candidates = append(candidates, sample)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk sample input %q: %w", input, err)
			}
			continue
		}

		sample, ok, err := loadSample(cleaned, filepath.ToSlash(filepath.Base(cleaned)), maxBytes)
		if err != nil {
			return nil, err
		}
		if ok {
			canonical, err := filepath.Abs(cleaned)
			if err != nil {
				return nil, err
			}
			canonical = strings.ToLower(filepath.Clean(canonical))
			if _, duplicate := seen[canonical]; !duplicate {
				seen[canonical] = struct{}{}
				candidates = append(candidates, sample)
			}
		}
	}

	byLanguage := make(map[string][]sourceSample)
	for _, sample := range candidates {
		byLanguage[sample.Language] = append(byLanguage[sample.Language], sample)
	}
	languages := make([]string, 0, len(byLanguage))
	for language := range byLanguage {
		languages = append(languages, language)
		sort.Slice(byLanguage[language], func(i, j int) bool {
			return byLanguage[language][i].selectKey < byLanguage[language][j].selectKey
		})
		if len(byLanguage[language]) > maxPerLanguage {
			byLanguage[language] = byLanguage[language][:maxPerLanguage]
		}
	}
	sort.Strings(languages)

	// Select round-robin across languages so the global cap cannot be consumed
	// by whichever language sorts first.
	selected := make([]sourceSample, 0, min(maxSamples, len(candidates)))
	for round := 0; len(selected) < maxSamples; round++ {
		added := false
		for _, language := range languages {
			if round < len(byLanguage[language]) {
				selected = append(selected, byLanguage[language][round])
				added = true
				if len(selected) == maxSamples {
					break
				}
			}
		}
		if !added {
			break
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Language != selected[j].Language {
			return selected[i].Language < selected[j].Language
		}
		return selected[i].Path < selected[j].Path
	})
	return selected, nil
}

func loadSample(path, displayPath string, maxBytes int) (sourceSample, bool, error) {
	language, ok := detectSampleLanguage(path)
	if !ok {
		return sourceSample{}, false, nil
	}
	if _, skip := skippedFilenames[strings.ToLower(filepath.Base(path))]; skip {
		return sourceSample{}, false, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return sourceSample{}, false, fmt.Errorf("open sample %q: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes+utf8.UTFMax+1)))
	if err != nil {
		return sourceSample{}, false, fmt.Errorf("read sample %q: %w", path, err)
	}
	if len(data) == 0 || bytes.IndexByte(data, 0) >= 0 {
		return sourceSample{}, false, nil
	}

	truncated := len(data) > maxBytes
	if truncated {
		data = data[:maxBytes]
		valid := false
		for trim := 0; trim < utf8.UTFMax && trim <= len(data); trim++ {
			candidate := data[:len(data)-trim]
			if utf8.Valid(candidate) {
				data = candidate
				valid = true
				break
			}
		}
		if !valid {
			return sourceSample{}, false, nil
		}
	} else if !utf8.Valid(data) {
		return sourceSample{}, false, nil
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return sourceSample{}, false, nil
	}

	digest := sha256.Sum256(data)
	selectionDigest := sha256.Sum256([]byte(filepath.ToSlash(strings.ToLower(displayPath))))
	return sourceSample{
		Path:       displayPath,
		Language:   language,
		Content:    data,
		Bytes:      len(data),
		Truncated:  truncated,
		ContentSHA: fmt.Sprintf("%x", digest),
		selectKey:  fmt.Sprintf("%x", selectionDigest),
	}, true, nil
}

func detectSampleLanguage(path string) (string, bool) {
	name := strings.ToLower(filepath.Base(path))
	if language, ok := filenameLanguages[name]; ok {
		return language, true
	}
	if strings.HasSuffix(name, ".d.ts") {
		return "TypeScript", true
	}
	language, ok := extensionLanguages[strings.ToLower(filepath.Ext(name))]
	return language, ok
}
