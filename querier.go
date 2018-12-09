package main

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
func (q *Querier) Query(input string) References {
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

	return results
}
