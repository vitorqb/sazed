package main

func CountPlaceholders(s string) int {
	return len(GetPlaceholders(s))
}

type Placeholder struct {
	Beg  int
	End  int
	Name string
}

func GetPlaceholders(s string) []Placeholder {
	insideBrackets := false
	placeholders := []Placeholder{}
	currentPlaceholder := Placeholder{}
	for i := 0; i < len(s)-1; i++ {
		char := s[i]
		nextChar := s[i+1]
		if insideBrackets {
			if char == '}' && nextChar == '}' {
				insideBrackets = false
				currentPlaceholder.End = i + 1
				placeholders = append(placeholders, currentPlaceholder)
				currentPlaceholder = Placeholder{}
			} else {
				currentPlaceholder.Name += string(char)
			}
		} else {
			if char == '{' && nextChar == '{' {
				insideBrackets = true
				currentPlaceholder.Beg = i
				i++
			}
		}
	}
	return placeholders
}

func NextPlaceholder(s string) (p Placeholder, success bool) {
	placeholders := GetPlaceholders(s)
	if len(placeholders) == 0 {
		return Placeholder{}, false
	}
	return placeholders[0], true
}

func ReplacePlaceholder(s string, p Placeholder, replacement string) string {
	beg := s[:p.Beg]
	end := s[p.End+1:]
	return beg + replacement + end
}

// Given a string `s` with placeholders like `{{foo}}`, replace them with the values in `placeholderValues`. The `i`th placeholder should be replaced with the `i`th value in `placeholderValues`.
func Render(s string, placeholderValues []string) string {
	for i := 0; true; i++ {
		placeholder, success := NextPlaceholder(s)
		if !success {
			break
		}
		value := ""
		if i < len(placeholderValues) {
			value = placeholderValues[i]
		}
		s = ReplacePlaceholder(s, placeholder, value)
	}
	return s
}
