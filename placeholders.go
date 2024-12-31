package main

import "strings"

func CountPlaceholders(s string) int {
	return len(GetPlaceholders(s))
}

type Placeholder struct {
	Name    string
	Pattern string
}

func GetPlaceholders(s string) []Placeholder {
	insideBrackets := false
	placeholders := []Placeholder{}
	currentPlaceholder := Placeholder{Pattern: "{{"}
	for i := 0; i < len(s)-1; i++ {
		char := s[i]
		nextChar := s[i+1]
		if insideBrackets {
			if char == '}' && nextChar == '}' {
				insideBrackets = false
				currentPlaceholder.Pattern += "}}"
				placeholders = append(placeholders, currentPlaceholder)
				currentPlaceholder = Placeholder{Pattern: "{{"}
			} else {
				currentPlaceholder.Name += string(char)
				currentPlaceholder.Pattern += string(char)
			}
		} else {
			if char == '{' && nextChar == '{' {
				insideBrackets = true
				i++
			}
		}
	}
	return placeholders
}

// Represents options for rendering a single placeholder
type RenderOpts struct {
	Optional bool
	Prefix   string
}

// Given a string `s`, replace the placeholder `p` using `userInput` and `opts`. Returns
// the full string.
func ReplacePlaceholder(s string, p Placeholder, replacement string, opts RenderOpts) string {
	// Not optional and no replacement - leave it there
	if !opts.Optional && replacement == "" {
		return s
	}

	// Optional and no replacement
	if opts.Optional && replacement == "" {
		// If we the pattern is surrounded by spaces, replace all of it to a single space.
		patternWithSpaces := " " + p.Pattern + " "
		if strings.Contains(s, patternWithSpaces) {
			return strings.ReplaceAll(s, patternWithSpaces, " ")
		}
		// Otherwise just replace the pattern
		return strings.ReplaceAll(s, p.Pattern, replacement)
	}

	// We have a replacement
	replacement = opts.Prefix + replacement
	return strings.ReplaceAll(s, p.Pattern, replacement)
}

// Gven a string `s` with placeholders like `{{foo}}`, render each placeholder
// using inputs from `userInputs` and options from `placeholdeOpts`
func Render(s string, userInputs []string, renderOpts []RenderOpts) string {
	for i, placeholder := range GetPlaceholders(s) {
		var userInput string
		if i < len(userInputs) {
			userInput = userInputs[i]
		}
		var renderOpt RenderOpts
		if i < len(renderOpts) {
			renderOpt = renderOpts[i]
		}
		s = ReplacePlaceholder(s, placeholder, userInput, renderOpt)
	}
	return s
}
