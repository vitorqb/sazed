// This file contains an adapter interface and struct for
// github.com/lithammer/fuzzysearch/fuzzy
package main

import (
	"math"
	"sort"

	"github.com/sahilm/fuzzy"
)

// ScoreByDescription is a helper for calculating score matching by Description
type ScoreByDescription []Memory
func (m ScoreByDescription) String(i int) string { return m[i].Description }
func (m ScoreByDescription) Len() int { return len(m) }

// ScoreByCommand is a helper for calculating score matching by Command
type ScoreByCommand []Memory
func (m ScoreByCommand) String(i int) string { return m[i].Command }
func (m ScoreByCommand) Len() int { return len(m) }

// Implement sort
type SortByScore struct {
	memories []Memory
	scores   []int
}

// Implement Sort
func (s SortByScore) Len() int { return len(s.memories) }
func (s SortByScore) Less(i, j int) bool { return s.scores[i] > s.scores[j] }
func (s SortByScore) Swap(i, j int) {
	s.memories[i], s.memories[j] = s.memories[j], s.memories[i]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}

type IFuzzy interface {
	SortByMatch(memories []Memory, input string)
}

type Fuzzy struct {}

func (Fuzzy) SortByMatch(arr []Memory, input string) {
	// Matches `input` on both Command and Description
	matchesByDescription := fuzzy.FindFromNoSort(input, ScoreByDescription(arr))
	matchesByCommand  := fuzzy.FindFromNoSort(input, ScoreByCommand(arr))
	matchResults := append(matchesByDescription, matchesByCommand...)

	// Make a `scores` array that maps i -> Score, where i is the index for `arr`
	scores := make([]int, len(arr))
	for i := range scores {
		var found bool
		for _, result := range matchResults {
			if result.Index == i {
				scores[i] += result.Score
				found = true
			}
		}
		if !found {
			scores[i] = math.MinInt
		}
	}

	// Sort the original array by score
	sort.Sort(SortByScore{arr, scores})
}

func NewFuzzy() Fuzzy {
	return Fuzzy{}
}
