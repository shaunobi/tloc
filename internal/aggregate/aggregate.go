// Package aggregate converts per-file measurements into deterministic report
// views, including cumulative folder trees.
package aggregate

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/shaunobi/tloc/internal/model"
)

const (
	rootFilesName = "(root files)"
	rootFilesKey  = "\x00root-files"
)

// Build prepares a renderer-ready report. A FileRecord always represents one
// file, so its Files counter is normalized to one before aggregation.
func Build(inputs []model.InputRoot, files []model.FileRecord, view model.View, sortKey model.SortKey, metadata model.Metadata) (model.Report, error) {
	if !view.Valid() {
		return model.Report{}, fmt.Errorf("unsupported view %d", view)
	}
	if sortKey == "" {
		sortKey = model.SortTokens
	}
	if !sortKey.Valid() {
		return model.Report{}, fmt.Errorf("unsupported sort key %q", sortKey)
	}

	inputByID := make(map[int]model.InputRoot, len(inputs))
	for _, input := range inputs {
		if _, exists := inputByID[input.ID]; exists {
			return model.Report{}, fmt.Errorf("duplicate input ID %d", input.ID)
		}
		input.Given = normalizeRootPath(input.Given)
		inputByID[input.ID] = input
	}

	languages := make(map[string]model.Metrics)
	var totals model.Metrics
	preparedFiles := make([]model.FileRecord, 0, len(files))
	var tree *folderTree
	if view == model.ViewFolder {
		tree = newFolderTree()
	}

	for _, file := range files {
		input, ok := inputByID[file.InputID]
		if !ok {
			return model.Report{}, fmt.Errorf("file %q references unknown input ID %d", file.Path, file.InputID)
		}
		if file.Language == "" {
			return model.Report{}, fmt.Errorf("file %q has no language", file.Path)
		}

		file.Metrics.Files = 1
		file.RelPath = normalizePath(file.RelPath)
		if file.RelPath == "." || file.RelPath == "" {
			file.RelPath = basePath(file.Path)
		}
		if file.RelPath == "." || file.RelPath == "" {
			return model.Report{}, fmt.Errorf("file %q has no relative path", file.Path)
		}
		if file.RelPath == ".." || strings.HasPrefix(file.RelPath, "../") || path.IsAbs(file.RelPath) || hasWindowsDriveRoot(file.RelPath) {
			return model.Report{}, fmt.Errorf("file %q has relative path outside input root: %q", file.Path, file.RelPath)
		}
		file.Path = normalizePath(file.Path)
		if file.Path == "." || file.Path == "" {
			file.Path = joinOutputPath(input.Given, file.RelPath)
		}

		metrics := languages[file.Language]
		metrics.Add(file.Metrics)
		languages[file.Language] = metrics
		totals.Add(file.Metrics)

		if view == model.ViewFile {
			preparedFiles = append(preparedFiles, file)
		}
		if tree != nil {
			tree.add(input, file)
		}
	}

	languageRows := make([]model.LanguageRow, 0, len(languages))
	for language, metrics := range languages {
		languageRows = append(languageRows, model.LanguageRow{Language: language, Metrics: metrics})
	}
	sortLanguages(languageRows, sortKey)

	report := model.Report{
		Languages: languageRows,
		Totals:    totals,
		Metadata:  metadata,
	}
	if view == model.ViewFile {
		sortFiles(preparedFiles, sortKey)
		report.Files = preparedFiles
	}
	if tree != nil {
		report.Folders = tree.flatten(sortKey)
	}
	return report, nil
}

type folderTree struct {
	roots map[int]*folderNode
}

type folderNode struct {
	inputID    int
	name       string
	outputPath string
	depth      int
	synthetic  bool
	metrics    model.Metrics
	children   map[string]*folderNode
}

func newFolderTree() *folderTree {
	return &folderTree{roots: make(map[int]*folderNode)}
}

func (t *folderTree) add(input model.InputRoot, file model.FileRecord) {
	if input.Kind == model.InputFile {
		root := t.roots[input.ID]
		if root == nil {
			root = &folderNode{
				inputID:    input.ID,
				name:       rootFilesName,
				outputPath: directFileBucketPath(input.Given),
				synthetic:  true,
				children:   make(map[string]*folderNode),
			}
			t.roots[input.ID] = root
		}
		root.metrics.Add(file.Metrics)
		return
	}

	root := t.roots[input.ID]
	if root == nil {
		rootName := input.Given
		if rootName == "" {
			rootName = "."
		}
		root = &folderNode{
			inputID:    input.ID,
			name:       rootName,
			outputPath: rootName,
			children:   make(map[string]*folderNode),
		}
		t.roots[input.ID] = root
	}
	root.metrics.Add(file.Metrics)

	directory := path.Dir(file.RelPath)
	if directory == "." {
		direct := root.children[rootFilesKey]
		if direct == nil {
			direct = &folderNode{
				inputID:    input.ID,
				name:       rootFilesName,
				outputPath: joinOutputPath(root.outputPath, rootFilesName),
				depth:      1,
				synthetic:  true,
				children:   make(map[string]*folderNode),
			}
			root.children[rootFilesKey] = direct
		}
		direct.metrics.Add(file.Metrics)
		return
	}

	node := root
	currentPath := root.outputPath
	for index, segment := range strings.Split(directory, "/") {
		if segment == "" || segment == "." {
			continue
		}
		currentPath = joinOutputPath(currentPath, segment)
		child := node.children[segment]
		if child == nil {
			child = &folderNode{
				inputID:    input.ID,
				name:       segment,
				outputPath: currentPath,
				depth:      index + 1,
				children:   make(map[string]*folderNode),
			}
			node.children[segment] = child
		}
		child.metrics.Add(file.Metrics)
		node = child
	}
}

