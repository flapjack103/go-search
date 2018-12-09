package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// default to current directory but if a directory is given use that one
	// as the root for parsing and indexing .go files
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	fmt.Println("Building index...")
	idx := BuildIndex(getFiles(root))
	cs := idx.Functions().BuildCallStack()
	if err := cs.Write("static/tree.json"); err != nil {
		fmt.Printf("Error creating callstack json: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Building prefix tree...")
	trie := TrieFromIndex(idx)

	fmt.Println("Initializing search...")
	q := NewQuerier(idx, trie)

	s := NewServer(q)
	if err := s.Listen(); err != nil {
		fmt.Printf("Could not start server: %s\n", err)
		os.Exit(1)
	}
}

func getFiles(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failed to access path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})

	return files
}
