package main

func IncreaseMatchCursor(m Model) Model {
	if m.MatchCursor < len(m.Memories)-1 {
		m.MatchCursor++
	}
	return m
}

func DecreaseMatchCursor(m Model) Model {
	if m.MatchCursor > 0 {
		m.MatchCursor--
	}
	return m
}