func (t *folderTree) flatten(sortKey model.SortKey) []model.FolderRow {
	roots := make([]*folderNode, 0, len(t.roots))
	for _, root := range t.roots {
		roots = append(roots, root)
	}
	sortFolderNodes(roots, sortKey)

	rows := make([]model.FolderRow, 0)
	var visit func(*folderNode)
	visit = func(node *folderNode) {
		displayName := node.name
		if !node.synthetic && node.depth == 1 && node.name == rootFilesName {
			// Keep the real directory visually distinct from the synthetic
			// direct-files bucket without changing its sort key.
			displayName += "/"
		}
		rows = append(rows, model.FolderRow{
			InputID:   node.inputID,
			Path:      node.outputPath,
			Name:      displayName,
			Depth:     node.depth,
			Synthetic: node.synthetic,
			Metrics:   node.metrics,
		})
		children := make([]*folderNode, 0, len(node.children))
		for _, child := range node.children {
			children = append(children, child)
		}
		sortFolderNodes(children, sortKey)
		for _, child := range children {
			visit(child)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	return rows
}

func sortLanguages(rows []model.LanguageRow, key model.SortKey) {
	sort.Slice(rows, func(i, j int) bool {
		if key != model.SortName {
			left, right := metricValue(rows[i].Metrics, key), metricValue(rows[j].Metrics, key)
			if left != right {
				return left > right
			}
		}
		return rows[i].Language < rows[j].Language
	})
}

func sortFiles(rows []model.FileRecord, key model.SortKey) {
	sort.Slice(rows, func(i, j int) bool {
		left, right := rows[i], rows[j]
		if key != model.SortName {
			leftValue, rightValue := metricValue(left.Metrics, key), metricValue(right.Metrics, key)
			if leftValue != rightValue {
				return leftValue > rightValue
			}
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Language != right.Language {
			return left.Language < right.Language
		}
		if left.InputID != right.InputID {
			return left.InputID < right.InputID
		}
		return left.RelPath < right.RelPath
	})
}

func sortFolderNodes(nodes []*folderNode, key model.SortKey) {
	sort.Slice(nodes, func(i, j int) bool {
		left, right := nodes[i], nodes[j]
		if key != model.SortName {
			leftValue, rightValue := metricValue(left.metrics, key), metricValue(right.metrics, key)
			if leftValue != rightValue {
				return leftValue > rightValue
			}
		}
		if left.name != right.name {
			return left.name < right.name
		}
		if left.outputPath != right.outputPath {
			return left.outputPath < right.outputPath
		}
		if left.synthetic != right.synthetic {
			return left.synthetic
		}
		return left.inputID < right.inputID
	})
}

func metricValue(metrics model.Metrics, key model.SortKey) int64 {
	switch key {
	case model.SortTokens:
		return metrics.Tokens
	case model.SortCode:
		return metrics.Code
	case model.SortLines:
		return metrics.Lines
	case model.SortFiles:
		return metrics.Files
	default:
		return 0
	}
}

func normalizePath(value string) string {
	if value == "" {
		return ""
	}
	return path.Clean(value)
}

func normalizeRootPath(value string) string {
	return value
}

func basePath(value string) string {
	normalized := normalizePath(value)
	if normalized == "" {
		return ""
	}
	return path.Base(normalized)
}

func hasWindowsDriveRoot(value string) bool {
	if len(value) < 3 || value[1] != ':' || value[2] != '/' {
		return false
	}
	letter := value[0]
	return (letter >= 'A' && letter <= 'Z') || (letter >= 'a' && letter <= 'z')
}

func joinOutputPath(root, relative string) string {
	root = normalizeRootPath(root)
	relative = normalizePath(relative)
	if root == "" || root == "." {
		return relative
	}
	return strings.TrimRight(root, "/") + "/" + strings.TrimLeft(relative, "/")
}

func directFileBucketPath(inputPath string) string {
	// Input paths are slash-normalized before aggregation. Derive the parent
	// lexically so labels such as "./src" and UNC prefixes remain as given.
	separator := strings.LastIndex(inputPath, "/")
	if separator < 0 {
		return rootFilesName
	}
	parent := inputPath[:separator]
	if parent == "" {
		parent = "/"
	}
	return joinOutputPath(parent, rootFilesName)
}
