package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// PreviewLineOffset determines how many lines above and below the requested
// line number we should show in the preview
const PreviewLineOffset = 5

// FileManager is a struct for managing files and file operations in the project
type FileManager struct {
	root  string
	files []string // relative file paths
}

// NewFileManager inits a FileManager from the given root. It digs into the root
// directory to capture all .go filepaths for the project
func NewFileManager(root string) *FileManager {
	fm := &FileManager{root: root}
	fm.findFiles()
	return fm
}

func (m *FileManager) findFiles() {
	m.files = []string{}
	filepath.Walk(m.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failed to access path %q: %v\n", path, err)
			return err
		}
		// don't walk the vendor directory
		if info.IsDir() && info.Name() == "vendor" {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			m.files = append(m.files, path)
		}
		return nil
	})
}

// Rel returns the relative path for the given target path
func (m *FileManager) Rel(targpath string) string {
	relpath, _ := filepath.Rel(m.root, targpath)
	return relpath
}

// Preview is the response type for a code preview. It contains a formatted
// string of the code snippet.
type Preview struct {
	Code string `json:"code"`
}

// GetFilePreview finds the file and reads the area around the requested line
// number to generate a code Preview. Return an error if the file is not found
// or an error is encountered while reading the file.
func (m *FileManager) GetFilePreview(file string, line int) (*Preview, error) {
	filepath := path.Join(m.root, file)

	input, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	start := line - PreviewLineOffset
	if line < 0 {
		start = 0
	}

	var lines []string
	scanner := bufio.NewScanner(input)
	for pos := 0; scanner.Scan() && pos < line+PreviewLineOffset; pos++ {
		if pos < start {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d\t%s", pos+1, scanner.Text()))
	}
	return &Preview{strings.Join(lines, "\n")}, scanner.Err()
}
