package main

func IncreaseCursor(m Model) Model {
	if (m.Cursor < len(m.Memories) - 1) {
		m.Cursor++
	}
	return m
}

func DecreaseCursor(m Model) Model {
	if (m.Cursor > 0) {
		m.Cursor--
	}
	return m
}
