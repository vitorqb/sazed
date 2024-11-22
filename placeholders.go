package main

func CountPlaceholders(m Memory) int {
	insideBrackets := false
	placeholdersCount := 0
	for i := 0; i < len(m.Command)-1; i++ {
		char := m.Command[i]
		nextChar := m.Command[i+1]
		if insideBrackets {
			if char == '}' && nextChar == '}' {
				insideBrackets = false
				continue
			}
		} else {
			if char == '{' && nextChar == '{' {
				insideBrackets = true
				placeholdersCount++
			}
		}
	}
	return placeholdersCount
}
