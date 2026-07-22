// Package analyze discovers and processes source files with a bounded worker
// pool. Each accepted file is read once and the same bytes are passed to scc
// and the selected tokenizer.
package analyze

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/boyter/gocodewalker"
	"github.com/boyter/scc/v3/processor"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

const DefaultMaxFileBytes int64 = 1_000_000

type InputKind string

const (
	InputDirectory InputKind = "directory"
	InputFile      InputKind = "file"
)

type InputRoot struct {
	ID    int
	Given string
	Abs   string
	Kind  InputKind
}

type Metrics struct {
	Files      int64
	Lines      int64
	Code       int64
	Comments   int64
	Blanks     int64
	Complexity int64
	Bytes      int64
	Tokens     int64
}

type FileRecord struct {
	InputID  int
	Path     string
	RelPath  string
	Language string
	Metrics  Metrics
}

type Options struct {
	IncludeExt   []string
	ExcludeExt   []string
	ExcludeDir   []string
	ExcludeFiles []string
	MaxFileBytes int64
	NoIgnore     bool
	NoGitignore  bool
	Workers      int
}

type fileTask struct {
	input           InputRoot
	location        string
	filename        string
	relPath         string
	display         string
	allowListActive bool
}

type processResult struct {
	record *FileRecord
	err    error
}

type fileExclusion struct {
	path string
	info os.FileInfo
}

type exclusionMatch uint8

const (
	exclusionNone exclusionMatch = iota
	exclusionExact
	exclusionAlias
)

var processConstantsOnce sync.Once

// Run discovers and analyzes paths. An empty path list scans the current
// directory. Returned records are deterministic even though file processing is
// concurrent.
func Run(paths []string, counter tokenizer.Counter, options Options) ([]InputRoot, []FileRecord, error) {
	return runWithReader(paths, counter, options, readFileLimited)
}

func runWithReader(paths []string, counter tokenizer.Counter, options Options, readFile func(string, int64) ([]byte, error)) ([]InputRoot, []FileRecord, error) {
	if counter == nil {
		return nil, nil, errors.New("token counter is nil")
	}
	if options.MaxFileBytes == 0 {
		options.MaxFileBytes = DefaultMaxFileBytes
	}
	if options.MaxFileBytes < 0 {
		return nil, nil, errors.New("max file bytes must be greater than zero")
	}
	if options.Workers == 0 {
		options.Workers = runtime.GOMAXPROCS(0)
	}
	if options.Workers < 1 {
		return nil, nil, errors.New("worker count must be greater than zero")
	}
	if readFile == nil {
		return nil, nil, errors.New("file reader is nil")
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}

	processConstantsOnce.Do(processor.ProcessConstants)
	inputs, err := resolveInputs(paths)
	if err != nil {
		return nil, nil, err
	}

	var tasks []fileTask
	var discoverErrors []error
	for _, input := range inputs {
		found, discoverErr := discover(input, options)
		tasks = append(tasks, found...)
		if discoverErr != nil {
			discoverErrors = append(discoverErrors, discoverErr)
		}
	}

	jobs := make(chan fileTask)
	results := make(chan processResult, options.Workers)
	var workers sync.WaitGroup
	for range options.Workers {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for task := range jobs {
				results <- process(task, counter, options.MaxFileBytes, readFile)
			}
		}()
	}
	go func() {
		for _, task := range tasks {
			jobs <- task
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()

	records := make([]FileRecord, 0, len(tasks))
	processErrors := make([]error, 0)
	for result := range results {
		if result.err != nil {
			processErrors = append(processErrors, result.err)
			continue
		}
		if result.record != nil {
			records = append(records, *result.record)
		}
	}
	slices.SortFunc(records, func(a, b FileRecord) int {
		if n := strings.Compare(a.Path, b.Path); n != 0 {
			return n
		}
		if n := strings.Compare(a.Language, b.Language); n != 0 {
			return n
		}
		return a.InputID - b.InputID
	})

	allErrors := append(discoverErrors, processErrors...)
	if len(allErrors) > 0 {
		slices.SortFunc(allErrors, func(a, b error) int { return strings.Compare(a.Error(), b.Error()) })
		return inputs, records, errors.Join(allErrors...)
	}
	return inputs, records, nil
}

func resolveInputs(paths []string) ([]InputRoot, error) {
	inputs := make([]InputRoot, 0, len(paths))
	for id, given := range paths {
		if strings.TrimSpace(given) == "" {
			return nil, errors.New("input path is empty")
		}
		abs, err := filepath.Abs(given)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", given, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("inspect %q: %w", given, err)
		}
		kind := InputFile
		if info.IsDir() {
			kind = InputDirectory
		} else if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("input %q is not a regular file or directory", given)
		}
		inputs = append(inputs, InputRoot{
			ID:    id,
			Given: filepath.ToSlash(given),
			Abs:   filepath.Clean(abs),
			Kind:  kind,
		})
	}
	return inputs, nil
}

