package main

import (
	"sort"
)

// SummaryTopResultsLimit is the number of results to return for each
// 'top' list category
const SummaryTopResultsLimit = 10

// WordCount is the type returned for the 'top' lists in the summary
type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// ByCounts sorts WordCounts by their Count field
type ByCounts []*WordCount

func (c ByCounts) Len() int           { return len(c) }
func (c ByCounts) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCounts) Less(i, j int) bool { return c[i].Count < c[j].Count }

// Summary is the type returned for the code project summary. It encapsulates
// interesting information about the code as a whole and also 'top' lists
type Summary struct {
	idx               *Index
	FileCount         int          `json:"file_count"`
	Files             []string     `json:"files"`
	UniqueWordCount   int          `json:"uniq_word_count"`
	FunctionCount     int          `json:"func_count"`
	MCWords           []*WordCount `json:"most_common_words"`
	MCVariableNames   []*WordCount `json:"most_common_vars"`
	MCFunctionNames   []*WordCount `json:"most_common_funcs"`
	AvgFunctionLength int          `json:"avg_func_len"`
	LargestFunction   *Function    `json:"largest_func"`
}

// Summary generates code stats and information from the Index
func (x *Index) Summary() *Summary {
	var (
		wordCounts    []*WordCount
		varCounts     []*WordCount
		funcCounts    []*WordCount
		funcSizeTotal int
		funcSizeMax   *Function
	)
	for v, refs := range x.references {
		vc := &WordCount{Word: v}
		fc := &WordCount{Word: v}
		for _, ref := range refs {
			switch r := ref.(type) {
			case *Variable:
				if r.IsDecl {
					vc.Count++
				}
			case *Function:
				if r.IsDecl {
					fc.Count++
					funcSizeTotal += r.Size
					if funcSizeMax == nil || r.Size > funcSizeMax.Size {
						funcSizeMax = r
					}
				}
			}
		}
		wordCounts = append(wordCounts, &WordCount{v, len(refs)})
		if vc.Count > 0 {
			varCounts = append(varCounts, vc)
		}
		if fc.Count > 0 {
			funcCounts = append(funcCounts, fc)
		}
	}

	sort.Sort(sort.Reverse(ByCounts(wordCounts)))
	sort.Sort(sort.Reverse(ByCounts(varCounts)))
	sort.Sort(sort.Reverse(ByCounts(funcCounts)))

	// truncate results for 'top' list
	if len(wordCounts) > SummaryTopResultsLimit {
		wordCounts = wordCounts[:SummaryTopResultsLimit]
	}
	if len(varCounts) > SummaryTopResultsLimit {
		varCounts = varCounts[:SummaryTopResultsLimit]
	}
	if len(funcCounts) > SummaryTopResultsLimit {
		funcCounts = funcCounts[:SummaryTopResultsLimit]
	}

	// annoying, but these need to be relative to root
	fileCount := len(x.fileMgr.files)
	relFiles := make([]string, 0, fileCount)
	for _, f := range x.fileMgr.files {
		relFiles = append(relFiles, x.fileMgr.Rel(f))
	}

	return &Summary{
		FileCount:         fileCount,
		Files:             relFiles,
		UniqueWordCount:   len(x.references),
		FunctionCount:     len(x.functions),
		MCWords:           wordCounts,
		MCVariableNames:   varCounts,
		MCFunctionNames:   funcCounts,
		AvgFunctionLength: funcSizeTotal / len(x.functions),
		LargestFunction:   funcSizeMax,
	}
}
