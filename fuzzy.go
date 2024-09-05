// This file contains an adapter interface and struct for
// github.com/lithammer/fuzzysearch/fuzzy
package main

import (
	"math"
	"sort"

	"github.com/sahilm/fuzzy"
)

// MemoriesFuzzy is a helper for manipulating []Memory
type MemoriesFuzzy []Memory

// Implement https://github.com/sahilm/fuzzy
func (m MemoriesFuzzy) String(i int) string {
	return m[i].Description
}

// Implement https://github.com/sahilm/fuzzy
func (m MemoriesFuzzy) Len() int {
	return len(m)
}

// Implement sort
type ByScore struct {
	memories []Memory
	scores   []int
}

// Implement Sort
func (s ByScore) Len() int { return len(s.memories) }
func (s ByScore) Less(i, j int) bool { return s.scores[i] > s.scores[j] }
func (s ByScore) Swap(i, j int) {
	s.memories[i], s.memories[j] = s.memories[j], s.memories[i]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}

type IFuzzy interface {
	SortByMatch(memories []Memory, input string)
}

type Fuzzy struct {}

func (Fuzzy) SortByMatch(arr []Memory, input string) {
	memories := make(MemoriesFuzzy, len(arr))
	for i, x := range arr {
		memories[i] = x
	}
	results := fuzzy.FindFromNoSort(input, memories)

	// Make a `scores` array that maps i -> Score, where i is the index for `arr`
	scores := make([]int, len(arr))
	for i := range scores {
		var found bool
		for _, result := range results {
			if result.Index == i {
				scores[i] = result.Score
				found = true
				break
			}
		}
		if !found {
			scores[i] = math.MinInt
		}
	}

	// Sort the original array by score
	sort.Sort(ByScore{arr, scores})
}

func NewFuzzy() Fuzzy {
	return Fuzzy{}
}
