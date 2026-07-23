package aggregate

import (
	"reflect"
	"testing"

	"github.com/shaunobi/tloc/internal/model"
)

func TestBuildLanguageAndFileViews(t *testing.T) {
	inputs, files := fixture()
	metadata := model.Metadata{Version: "test", Tokenizer: "o200k", CalibrationFactor: 1}

	languageReport, err := Build(inputs, files, model.ViewLanguage, model.SortTokens, metadata)
	if err != nil {
		t.Fatal(err)
	}
	wantLanguages := []string{"Go", "JavaScript", "Python", "Markdown"}
	if got := languageNames(languageReport.Languages); !reflect.DeepEqual(got, wantLanguages) {
		t.Fatalf("languages = %v, want %v", got, wantLanguages)
	}
	if languageReport.Files != nil || languageReport.Folders != nil {
		t.Fatalf("language view unexpectedly contains detail rows: %+v", languageReport)
	}
	wantTotals := model.Metrics{Files: 7, Lines: 68, Code: 50, Comments: 11, Blanks: 7, Complexity: 14, Bytes: 688, Tokens: 142}
	if languageReport.Totals != wantTotals {
		t.Fatalf("totals = %+v, want %+v", languageReport.Totals, wantTotals)
	}
	if languageReport.Languages[0].Metrics.Files != 2 || languageReport.Languages[0].Metrics.Tokens != 61 {
		t.Fatalf("Go metrics = %+v", languageReport.Languages[0].Metrics)
	}

	fileReport, err := Build(inputs, files, model.ViewFile, model.SortTokens, metadata)
	if err != nil {
		t.Fatal(err)
	}
	wantPaths := []string{
		"internal/a.go",
		"other/root/src/app.js",
		"internal/sub/b.py",
		"main.go",
		"docs/x.md",
		"README.md",
		"other/root/root.py",
	}
	if got := filePaths(fileReport.Files); !reflect.DeepEqual(got, wantPaths) {
		t.Fatalf("files = %v, want %v", got, wantPaths)
	}
	for _, file := range fileReport.Files {
		if file.Metrics.Files != 1 {
			t.Fatalf("%s Files = %d, want 1", file.Path, file.Metrics.Files)
		}
		if len(file.Path) > 0 && containsBackslash(file.Path) {
			t.Fatalf("path was not normalized: %q", file.Path)
		}
	}
	if len(fileReport.Languages) != 4 {
		t.Fatalf("file view lost language summary: %+v", fileReport.Languages)
	}
}

func TestBuildFolderTreeCumulativeAndSorted(t *testing.T) {
	inputs, files := fixture()
	report, err := Build(inputs, files, model.ViewFolder, model.SortTokens, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}

	wantPaths := []string{
		".",
		"internal",
		"internal/sub",
		"(root files)",
		"docs",
		"other/root",
		"other/root/src",
		"other/root/(root files)",
	}
	if got := folderPaths(report.Folders); !reflect.DeepEqual(got, wantPaths) {
		t.Fatalf("folder preorder = %v, want %v", got, wantPaths)
	}

	assertFolder := func(path string, depth int, synthetic bool, metrics model.Metrics) {
		t.Helper()
		for _, folder := range report.Folders {
			if folder.Path == path {
				if folder.Depth != depth || folder.Synthetic != synthetic || folder.Metrics != metrics {
					t.Fatalf("folder %q = %+v, want depth=%d synthetic=%v metrics=%+v", path, folder, depth, synthetic, metrics)
				}
				return
			}
		}
		t.Fatalf("folder %q not found", path)
	}
	assertFolder(".", 0, false, model.Metrics{Files: 5, Lines: 51, Code: 36, Comments: 9, Blanks: 6, Complexity: 9, Bytes: 518, Tokens: 102})
	assertFolder("internal", 1, false, model.Metrics{Files: 2, Lines: 32, Code: 25, Comments: 4, Blanks: 3, Complexity: 7, Bytes: 320, Tokens: 75})
	assertFolder("internal/sub", 2, false, model.Metrics{Files: 1, Lines: 12, Code: 10, Comments: 1, Blanks: 1, Complexity: 3, Bytes: 120, Tokens: 30})
	assertFolder("(root files)", 1, true, model.Metrics{Files: 2, Lines: 15, Code: 8, Comments: 5, Blanks: 2, Complexity: 2, Bytes: 158, Tokens: 21})
	assertFolder("other/root", 0, false, model.Metrics{Files: 2, Lines: 17, Code: 14, Comments: 2, Blanks: 1, Complexity: 5, Bytes: 170, Tokens: 40})
	assertFolder("other/root/(root files)", 1, true, model.Metrics{Files: 1, Lines: 2, Code: 2, Bytes: 20, Tokens: 4})
}

