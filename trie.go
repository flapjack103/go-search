package main

// Terminator is used to mark the end of a word in the Trie
const Terminator = '\\'

// Trie is a prefix tree for fast searching
type Trie struct {
	value    rune
	children map[rune]*Trie
}

// NewTrie creates an empty Trie
func NewTrie() *Trie {
	return &Trie{
		children: make(map[rune]*Trie),
	}
}

// TrieFromIndex builds a prefix tree from the given index
func TrieFromIndex(idx *Index) *Trie {
	t := NewTrie()
	for word := range idx.wordMap {
		t.Insert(word)
	}
	return t
}

// Insert data into the Trie
func (t *Trie) Insert(word string) {
	node := t
	idx := -1
	for i, c := range word {
		n, ok := node.children[c]
		if !ok {
			idx = i
			break
		}
		node = n
	}

	if idx >= 0 {
		for _, r := range word[idx:] {
			node.children[r] = NewTrie()
			node = node.children[r]
		}
	}
	node.children[Terminator] = NewTrie()
}

// Find the word in the prefix tree, returns the node and true if it exists
// and false otherwise.
func (t *Trie) Find(word string) (*Trie, bool) {
	node := t
	for _, c := range word {
		n, ok := node.children[c]
		if !ok {
			return nil, false
		}
		node = n
	}

	return node, true
}

// Prefixes returns all prefixes stored in the trie
func (t *Trie) Prefixes() []string {
	return buildWords("", t)
}

func buildWords(prefix string, t *Trie) []string {
	var words []string

	for r, n := range t.children {
		if r == Terminator {
			words = append(words, prefix)
			continue
		}
		words = append(words, buildWords(prefix+string(r), n)...)
	}

	return words
}
