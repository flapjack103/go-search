package main

import (
	"fmt"
	"os"
)

func main() {
	// default to current directory but if a directory is given use that one
	// as the root for parsing and indexing .go files
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	// fetch all project files
	fmt.Println("Fetching files...")
	fm := NewFileManager(root)

	// construct the index
	fmt.Println("Building index...")
	idx := BuildIndex(fm)

	// build the function tree
	fmt.Println("Building callstack...")
	cs := idx.Functions().BuildCallStack()
	if err := cs.Write("static/tree.json"); err != nil {
		fmt.Printf("Error creating callstack json: %s\n", err)
		os.Exit(1)
	}

	// build the trie
	fmt.Println("Building prefix tree...")
	trie := TrieFromIndex(idx)

	// init the querier
	fmt.Println("Initializing search...")
	q := NewQuerier(idx, trie)

	// start the http server listening, default port is :8080
	s := NewServer(q, fm)
	if err := s.Listen(); err != nil {
		fmt.Printf("Could not start server: %s\n", err)
		os.Exit(1)
	}
}