func TestBuildFolderViewRepresentsDirectFileInputOnce(t *testing.T) {
	inputs := []model.InputRoot{{
		ID:    7,
		Given: "src/only.go",
		Kind:  model.InputFile,
	}}
	files := []model.FileRecord{{
		InputID:  7,
		Path:     "src/only.go",
		RelPath:  "only.go",
		Language: "Go",
		Metrics:  model.Metrics{Lines: 3, Code: 2, Tokens: 5},
	}}

	report, err := Build(inputs, files, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Folders) != 1 {
		t.Fatalf("folder rows = %+v, want one synthetic bucket", report.Folders)
	}
	wantMetrics := model.Metrics{Files: 1, Lines: 3, Code: 2, Tokens: 5}
	want := model.FolderRow{
		InputID:   7,
		Path:      "src/(root files)",
		Name:      rootFilesName,
		Depth:     0,
		Synthetic: true,
		Metrics:   wantMetrics,
	}
	if got := report.Folders[0]; got != want {
		t.Fatalf("direct-file folder row = %+v, want %+v", got, want)
	}
	if report.Totals != wantMetrics {
		t.Fatalf("totals = %+v, want %+v", report.Totals, wantMetrics)
	}
}

func TestDirectFileBucketPathIsDeterministic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"main.go", "(root files)"},
		{"./src/main.go", "./src/(root files)"},
		{"C:/repo/main.go", "C:/repo/(root files)"},
		{"/main.go", "/(root files)"},
		{"//server/share/main.go", "//server/share/(root files)"},
	}
	for _, test := range tests {
		if got := directFileBucketPath(test.input); got != test.want {
			t.Errorf("directFileBucketPath(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestFolderNameSortPreservesHierarchy(t *testing.T) {
	inputs, files := fixture()
	report, err := Build(inputs, files, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		".", "(root files)", "docs", "internal", "internal/sub",
		"other/root", "other/root/(root files)", "other/root/src",
	}
	if got := folderPaths(report.Folders); !reflect.DeepEqual(got, want) {
		t.Fatalf("name-sorted folders = %v, want %v", got, want)
	}
}

func TestAllSortKeysAndDeterministicTies(t *testing.T) {
	inputs, files := fixture()
	tests := []struct {
		key  model.SortKey
		want []string
	}{
		{model.SortTokens, []string{"Go", "JavaScript", "Python", "Markdown"}},
		{model.SortCode, []string{"Go", "JavaScript", "Python", "Markdown"}},
		{model.SortLines, []string{"Go", "JavaScript", "Python", "Markdown"}},
		{model.SortFiles, []string{"Go", "Markdown", "Python", "JavaScript"}},
		{model.SortName, []string{"Go", "JavaScript", "Markdown", "Python"}},
	}
	for _, test := range tests {
		t.Run(string(test.key), func(t *testing.T) {
			report, err := Build(inputs, files, model.ViewLanguage, test.key, model.Metadata{})
			if err != nil {
				t.Fatal(err)
			}
			if got := languageNames(report.Languages); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}

	fileSorts := []struct {
		key  model.SortKey
		want []string
	}{
		{model.SortTokens, []string{"internal/a.go", "other/root/src/app.js", "internal/sub/b.py", "main.go", "docs/x.md", "README.md", "other/root/root.py"}},
		{model.SortCode, []string{"internal/a.go", "other/root/src/app.js", "internal/sub/b.py", "main.go", "docs/x.md", "other/root/root.py", "README.md"}},
		{model.SortLines, []string{"internal/a.go", "other/root/src/app.js", "internal/sub/b.py", "main.go", "README.md", "docs/x.md", "other/root/root.py"}},
		{model.SortFiles, []string{"README.md", "docs/x.md", "internal/a.go", "internal/sub/b.py", "main.go", "other/root/root.py", "other/root/src/app.js"}},
		{model.SortName, []string{"README.md", "docs/x.md", "internal/a.go", "internal/sub/b.py", "main.go", "other/root/root.py", "other/root/src/app.js"}},
	}
	for _, test := range fileSorts {
		report, buildErr := Build(inputs, files, model.ViewFile, test.key, model.Metadata{})
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		if got := filePaths(report.Files); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("file sort %s = %v, want %v", test.key, got, test.want)
		}
	}

	folderSorts := []struct {
		key  model.SortKey
		want []string
	}{
		{model.SortTokens, []string{".", "internal", "internal/sub", "(root files)", "docs", "other/root", "other/root/src", "other/root/(root files)"}},
		{model.SortCode, []string{".", "internal", "internal/sub", "(root files)", "docs", "other/root", "other/root/src", "other/root/(root files)"}},
		{model.SortLines, []string{".", "internal", "internal/sub", "(root files)", "docs", "other/root", "other/root/src", "other/root/(root files)"}},
		{model.SortFiles, []string{".", "(root files)", "internal", "internal/sub", "docs", "other/root", "other/root/(root files)", "other/root/src"}},
		{model.SortName, []string{".", "(root files)", "docs", "internal", "internal/sub", "other/root", "other/root/(root files)", "other/root/src"}},
	}
	for _, test := range folderSorts {
		report, buildErr := Build(inputs, files, model.ViewFolder, test.key, model.Metadata{})
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		if got := folderPaths(report.Folders); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("folder sort %s = %v, want %v", test.key, got, test.want)
		}
	}

	baseline, err := Build(inputs, files, model.ViewFolder, model.SortTokens, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	for shift := range files {
		rotated := append(append([]model.FileRecord(nil), files[shift:]...), files[:shift]...)
		got, buildErr := Build(inputs, rotated, model.ViewFolder, model.SortTokens, model.Metadata{})
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		if !reflect.DeepEqual(got, baseline) {
			t.Fatalf("arrival order changed report at shift %d\ngot:  %+v\nwant: %+v", shift, got, baseline)
		}
	}
}

func TestSeparateRootsWithSameDisplayName(t *testing.T) {
	inputs := []model.InputRoot{{ID: 8, Given: "same"}, {ID: 3, Given: "same"}}
	files := []model.FileRecord{
		{InputID: 8, Path: "same/a.go", RelPath: "a.go", Language: "Go", Metrics: model.Metrics{Tokens: 1}},
		{InputID: 3, Path: "same/b.go", RelPath: "b.go", Language: "Go", Metrics: model.Metrics{Tokens: 1}},
	}
	report, err := Build(inputs, files, model.ViewFolder, model.SortTokens, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Folders) != 4 {
		t.Fatalf("got %d folder rows, want 4: %+v", len(report.Folders), report.Folders)
	}
	if report.Folders[0].InputID != 3 || report.Folders[2].InputID != 8 {
		t.Fatalf("same-name roots were not kept separate/deterministic: %+v", report.Folders)
	}
}

func TestFolderTreeKeepsSyntheticAndRealRootFilesFolderSeparate(t *testing.T) {
	inputs := []model.InputRoot{{ID: 0, Given: "."}}
	files := []model.FileRecord{
		{InputID: 0, Path: "main.go", RelPath: "main.go", Language: "Go", Metrics: model.Metrics{Tokens: 2}},
		{InputID: 0, Path: "(root files)/nested.go", RelPath: "(root files)/nested.go", Language: "Go", Metrics: model.Metrics{Tokens: 3}},
	}

	baseline, err := Build(inputs, files, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.Folders) != 3 {
		t.Fatalf("folder rows = %+v", baseline.Folders)
	}
	if !baseline.Folders[1].Synthetic || baseline.Folders[1].Metrics.Tokens != 2 {
		t.Fatalf("synthetic row = %+v", baseline.Folders[1])
	}
	if baseline.Folders[2].Synthetic || baseline.Folders[2].Metrics.Tokens != 3 {
		t.Fatalf("real row = %+v", baseline.Folders[2])
	}
	if baseline.Folders[1].Name != "(root files)" || baseline.Folders[2].Name != "(root files)/" {
		t.Fatalf("collision labels = %+v", baseline.Folders)
	}

	reversed := []model.FileRecord{files[1], files[0]}
	got, err := Build(inputs, reversed, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, baseline) {
		t.Fatalf("arrival order changed collision rows\ngot:  %+v\nwant: %+v", got, baseline)
	}
}

func TestLiteralBackslashRemainsAFileNameCharacter(t *testing.T) {
	inputs := []model.InputRoot{{ID: 0, Given: "."}}
	files := []model.FileRecord{{
		InputID: 0, Path: `a\b.go`, RelPath: `a\b.go`, Language: "Go",
		Metrics: model.Metrics{Tokens: 1},
	}}

	report, err := Build(inputs, files, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := folderPaths(report.Folders), []string{".", "(root files)"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("folders = %v, want %v", got, want)
	}
	fileReport, err := Build(inputs, files, model.ViewFile, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	if fileReport.Files[0].Path != `a\b.go` {
		t.Fatalf("literal backslash path = %q", fileReport.Files[0].Path)
	}
}

func TestFolderTreePreservesInputLabelsAsGiven(t *testing.T) {
	inputs := []model.InputRoot{
		{ID: 0, Given: "./src"},
		{ID: 1, Given: "src/"},
	}
	files := []model.FileRecord{
		{InputID: 0, Path: "src/a.go", RelPath: "pkg/a.go", Language: "Go", Metrics: model.Metrics{Tokens: 1}},
		{InputID: 1, Path: "src/b.go", RelPath: "pkg/b.go", Language: "Go", Metrics: model.Metrics{Tokens: 1}},
	}

	report, err := Build(inputs, files, model.ViewFolder, model.SortName, model.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"./src", "./src/pkg", "src/", "src/pkg"}
	if got := folderPaths(report.Folders); !reflect.DeepEqual(got, want) {
		t.Fatalf("folder labels = %v, want %v", got, want)
	}
}

func TestBuildValidation(t *testing.T) {
	validInput := []model.InputRoot{{ID: 0, Given: "."}}
	validFile := []model.FileRecord{{InputID: 0, Path: "a.go", RelPath: "a.go", Language: "Go"}}
	tests := []struct {
		name   string
		inputs []model.InputRoot
		files  []model.FileRecord
		view   model.View
		sort   model.SortKey
	}{
		{"invalid view", validInput, validFile, model.View(99), model.SortTokens},
		{"invalid sort", validInput, validFile, model.ViewLanguage, model.SortKey("bytes")},
		{"duplicate input", []model.InputRoot{{ID: 0}, {ID: 0}}, nil, model.ViewLanguage, model.SortTokens},
		{"unknown input", validInput, []model.FileRecord{{InputID: 9, Path: "a.go", RelPath: "a.go", Language: "Go"}}, model.ViewLanguage, model.SortTokens},
		{"missing language", validInput, []model.FileRecord{{InputID: 0, Path: "a.go", RelPath: "a.go"}}, model.ViewLanguage, model.SortTokens},
		{"escaping path", validInput, []model.FileRecord{{InputID: 0, Path: "a.go", RelPath: "../a.go", Language: "Go"}}, model.ViewLanguage, model.SortTokens},
		{"drive-rooted path", validInput, []model.FileRecord{{InputID: 0, Path: "a.go", RelPath: `C:/repo/a.go`, Language: "Go"}}, model.ViewLanguage, model.SortTokens},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Build(test.inputs, test.files, test.view, test.sort, model.Metadata{}); err == nil {
				t.Fatal("Build() unexpectedly succeeded")
			}
		})
	}
}

func fixture() ([]model.InputRoot, []model.FileRecord) {
	inputs := []model.InputRoot{
		{ID: 0, Given: ".", Abs: `C:\repo`, Kind: model.InputDirectory},
		{ID: 1, Given: `other/root`, Abs: `C:\other\root`, Kind: model.InputDirectory},
	}
	files := []model.FileRecord{
		{InputID: 0, Path: `main.go`, RelPath: `main.go`, Language: "Go", Metrics: model.Metrics{Lines: 10, Code: 8, Comments: 1, Blanks: 1, Complexity: 2, Bytes: 100, Tokens: 16}},
		{InputID: 0, Path: `README.md`, RelPath: `README.md`, Language: "Markdown", Metrics: model.Metrics{Lines: 5, Comments: 4, Blanks: 1, Bytes: 58, Tokens: 5}},
		{InputID: 0, Path: `internal/a.go`, RelPath: `internal/a.go`, Language: "Go", Metrics: model.Metrics{Lines: 20, Code: 15, Comments: 3, Blanks: 2, Complexity: 4, Bytes: 200, Tokens: 45}},
		{InputID: 0, Path: `internal/sub/b.py`, RelPath: `internal/sub/b.py`, Language: "Python", Metrics: model.Metrics{Lines: 12, Code: 10, Comments: 1, Blanks: 1, Complexity: 3, Bytes: 120, Tokens: 30}},
		{InputID: 0, Path: `docs/x.md`, RelPath: `docs/x.md`, Language: "Markdown", Metrics: model.Metrics{Lines: 4, Code: 3, Blanks: 1, Bytes: 40, Tokens: 6}},
		{InputID: 1, Path: `other/root/src/app.js`, RelPath: `src/app.js`, Language: "JavaScript", Metrics: model.Metrics{Lines: 15, Code: 12, Comments: 2, Blanks: 1, Complexity: 5, Bytes: 150, Tokens: 36}},
		{InputID: 1, Path: `other/root/root.py`, RelPath: `root.py`, Language: "Python", Metrics: model.Metrics{Lines: 2, Code: 2, Bytes: 20, Tokens: 4}},
	}
	return inputs, files
}

func languageNames(rows []model.LanguageRow) []string {
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row.Language
	}
	return result
}

func filePaths(rows []model.FileRecord) []string {
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row.Path
	}
	return result
}

func folderPaths(rows []model.FolderRow) []string {
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row.Path
	}
	return result
}

func containsBackslash(value string) bool {
	for _, character := range value {
		if character == '\\' {
			return true
		}
	}
	return false
}
