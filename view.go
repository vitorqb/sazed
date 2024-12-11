package main

import (
	"fmt"
	"strings"
)

func ViewCommandSelection(m Model) string {
	body := "Please select a command\n"
	body += m.SearchTextInput.View() + "\n"
	body += "----------------------\n"

	for i, match := range m.Matches {
		cursor := " "
		if i == m.MatchCursor {
			cursor = ">>"
		}
		printLength := fmt.Sprintf("%d", m.AppOpts.CommandPrintLength)

		// Prints command on first line
		format := "%-2s %-" + printLength + "." + printLength + "s\n"
		body += fmt.Sprintf(format, cursor, match.Memory.Command)

		// Prints description on second line
		body += fmt.Sprintf("      |%s\n", match.Memory.Description)
	}

	return body
}

func ViewCommandEdit(m Model) string {
	// Displays the command with the placeholders replaced by the values
	originalCmd := m.SelectedMemory.Command
	placeholderValues := m.GetPlaceholderValues()
	renderedCmd := Render(originalCmd, placeholderValues)
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("Command: ")
	stringBuilder.WriteString(renderedCmd)
	stringBuilder.WriteString("\n")

	// Allow user to input values for each placeholder
	for _, input := range m.EditTextInputs {
		stringBuilder.WriteString(input.View())
		stringBuilder.WriteString("\n")
	}

	return stringBuilder.String()
}
