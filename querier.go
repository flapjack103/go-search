package main

import "sort"

const (
	// Filter types for query results

	// ResultsAll is no filter on type
	ResultsAll = "all"
	// ResultsFunctions filters on functions
	ResultsFunctions = "functions"
	// ResultsStructs filters on structs
	ResultsStructs = "structs"
	// ResultsVariables filters on variables
	ResultsVariables = "variables"

	// DefaultResultsLimit defines the number of results to return for query
	DefaultResultsLimit = 10
)

// QueryOptions defines filters and other options for querying
type QueryOptions struct {
	wtype string
	file  string
	limit int
}

// DefaultQueryOptions returns default settings for QueryOptions which is no
// filtering and a result limit 10
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		wtype: ResultsAll,
		file:  ResultsAll,
		limit: DefaultResultsLimit,
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
func (q *Querier) Query(input string, opts *QueryOptions) References {
	n, ok := q.trie.Find(input)
	if !ok {
		return nil
	}

	var results []Reference
	words := n.Prefixes()
	for _, w := range words {
		w = input + w
		refs, ok := q.idx.References(w)
		if !ok {
			continue
		}
		results = append(results, refs...)
	}

	// filter if needed
	resultsFiltered := results
	if opts.file != ResultsAll || opts.wtype != ResultsAll {
		resultsFiltered = []Reference{}
		for _, res := range results {
			if isMatch(res, opts) {
				resultsFiltered = append(resultsFiltered, res)
			}
		}
	}

	sort.Sort(sort.Reverse(SmartSort(resultsFiltered)))

	if len(resultsFiltered) > opts.limit {
		resultsFiltered = resultsFiltered[:opts.limit]
	}

	return resultsFiltered
}

func isMatch(ref Reference, opts *QueryOptions) bool {
	if opts.wtype != ResultsAll {
		// filter on type
		switch ref.(type) {
		case *Function:
			if opts.wtype != ResultsFunctions {
				return false
			}
		case *Variable:
			if opts.wtype != ResultsVariables {
				return false
			}
		case *Struct:
			if opts.wtype != ResultsStructs {
				return false
			}
		}
	}

	// filter on file location
	return opts.file == ResultsAll || ref.GetLocation().File == opts.file
}
