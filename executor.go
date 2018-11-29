package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type handlerFunc func([]string) error

var cmdHelp = map[string]string{
	"search": `usage: search <word> [options]

    Description: The search command searches go source files for the given word. Works
                 on partial matches (eg. 'err' will return instances of 'err' and also
                  'error') making it equivalent to <word>*.

    The options are as follows:

    -l    Number of results to display (limit). Default is 20.

    -s    Sort type which determines how results are ranked. Available sorting types
          are 'lex' (lexicographically), 'count', and 'rel' (relevance). Default is
          'lex.'

    -f    Filter on a subset of files to search within. Paths must be absolute or
          relative to current directory. Pass as a comma seperated list. No filter
          by default.

    Example: search err -l 10 -s rel -f main.go
    `,
	"list": `usage: list <word> [options]

      Description: The list command returns all files and locations for instances of
                   the given word. Returns only exact matches of the word.

      The options are as follows:

      -l    Number of results to display (limit). Default is 20.

      -s    Sort type which determines how results are ranked. Available sorting types
            are 'pos' (position) which sorts lexicographically by filename then by position
            in the file (ascending by line and column), and 'rel' (relevance).

      -f    Filter on a subset of files to search within. Paths must be absolute or
            relative to current directory. Pass as a comma seperated list. No filter
            by default.

      Example: list err -l 10 -s rel -f main.go
    `,
	"help": `usage: help

      Description: Outputs available commands
    `,
}

// Executor is the cmd line program driver for mapping input to functionality
type Executor struct {
	handlers map[string]handlerFunc
	querier  *Querier
}

// NewExecutor creates an Executor. It takes a Querier for executing queries.
func NewExecutor(querier *Querier) *Executor {
	e := &Executor{
		querier: querier,
	}

	e.handlers = map[string]handlerFunc{
		"search": e.search,
		"list":   e.list,
		"help":   e.help,
		"man":    e.man,
	}

	return e
}

// Run parses the user input and runs the command if it's defined and valid
func (e *Executor) Run(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	} else if input == "quit" || input == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	parts := strings.Split(input, " ")
	if len(parts) == 0 {
		return
	}
	cmd := parts[0]
	args := parts[1:]

	fn, ok := e.handlers[cmd]
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmd)
		return
	}

	if err := fn(args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func parseSortOption(opt string) (RankBy, error) {
	opt = strings.ToLower(opt)
	switch opt {
	case "lex":
		return RankLexicographically, nil
	case "count":
		return RankCount, nil
	case "rel":
		return RankRelevance, nil
	default:
		return RankLexicographically, fmt.Errorf("Unknown sort option %s", opt)
	}
}

// TODO
func parseLimitOption(opt string) (int, error) {
	return defaultResultLimit, nil
}

// TODO
func parseFilterOption(opt string) (*QueryFilter, error) {
	return &QueryFilter{}, nil
}

func parseSearchArgs(args []string, opts *QueryOptions) (word string, err error) {
	if len(args) == 0 {
		return "", errors.New("Not enough arguments")
	}
	word = args[0]
	for i := 1; i < len(args); i += 2 {
		a := args[i]
		switch a {
		case "-s":
			if i+1 >= len(args) {
				return word, errors.New("-s requires a sort type")
			}
			opts.ranking, err = parseSortOption(a)
			if err != nil {
				return word, err
			}
		case "-f":
			if i+1 >= len(args) {
				return word, errors.New("-f requires a list of files")
			}
			opts.filter, err = parseFilterOption(a)
			if err != nil {
				return word, err
			}
		case "-l":
			if i+1 >= len(args) {
				return word, errors.New("-l requires a number")
			}
			opts.limit, err = parseLimitOption(a)
			if err != nil {
				return word, err
			}
		default:
			return word, fmt.Errorf("Unknown option %s", a)
		}
	}
	return
}

func (e *Executor) search(args []string) error {
	if len(args) == 0 {
		return errors.New("'search' requires at least one argument")
	}

	opts := DefaultQueryOptions()
	word, err := parseSearchArgs(args, opts)
	if err != nil {
		return err
	}

	results := e.querier.Query(word, opts)
	results.Print()

	return nil
}

func (e *Executor) list(args []string) error {
	if len(args) == 0 {
		return errors.New("'show' requires at least one argument")
	}

	refs, ok := e.querier.idx.References(args[0])
	if !ok {
		return fmt.Errorf("No references found for word %s", args[0])
	}

	refs.Print()

	return nil
}

func (e *Executor) help(args []string) error {
	fmt.Printf("Commands: search, list\n\nFor more information try 'man search' etc.\n")
	return nil
}

func (e *Executor) man(args []string) error {
	if len(args) == 0 {
		return errors.New("'man' requires at least one argument")
	}
	msg, ok := cmdHelp[args[0]]
	if !ok {
		return fmt.Errorf("No such command %s", args[0])
	}

	fmt.Println(msg)
	return nil
}
