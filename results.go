package main

import (
	"os"

	"github.com/jedib0t/go-pretty/table"
)

// RankBy defines the type of ranking algorithm we want to use
type RankBy byte

// Enums of the types of ranking options
const (
	RankLexicographically RankBy = iota
	RankCount
	RankRelevance
)

// Result holds the search result information
type Result struct {
	word  string
	file  string
	count int
}

// Results is a list of Result objects
type Results []*Result

// Print Results list in a table view
func (r Results) Print() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Word", "File", "Count"})
	for i, res := range r {
		t.AppendRow([]interface{}{i + 1, res.word, res.file, res.count})
	}
	t.Render()
}

// Default sort for results - lexicographically by word first and secondary sort
// by count
func (r Results) Len() int      { return len(r) }
func (r Results) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool {
	switch {
	case r[i].word < r[j].word:
		return true
	case r[i].word > r[j].word:
		return false
	default:
		return r[i].count < r[j].count
	}
}

// ResultsByCount sorts results by the number of times they occur and secondary
// sort lexicographically
type ResultsByCount []*Result

func (r ResultsByCount) Len() int      { return len(r) }
func (r ResultsByCount) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ResultsByCount) Less(i, j int) bool {
	switch {
	case r[i].count < r[j].count:
		return true
	case r[i].count > r[j].count:
		return false
	default:
		return r[i].word < r[j].word
	}
}

// ResultsByRelevance defines result sorting by relevance
// TODO: what do we mean by relevance?
type ResultsByRelevance []*Result

func (r ResultsByRelevance) Len() int      { return len(r) }
func (r ResultsByRelevance) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ResultsByRelevance) Less(i, j int) bool {
	switch {
	case r[i].word < r[j].word:
		return true
	case r[i].word > r[j].word:
		return false
	default:
		return r[i].count < r[j].count
	}
}