func discover(input InputRoot, options Options) ([]fileTask, error) {
	excludedFiles := prepareFileExclusions(options.ExcludeFiles)
	if input.Kind == InputFile {
		if match := excludedFiles.match(input.Abs); match != exclusionNone {
			return nil, fmt.Errorf("input %q is the output file or an alias of it", input.Given)
		}
		if !extensionAllowed(filepath.Base(input.Abs), options.IncludeExt, options.ExcludeExt) {
			return nil, nil
		}
		return []fileTask{{
			input:           input,
			location:        input.Abs,
			filename:        filepath.Base(input.Abs),
			relPath:         filepath.ToSlash(filepath.Base(input.Abs)),
			display:         input.Given,
			allowListActive: len(options.IncludeExt) > 0,
		}}, nil
	}

	queue := make(chan *gocodewalker.File, max(32, runtime.GOMAXPROCS(0)*2))
	walker := gocodewalker.NewFileWalker(input.Abs, queue)
	walker.IncludeHidden = true
	walker.IgnoreIgnoreFile = options.NoIgnore
	walker.IgnoreGitIgnore = options.NoGitignore
	walker.ExcludeDirectory = uniqueStrings(append([]string{".git", ".hg", ".svn"}, options.ExcludeDir...))
	walker.ExcludeFilename = []string{".gitignore", ".ignore", ".sccignore"}
	walker.SetConcurrency(runtime.GOMAXPROCS(0))
	if !options.NoIgnore {
		walker.CustomIgnore = []string{".sccignore"}
	}
	walker.SetErrorHandler(func(error) bool { return false })

	errQueue := make(chan error, 1)
	go func() { errQueue <- walker.Start() }()

	tasks := make([]fileTask, 0)
	var aliasError error
	for file := range queue {
		location := filepath.Clean(filepath.FromSlash(file.Location))
		switch excludedFiles.match(location) {
		case exclusionExact:
			continue
		case exclusionAlias:
			if aliasError == nil {
				aliasError = fmt.Errorf("scanned file %q aliases the output file", location)
			}
			continue
		}
		// Apply extension filters after walking. gocodewalker's exclude-extension
		// pass can overwrite earlier ignore decisions, causing ignored files to be
		// emitted when an exclude list is present.
		if !extensionAllowed(file.Filename, options.IncludeExt, options.ExcludeExt) {
			continue
		}
		rel, err := filepath.Rel(input.Abs, location)
		if err != nil {
			return nil, fmt.Errorf("make %q relative to %q: %w", location, input.Given, err)
		}
		rel = filepath.ToSlash(rel)
		tasks = append(tasks, fileTask{
			input:           input,
			location:        location,
			filename:        file.Filename,
			relPath:         rel,
			display:         joinDisplayPath(input.Given, rel),
			allowListActive: len(options.IncludeExt) > 0,
		})
	}
	if err := <-errQueue; err != nil {
		return tasks, fmt.Errorf("walk %q: %w", input.Given, err)
	}
	if aliasError != nil {
		return tasks, aliasError
	}
	return tasks, nil
}

