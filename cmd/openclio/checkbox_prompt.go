package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	initBlue  = "\033[38;5;39m"
	initDim   = "\033[2m"
	initBold  = "\033[1m"
	initReset = "\033[0m"
)

func initAnsi(code string) string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return code
	}
	return ""
}

func initBlueText() string  { return initAnsi(initBlue) }
func initDimText() string   { return initAnsi(initDim) }
func initBoldText() string  { return initAnsi(initBold) }
func initResetText() string { return initAnsi(initReset) }

type checkboxItem struct {
	Label   string
	Hint    string
	Value   string
	Checked bool
}

type selectItem struct {
	Label string
	Hint  string
	Value string
}

func promptCheckboxMulti(label string, items []checkboxItem, defaultIndex int) []string {
	if len(items) == 0 {
		return nil
	}
	if defaultIndex < 0 || defaultIndex >= len(items) {
		defaultIndex = 0
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return promptCheckboxMultiFallback(label, items)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return promptCheckboxMultiFallback(label, items)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	reader := bufio.NewReader(os.Stdin)
	cursor := defaultIndex
	renderedLines := 0

	render := func() {
		if renderedLines > 0 {
			fmt.Printf("\x1b[%dA", renderedLines)
		}
		lines := make([]string, 0, len(items)+2)
		lines = append(lines, fmt.Sprintf("%s%s◉ %s%s", initBoldText(), initBlueText(), label, initResetText()))
		lines = append(lines, fmt.Sprintf("%s   Space = toggle • Enter = confirm%s", initDimText(), initResetText()))
		for i, item := range items {
			pointer := " "
			if i == cursor {
				pointer = initBlueText() + "›" + initResetText()
			}
			box := "[ ]"
			if item.Checked {
				box = initBlueText() + "[x]" + initResetText()
			}
			labelText := item.Label
			if i == cursor {
				labelText = initBoldText() + labelText + initResetText()
			}
			line := fmt.Sprintf(" %s %s %s", pointer, box, labelText)
			if item.Hint != "" {
				line += fmt.Sprintf(" %s— %s%s", initDimText(), item.Hint, initResetText())
			}
			lines = append(lines, clearLine(line))
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
			fmt.Printf("\x1b[%dA", renderedLines)
			for i := 0; i < renderedLines; i++ {
				fmt.Print("\r\x1b[2K\n")
			}
			fmt.Printf("\x1b[%dA", renderedLines)
			return checkedValues(items)
		case ' ':
			items[cursor].Checked = !items[cursor].Checked
			render()
		case 'k':
			if cursor > 0 {
				cursor--
				render()
			}
		case 'j':
			if cursor < len(items)-1 {
				cursor++
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
		}
	}

	return checkedValues(items)
}

func promptSelectSingle(label string, items []selectItem, defaultIndex int) string {
	if len(items) == 0 {
		return ""
	}
	if defaultIndex < 0 || defaultIndex >= len(items) {
		defaultIndex = 0
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
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

	render := func() {
		if renderedLines > 0 {
			fmt.Printf("\x1b[%dA", renderedLines)
		}
		lines := make([]string, 0, len(items)+2)
		lines = append(lines, fmt.Sprintf("%s%s◉ %s%s", initBoldText(), initBlueText(), label, initResetText()))
		lines = append(lines, fmt.Sprintf("%s   Use ↑/↓ to move • Enter to confirm%s", initDimText(), initResetText()))
		for i, item := range items {
			pointer := " "
			if i == cursor {
				pointer = initBlueText() + "›" + initResetText()
			}
			mark := "( )"
			if i == cursor {
				mark = initBlueText() + "(•)" + initResetText()
			}
			labelText := item.Label
			if i == cursor {
				labelText = initBoldText() + labelText + initResetText()
			}
			line := fmt.Sprintf(" %s %s %s", pointer, mark, labelText)
			if item.Hint != "" {
				line += fmt.Sprintf(" %s— %s%s", initDimText(), item.Hint, initResetText())
			}
			lines = append(lines, clearLine(line))
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
			fmt.Printf("\x1b[%dA", renderedLines)
			for i := 0; i < renderedLines; i++ {
				fmt.Print("\r\x1b[2K\n")
			}
			fmt.Printf("\x1b[%dA", renderedLines)
			return items[cursor].Value
		case 'k':
			if cursor > 0 {
				cursor--
				render()
			}
		case 'j':
			if cursor < len(items)-1 {
				cursor++
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
		}
	}

	return items[cursor].Value
}

func promptCheckboxMultiFallback(label string, items []checkboxItem) []string {
	fmt.Printf("📝 %s\n", label)
	fmt.Println("   Enter comma-separated numbers to select multiple items.")
	for i, item := range items {
		fmt.Printf("   %d) %s", i+1, item.Label)
		if item.Hint != "" {
			fmt.Printf(" — %s", item.Hint)
		}
		fmt.Println()
	}
	fmt.Print("   Selection: ")
	line, _ := initReader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return checkedValues(items)
	}
	parts := strings.Split(line, ",")
	selected := make([]string, 0, len(parts))
	seen := make(map[string]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(items) {
			continue
		}
		value := items[idx-1].Value
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		selected = append(selected, value)
	}
	return selected
}

func promptSelectSingleFallback(label string, items []selectItem, defaultIndex int) string {
	fmt.Printf("📝 %s\n", label)
	for i, item := range items {
		fmt.Printf("   %d) %s", i+1, item.Label)
		if item.Hint != "" {
			fmt.Printf(" — %s", item.Hint)
		}
		fmt.Println()
	}
	defaultChoice := fmt.Sprintf("%d", defaultIndex+1)
	choices := make([]string, 0, len(items))
	for i := range items {
		choices = append(choices, fmt.Sprintf("%d", i+1))
	}
	choice := promptChoice("Selection", choices, defaultChoice)
	index := defaultIndex
	fmt.Sscanf(choice, "%d", &index)
	if index < 1 || index > len(items) {
		return items[defaultIndex].Value
	}
	return items[index-1].Value
}

func checkedValues(items []checkboxItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		if item.Checked {
			values = append(values, item.Value)
		}
	}
	return values
}

func clearLine(line string) string {
	return line + "\x1b[K"
}
