package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type promptSelectItem struct {
	Label       string
	Hint        string
	Value       string
	Group       string
	Description string
}

func isInteractiveTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func promptSelectSingle(label string, items []promptSelectItem, defaultIndex int) string {
	return promptSelectSingleWithSearch(label, items, defaultIndex, false)
}

func promptSelectSingleSearch(label string, items []promptSelectItem, defaultIndex int) string {
	return promptSelectSingleWithSearch(label, items, defaultIndex, true)
}

func promptSelectSingleWithSearch(label string, items []promptSelectItem, defaultIndex int, searchable bool) string {
	if len(items) == 0 {
		return ""
	}
	if defaultIndex < 0 || defaultIndex >= len(items) {
		defaultIndex = 0
	}
	if !isInteractiveTTY() {
		return promptSelectSingleFallback(label, items, defaultIndex)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return promptSelectSingleFallback(label, items, defaultIndex)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	reader := bufio.NewReader(os.Stdin)
	cursor := defaultIndex
	renderedLines := 0
	query := ""

	render := func() {
		visible := filterPromptSelectItems(items, query)
		if len(visible) == 0 {
			visible = []promptSelectItem{{Label: "No matching commands", Hint: "Keep typing or clear the filter", Value: ""}}
		}
		if cursor >= len(visible) {
			cursor = len(visible) - 1
		}
		if cursor < 0 {
			cursor = 0
		}
		if renderedLines > 0 {
			fmt.Printf("\x1b[%dA", renderedLines)
		}
		lines := make([]string, 0, len(visible)+4)
		lines = append(lines, clearLine(fmt.Sprintf("%s%s◉ %s%s", colorBold(), colorCyan(), label, colorReset())))
		if searchable {
			lines = append(lines, clearLine(fmt.Sprintf("%s   Filter: %s%s", colorDim(), nonEmpty(query, "type to search"), colorReset())))
			lines = append(lines, clearLine(fmt.Sprintf("%s   Use ↑/↓ (or j/k) • Type to filter • Backspace to edit • Enter to run%s", colorDim(), colorReset())))
		} else {
			lines = append(lines, clearLine(fmt.Sprintf("%s   Use ↑/↓ (or j/k) • Enter to run%s", colorDim(), colorReset())))
		}
		lastGroup := ""
		for i, item := range visible {
			if group := strings.TrimSpace(item.Group); group != "" && group != lastGroup {
				lines = append(lines, clearLine(fmt.Sprintf("   %s%s%s", colorCyan(), strings.ToUpper(group), colorReset())))
				lastGroup = group
			}
			pointer := " "
			if i == cursor {
				pointer = colorCyan() + "›" + colorReset()
			}
			mark := "( )"
			if i == cursor {
				mark = colorCyan() + "(•)" + colorReset()
			}
			labelText := item.Label
			if i == cursor {
				labelText = colorBold() + labelText + colorReset()
			}
			line := fmt.Sprintf(" %s %s %s", pointer, mark, labelText)
			if item.Hint != "" {
				line += fmt.Sprintf(" %s— %s%s", colorDim(), item.Hint, colorReset())
			}
			lines = append(lines, clearLine(line))
			if i == cursor && strings.TrimSpace(item.Description) != "" {
				lines = append(lines, clearLine(fmt.Sprintf("     %s%s%s", colorDim(), item.Description, colorReset())))
			}
		}
		for _, line := range lines {
			fmt.Printf("\r%s\n", line)
		}
		renderedLines = len(lines)
	}

	render()
	for {
		b, readErr := reader.ReadByte()
		if readErr != nil {
			break
		}
		switch b {
		case '\r', '\n':
			visible := filterPromptSelectItems(items, query)
			if len(visible) == 0 {
				continue
			}
			fmt.Printf("\x1b[%dA", renderedLines)
			for i := 0; i < renderedLines; i++ {
				fmt.Print("\r\x1b[2K\n")
			}
			fmt.Printf("\x1b[%dA", renderedLines)
			return visible[cursor].Value
		case 'k':
			if cursor > 0 {
				cursor--
				render()
			}
		case 'j':
			visible := filterPromptSelectItems(items, query)
			if cursor < len(visible)-1 {
				cursor++
				render()
			}
		case 127, 8:
			if searchable && len(query) > 0 {
				query = query[:len(query)-1]
				cursor = 0
				render()
			}
		case 3:
			fmt.Println()
			os.Exit(130)
		case 27:
			next, _ := reader.ReadByte()
			if next != '[' {
				continue
			}
			arrow, _ := reader.ReadByte()
			switch arrow {
			case 'A':
				if cursor > 0 {
					cursor--
					render()
				}
			case 'B':
				if cursor < len(items)-1 {
					cursor++
					render()
				}
			}
		default:
			if searchable && b >= 32 && b <= 126 {
				query += string(b)
				cursor = 0
				render()
			}
		}
	}

	return items[cursor].Value
}

func promptSelectSingleFallback(label string, items []promptSelectItem, defaultIndex int) string {
	fmt.Println()
	fmt.Printf("%s%s%s\n", colorBold(), label, colorReset())
	for i, item := range items {
		fmt.Printf("  %d) %s", i+1, item.Label)
		if item.Hint != "" {
			fmt.Printf(" %s— %s%s", colorDim(), item.Hint, colorReset())
		}
		fmt.Println()
	}
	fmt.Printf("\n%sChoice [%d]: %s", colorDim(), defaultIndex+1, colorReset())
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return items[defaultIndex].Value
	}
	var idx int
	if _, err := fmt.Sscanf(line, "%d", &idx); err != nil || idx < 1 || idx > len(items) {
		return items[defaultIndex].Value
	}
	return items[idx-1].Value
}

func clearLine(line string) string {
	return line + "\x1b[K"
}

func filterPromptSelectItems(items []promptSelectItem, query string) []promptSelectItem {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return append([]promptSelectItem(nil), items...)
	}
	filtered := make([]promptSelectItem, 0, len(items))
	for _, item := range items {
		hay := strings.ToLower(item.Label + " " + item.Hint + " " + item.Value)
		if strings.Contains(hay, query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
