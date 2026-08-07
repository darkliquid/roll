package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/darkliquid/roll"
)

const visibleHistory = 8

type historyEntry struct {
	expression string
	output     string
	failed     bool
}

type rollResultMsg struct {
	expression string
	output     string
	err        error
}

type model struct {
	input            []rune
	cursor           int
	history          []historyEntry
	historyInputs    []string
	historyIndex     int
	historyDraft     string
	quitting         bool
	evaluator        func(string) (string, error)
	clipboardReady   func() tea.Msg
	catalog          []Component
	completionActive bool
	completionItems  []Component
	completionIndex   int
}

func newModel() model {
	return model{
		evaluator:      evaluateExpression,
		clipboardReady: tea.ReadClipboard,
		catalog:        DefaultCatalog(),
	}
}

func evaluateExpression(expression string) (string, error) {
	return roll.ParseString(expression)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.completionActive {
				m.completionActive = false
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "tab", "ctrl+space", "ctrl+ ":
			if m.completionActive {
				if len(m.completionItems) > 0 {
					m.completionIndex = (m.completionIndex + 1) % len(m.completionItems)
				}
			} else {
				items := FilterComponents(m.catalog, string(m.input), m.cursor)
				if len(items) > 0 {
					m.completionActive = true
					m.completionItems = items
					m.completionIndex = 0
				}
			}
			return m, nil
		case "enter":
			if m.completionActive {
				m.applyCompletion()
				return m, nil
			}
			expression := strings.TrimSpace(string(m.input))
			if expression == "" {
				return m, nil
			}
			m.rememberExpression(expression)
			m.input = nil
			m.cursor = 0
			m.resetHistoryNavigation()
			return m, m.submitRoll(expression)
		case "up", "ctrl+p":
			if m.completionActive {
				if len(m.completionItems) > 0 {
					m.completionIndex = (m.completionIndex - 1 + len(m.completionItems)) % len(m.completionItems)
				}
			} else {
				m.navigateHistory(-1)
			}
			return m, nil
		case "down", "ctrl+n":
			if m.completionActive {
				if len(m.completionItems) > 0 {
					m.completionIndex = (m.completionIndex + 1) % len(m.completionItems)
				}
			} else {
				m.navigateHistory(1)
			}
			return m, nil
		case "ctrl+v":
			return m, func() tea.Msg {
				return m.clipboardReady()
			}
		case "left", "ctrl+b":
			if m.cursor > 0 {
				m.cursor--
				m.refilterCompletion()
			}
			return m, nil
		case "right", "ctrl+f":
			if m.cursor < len(m.input) {
				m.cursor++
				m.refilterCompletion()
			}
			return m, nil
		case "home", "ctrl+a":
			m.cursor = 0
			m.refilterCompletion()
			return m, nil
		case "end", "ctrl+e":
			m.cursor = len(m.input)
			m.refilterCompletion()
			return m, nil
		case "backspace":
			if m.cursor > 0 {
				m.input = append(m.input[:m.cursor-1], m.input[m.cursor:]...)
				m.cursor--
				m.refilterCompletion()
			}
			return m, nil
		case "delete":
			if m.cursor < len(m.input) {
				m.input = append(m.input[:m.cursor], m.input[m.cursor+1:]...)
				m.refilterCompletion()
			}
			return m, nil
		case "ctrl+u":
			m.input = nil
			m.cursor = 0
			m.completionActive = false
			m.resetHistoryNavigation()
			return m, nil
		default:
			if text := msg.Key().Text; text != "" {
				m.insertRunes([]rune(text))
			}
			return m, nil
		}
	case tea.PasteMsg:
		m.insertRunes([]rune(msg.Content))
		return m, nil
	case tea.ClipboardMsg:
		if msg.Content != "" {
			m.insertRunes([]rune(msg.Content))
		}
		return m, nil
	case rollResultMsg:
		entry := historyEntry{
			expression: msg.expression,
			failed:     msg.err != nil,
		}
		if msg.err != nil {
			entry.output = msg.err.Error()
		} else {
			entry.output = msg.output
		}
		m.history = append([]historyEntry{entry}, m.history...)
		return m, nil
	}

	return m, nil
}

func (m model) submitRoll(expression string) tea.Cmd {
	evaluator := m.evaluator
	return func() tea.Msg {
		output, err := evaluator(expression)
		return rollResultMsg{
			expression: expression,
			output:     output,
			err:        err,
		}
	}
}

func (m *model) insertRunes(runes []rune) {
	if len(runes) == 0 {
		return
	}
	m.resetHistoryNavigation()
	updated := make([]rune, 0, len(m.input)+len(runes))
	updated = append(updated, m.input[:m.cursor]...)
	updated = append(updated, runes...)
	updated = append(updated, m.input[m.cursor:]...)
	m.input = updated
	m.cursor += len(runes)
	m.refilterCompletion()
}

func (m *model) refilterCompletion() {
	if !m.completionActive {
		return
	}
	m.completionItems = FilterComponents(m.catalog, string(m.input), m.cursor)
	if len(m.completionItems) == 0 {
		m.completionActive = false
		m.completionIndex = 0
	} else if m.completionIndex >= len(m.completionItems) {
		m.completionIndex = 0
	}
}