func process(task fileTask, counter tokenizer.Counter, maxFileBytes int64, readFile func(string, int64) ([]byte, error)) processResult {
	info, err := os.Lstat(task.location)
	if err != nil {
		return processResult{err: fmt.Errorf("inspect %q: %w", task.display, err)}
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileBytes {
		return processResult{}
	}

	possibleLanguages, fallbackLanguage := detectLanguage(task.filename, task.allowListActive)
	if len(possibleLanguages) == 0 {
		return processResult{}
	}
	for _, possibleLanguage := range possibleLanguages {
		if strings.EqualFold(possibleLanguage, "ignore") || strings.EqualFold(possibleLanguage, "gitignore") {
			return processResult{}
		}
	}
	content, err := readFile(task.location, maxFileBytes)
	if err != nil {
		return processResult{err: fmt.Errorf("read %q: %w", task.display, err)}
	}
	if int64(len(content)) > maxFileBytes {
		return processResult{}
	}

	language := processor.DetermineLanguage(task.filename, fallbackLanguage, possibleLanguages, content)
	if language == processor.SheBang {
		cutoff := min(200, len(content))
		language, err = processor.DetectSheBang(string(content[:cutoff]))
		if err != nil {
			return processResult{}
		}
	}
	if language == "" {
		return processResult{}
	}

	job := &processor.FileJob{
		Filename: task.filename,
		Location: task.location,
		Language: language,
		Content:  content,
		Bytes:    int64(len(content)),
	}
	processor.CountStats(job)
	if job.Binary {
		return processResult{}
	}
	tokens, err := tokenizer.CountForLanguage(counter, content, job.Language)
	if err != nil {
		return processResult{err: fmt.Errorf("tokenize %q: %w", task.display, err)}
	}
	record := &FileRecord{
		InputID:  task.input.ID,
		Path:     filepath.ToSlash(task.display),
		RelPath:  task.relPath,
		Language: job.Language,
		Metrics: Metrics{
			Files:      1,
			Lines:      job.Lines,
			Code:       job.Code,
			Comments:   job.Comment,
			Blanks:     job.Blank,
			Complexity: job.Complexity,
			Bytes:      job.Bytes,
			Tokens:     tokens,
		},
	}
	return processResult{record: record}
}

// readFileLimited reads at most maxFileBytes+1 bytes. The extra byte lets the
// caller distinguish an exact-cap file from one that grew after the stat
// check without allowing an unbounded allocation.
func readFileLimited(filename string, maxFileBytes int64) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	readLimit := maxFileBytes
	if readLimit < 1<<63-1 {
		readLimit++
	}
	return io.ReadAll(io.LimitReader(file, readLimit))
}

func extensionAllowed(filename string, includes, excludes []string) bool {
	extension, recognized := extensionForFilter(filename, len(includes) > 0)
	if slices.Contains(excludes, extension) {
		return false
	}
	if len(includes) == 0 {
		return true
	}
	return recognized && slices.Contains(includes, extension)
}

// extensionForFilter mirrors the extension identity scc uses for its allow
// and deny lists. With an allow list, scc deliberately bypasses special full
// filename mappings (for example xmake.lua becomes lua); without one it uses
// the normal language detector.
func extensionForFilter(filename string, allowListActive bool) (string, bool) {
	languages, extension := detectLanguage(filename, allowListActive)
	return normalizeExtension(extension), len(languages) > 0
}

func detectLanguage(filename string, allowListActive bool) ([]string, string) {
	if !allowListActive {
		return processor.DetectLanguage(filename)
	}

	name := strings.ToLower(filename)
	extension := ""
	languages, ok := processor.ExtensionToLanguage[name]
	if !ok {
		extension = sccCompoundExtension(name)
		languages, ok = processor.ExtensionToLanguage[extension]
	}
	if !ok {
		extension = sccCompoundExtension(extension)
		languages = processor.ExtensionToLanguage[extension]
	}
	return languages, extension
}

func sccCompoundExtension(name string) string {
	name = strings.ToLower(name)
	extension := filepath.Ext(name)
	if extension == "" || strings.LastIndex(name, ".") == 0 {
		return name
	}
	subExtension := filepath.Ext(strings.TrimSuffix(name, extension))
	return strings.TrimPrefix(subExtension+extension, ".")
}

func normalizeExtension(extension string) string {
	return strings.ToLower(strings.TrimPrefix(extension, "."))
}

func prepareFileExclusions(paths []string) fileExclusions {
	exclusions := make(fileExclusions, 0, len(paths))
	for _, filename := range paths {
		filename = filepath.Clean(filename)
		info, _ := os.Stat(filename)
		exclusions = append(exclusions, fileExclusion{path: filename, info: info})
	}
	return exclusions
}

type fileExclusions []fileExclusion

func (excluded fileExclusions) match(filename string) exclusionMatch {
	filename = filepath.Clean(filename)
	var info os.FileInfo
	for _, candidate := range excluded {
		if filename == candidate.path || (runtime.GOOS == "windows" && strings.EqualFold(filename, candidate.path)) {
			return exclusionExact
		}
		if candidate.info == nil {
			continue
		}
		if info == nil {
			info, _ = os.Stat(filename)
		}
		if info != nil && os.SameFile(info, candidate.info) {
			return exclusionAlias
		}
	}
	return exclusionNone
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func joinDisplayPath(given, rel string) string {
	if given == "." {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(filepath.Join(filepath.FromSlash(given), filepath.FromSlash(rel)))
}
