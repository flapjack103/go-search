package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"
)

// TODO: define a proper Completer for user input
var noopCompleter = prompt.Completer(func(arg1 prompt.Document) []prompt.Suggest { return []prompt.Suggest{} })

func main() {
	// default to current directory but if a directory is given use that one
	// as the root for parsing and indexing .go files
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

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

	fmt.Printf("Constructing index on %d files...\n", len(files))
	idx := buildIndex(files)

	fmt.Println("Building prefix tree...")
	trie := TrieFromIndex(idx)

	fmt.Println("Initializing search")
	q := NewQuerier(idx, trie)
	exec := NewExecutor(q)

	fmt.Println("Please use `exit` or `Ctrl-D` to exit this program.")
	defer fmt.Println("Bye!")

	p := prompt.New(
		exec.Run,
		noopCompleter,
		prompt.OptionTitle("go-search: interactive go program search client"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
	p.Run()
}
