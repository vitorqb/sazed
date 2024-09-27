// This file contains an adapter interface and struct for
// github.com/lithammer/fuzzysearch/fuzzy
package main

import (
	"sort"

	"github.com/sahilm/fuzzy"
)

// Match represents a memory that matches a string input
type Match struct {
	Memory                    Memory
	Score                     int
	CommandMatchedIndexes     []int
	DescriptionMatchedIndexes []int
}

// ScoreByDescription is a helper for calculating score matching by Description
type ScoreByDescription []Memory

func (m ScoreByDescription) String(i int) string { return m[i].Description }
func (m ScoreByDescription) Len() int            { return len(m) }

// ScoreByCommand is a helper for calculating score matching by Command
type ScoreByCommand []Memory

func (m ScoreByCommand) String(i int) string { return m[i].Command }
func (m ScoreByCommand) Len() int            { return len(m) }

// IFuzzy is an interface for fuzzy matching memories with an input string
type IFuzzy interface {
	GetMatches(memories []Memory, input string) []Match
}

type Fuzzy struct{}

// GetMatches returns a list of fuzzy matches for
func (Fuzzy) GetMatches(memories []Memory, input string) []Match {
	// Handle special case of empty input
	if input == "" {
		var matches []Match
		for _, memory := range memories {
			matches = append(matches, Match{Memory: memory})
		}
		return matches
	}

	// Matches `input` on both Command and Description
	matchesByDescription := fuzzy.FindFromNoSort(input, ScoreByDescription(memories))
	matchesByCommand := fuzzy.FindFromNoSort(input, ScoreByCommand(memories))

	// matchesMap is a map of index -> Match
	matchesMap := make(map[int]Match)
	for _, result := range matchesByDescription {
		matchesMap[result.Index] = Match{
			Memory:                    memories[result.Index],
			Score:                     result.Score,
			DescriptionMatchedIndexes: result.MatchedIndexes,
		}
	}
	for _, result := range matchesByCommand {
		if match, ok := matchesMap[result.Index]; ok {
			match.Score += result.Score
			match.CommandMatchedIndexes = result.MatchedIndexes
		} else {
			matchesMap[result.Index] = Match{
				Memory:                memories[result.Index],
				Score:                 result.Score,
				CommandMatchedIndexes: result.MatchedIndexes,
			}
		}
	}

	// Convert the map to a list
	matches := make([]Match, 0, len(matchesMap))
	for _, match := range matchesMap {
		matches = append(matches, match)
	}

	// Sort the list by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}

func NewFuzzy() Fuzzy {
	return Fuzzy{}
}
