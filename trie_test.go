package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieBuild(t *testing.T) {
	assert := assert.New(t)
	words := []string{
		"t",
		"tag",
		"tags",
		"test",
		"string",
		"stripe",
		"set",
	}

	idx := newIndex(nil)
	for _, w := range words {
		idx.wordMap[w] = nil
	}

	tri := TrieFromIndex(idx)
	assert.NotNil(tri)

	// top level sanity check (s and t)
	assert.Len(tri.children, 2)
	assert.Contains(tri.children, 's')
	assert.Contains(tri.children, 't')

	// retrieve all words from the trie
	twords := tri.Prefixes()

	sort.Strings(words)
	sort.Strings(twords)
	assert.EqualValues(words, twords)
}

func TestTrieFind(t *testing.T) {
	assert := assert.New(t)
	words := []string{
		"t",
		"tag",
		"tags",
		"test",
		"string",
		"stripe",
		"set",
	}

	idx := newIndex(nil)
	for _, w := range words {
		idx.wordMap[w] = nil
	}

	tri := TrieFromIndex(idx)
	assert.NotNil(tri)

	_, ok := tri.Find("tag")
	assert.True(ok)

	n, ok := tri.Find("ta")
	assert.True(ok)
	twords := n.Prefixes()
	sort.Strings(twords)

	assert.Equal([]string{"g", "gs"}, twords)

	_, ok = tri.Find("tagz")
	assert.False(ok)
}