func (m *model) applyCompletion() {
	if !m.completionActive || len(m.completionItems) == 0 {
		m.completionActive = false
		return
	}

	if m.completionIndex < 0 || m.completionIndex >= len(m.completionItems) {
		m.completionIndex = 0
	}
	comp := m.completionItems[m.completionIndex]

	token := ExtractTargetToken(string(m.input), m.cursor)
	tokenRunesLen := len([]rune(token))

	tokenStart := m.cursor - tokenRunesLen
	if tokenStart < 0 {
		tokenStart = 0
	}

	prefix := make([]rune, tokenStart)
	copy(prefix, m.input[:tokenStart])

	templateRunes := []rune(comp.Template)

	suffix := make([]rune, len(m.input)-m.cursor)
	copy(suffix, m.input[m.cursor:])

	updated := make([]rune, 0, len(prefix)+len(templateRunes)+len(suffix))
	updated = append(updated, prefix...)
	updated = append(updated, templateRunes...)
	updated = append(updated, suffix...)

	m.input = updated
	m.cursor = tokenStart + len(templateRunes)
	m.completionActive = false
}

func (m *model) rememberExpression(expression string) {
	if expression == "" {
		return
	}
	if len(m.historyInputs) > 0 && m.historyInputs[0] == expression {
		return
	}
	m.historyInputs = append([]string{expression}, m.historyInputs...)
}

func (m *model) navigateHistory(direction int) {
	if len(m.historyInputs) == 0 {
		return
	}
	if direction < 0 {
		if m.historyIndex == -1 {
			m.historyDraft = string(m.input)
			m.historyIndex = 0
		} else if m.historyIndex < len(m.historyInputs)-1 {
			m.historyIndex++
		}
	} else if direction > 0 {
		if m.historyIndex == -1 {
			return
		}
		if m.historyIndex == 0 {
			m.historyIndex = -1
		} else {
			m.historyIndex--
		}
	}

	if m.historyIndex == -1 {
		m.input = []rune(m.historyDraft)
	} else {
		m.input = []rune(m.historyInputs[m.historyIndex])
	}
	m.cursor = len(m.input)
}

func (m *model) resetHistoryNavigation() {
	m.historyIndex = -1
	m.historyDraft = ""
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.View{}
	}

	var builder strings.Builder
	builder.WriteString("Dice rolling REPL\n")
	builder.WriteString("Enter a dice expression and press Enter.\n")
	builder.WriteString("Keys: q/esc/ctrl+c quit, up/down browse history, ctrl+v paste clipboard, left/right move, backspace/delete edit, ctrl+u clear.\n\n")
	builder.WriteString(m.renderInput())
	builder.WriteString(m.renderCompletion())
	builder.WriteString("\n\nRecent rolls:\n")

	if len(m.history) == 0 {
		builder.WriteString("  (none yet)\n")
		view := tea.NewView(builder.String())
		view.AltScreen = true
		return view
	}

	for _, entry := range m.history[:min(len(m.history), visibleHistory)] {
		if entry.failed {
			builder.WriteString(fmt.Sprintf("  [err] %s -> %s\n", entry.expression, entry.output))
			continue
		}
		builder.WriteString(fmt.Sprintf("  [ok]  %s\n", entry.output))
	}

	view := tea.NewView(builder.String())
	view.AltScreen = true
	return view
}

func (m model) renderCompletion() string {
	if !m.completionActive || len(m.completionItems) == 0 {
		return ""
	}

	if m.completionIndex < 0 || m.completionIndex >= len(m.completionItems) {
		m.completionIndex = 0
	}

	var builder strings.Builder
	builder.WriteString("\nCompletion options (Tab/Down select, Enter insert, Esc cancel):\n")

	maxVisible := 5
	start := 0
	if m.completionIndex >= maxVisible {
		start = m.completionIndex - maxVisible + 1
	}
	end := min(start+maxVisible, len(m.completionItems))

	for i := start; i < end; i++ {
		item := m.completionItems[i]
		prefix := "  "
		if i == m.completionIndex {
			prefix = "> "
		}
		builder.WriteString(fmt.Sprintf("%s[%-3s] %-25s (%s)\n", prefix, item.Syntax, item.Name, item.Category))
	}

	selected := m.completionItems[m.completionIndex]
	innerWidth := 64
	topPad := max(0, innerWidth-9-utf8.RuneCountInString(selected.Name))
	builder.WriteString("\n┌─ Help: " + selected.Name + " " + strings.Repeat("─", topPad) + "┐\n")

	valWidth := innerWidth - 2 - 13
	builder.WriteString("│ Syntax:      " + padTo(selected.Syntax, valWidth) + " │\n")
	builder.WriteString("│ Description: " + padTo(selected.Description, valWidth) + " │\n")
	builder.WriteString("│ Example:     " + padTo(selected.Example, valWidth) + " │\n")
	builder.WriteString("└" + strings.Repeat("─", innerWidth) + "┘\n")

	return builder.String()
}

func padTo(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}


func (m model) renderInput() string {
	var builder strings.Builder
	builder.WriteString("roll> ")

	if len(m.input) == 0 {
		builder.WriteRune('█')
		builder.WriteString(" try 3d6+2 or {d6, d8}")
		return builder.String()
	}

	if m.cursor > len(m.input) {
		m.cursor = len(m.input)
	}

	builder.WriteString(string(m.input[:m.cursor]))
	builder.WriteRune('█')
	builder.WriteString(string(m.input[m.cursor:]))
	return builder.String()
}
