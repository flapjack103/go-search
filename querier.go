package main

import (
	"sort"
)

var defaultResultLimit = 20

// QueryFilter defines filtering options when executing a query, such as
// what files to include and (TODO) other filter types
type QueryFilter struct {
	files []string
}

// QueryOptions allow flexibility for querying by controlling how many results
// are returned, what filters are applied, and how the results are ranked
type QueryOptions struct {
	filter  *QueryFilter
	ranking RankBy
	limit   int
}

// DefaultQueryOptions provides defaults for querying
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		ranking: RankLexicographically,
		limit:   defaultResultLimit,
	}
}

// Querier manages the logic for returning search results
type Querier struct {
	idx  *Index
	trie *Trie
}

// NewQuerier returns a Querier object
func NewQuerier(idx *Index, trie *Trie) *Querier {
	return &Querier{
		idx,
		trie,
	}
}

// Query runs a query for the input and returns a list of results
func (q *Querier) Query(input string, options *QueryOptions) Results {
	n, ok := q.trie.Find(input)
	if !ok {
		return nil
	}

	var results Results
	words := n.Prefixes()
	for _, w := range words {
		w = input + w
		refs, ok := q.idx.References(w)
		if !ok {
			continue
		}
		for file, r := range refs {
			results = append(
				results,
				&Result{word: w, file: file, count: len(r)},
			)
		}
	}

	// Order the results
	switch {
	case options.ranking == RankCount:
		sort.Sort(ResultsByCount(results))
	case options.ranking == RankRelevance:
		sort.Sort(ResultsByRelevance(results))
	default:
		sort.Sort(results)
	}

	// Truncate if needed
	if len(results) > options.limit {
		results = results[:options.limit]
	}

	return results
}
