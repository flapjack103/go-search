package main

import "sort"

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type ByCounts []*WordCount

func (c ByCounts) Len() int           { return len(c) }
func (c ByCounts) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCounts) Less(i, j int) bool { return c[i].Count < c[j].Count }

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
	AvgCallDepth      int          `json:"avg_call_depth"`
	MaxCallDepth      int          `json:"max_call_depth"`
}

// Summary of stats from the index
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

	if len(wordCounts) > 10 {
		wordCounts = wordCounts[:10]
	}
	if len(varCounts) > 10 {
		varCounts = varCounts[:10]
	}
	if len(funcCounts) > 10 {
		funcCounts = funcCounts[:10]
	}

	return &Summary{
		FileCount:         len(x.files),
		Files:             x.files,
		UniqueWordCount:   len(x.references),
		FunctionCount:     len(x.functions),
		MCWords:           wordCounts,
		MCVariableNames:   varCounts,
		MCFunctionNames:   funcCounts,
		AvgFunctionLength: funcSizeTotal / len(x.functions),
		LargestFunction:   funcSizeMax,
	}
}
