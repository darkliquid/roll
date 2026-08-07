# REPL Auto-Completion & Component Help Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the `roll` CLI REPL with interactive auto-completion, component selection, and syntax help documentation for all dice rolling modifiers.

**Architecture:** Create a modular component catalog (`cmd/repl/catalog.go`) containing metadata, descriptions, and examples for all supported `roll` syntax options. Extend Bubble Tea `model` in `cmd/repl/model.go` to support prompt prefix matching, popup completion menu, help text box, and keybinding handlers (`Tab`, `Down`, `Up`, `Enter`, `Esc`).

**Tech Stack:** Go 1.22+, Bubble Tea (`charm.land/bubbletea/v2`), `github.com/darkliquid/roll`

---

### Task 1: Component Catalog & Prefix Filtering (`cmd/repl/catalog.go`)

**Files:**
- Create: `cmd/repl/catalog.go`
- Create: `cmd/repl/catalog_test.go`

- [ ] **Step 1: Write failing catalog tests**

```go
package main

import (
	"testing"
)

func TestDefaultCatalogContainsSupportedComponents(t *testing.T) {
	cat := DefaultCatalog()
	if len(cat) == 0 {
		t.Fatal("expected non-empty default catalog")
	}

	foundExploding := false
	for _, c := range cat {
		if c.Syntax == "!" {
			foundExploding = true
			if c.Name != "Exploding Dice" {
				t.Errorf("unexpected name for !: %q", c.Name)
			}
			if c.Description == "" || c.Example == "" {
				t.Errorf("missing description or example for !: %#v", c)
			}
		}
	}
	if !foundExploding {
		t.Fatal("expected exploding dice component in default catalog")
	}
}

func TestFilterComponentsByPrefix(t *testing.T) {
	cat := DefaultCatalog()

	// Empty input matches all items
	items := FilterComponents(cat, "", 0)
	if len(items) != len(cat) {
		t.Fatalf("expected all %d items on empty input, got %d", len(cat), len(items))
	}

	// Prefix "k" matches kh and kl
	items = FilterComponents(cat, "3d6k", 4)
	if len(items) < 2 {
		t.Fatalf("expected at least 2 keep components for prefix 'k', got %d", len(items))
	}
	for _, item := range items {
		if item.Syntax != "kh" && item.Syntax != "kl" {
			t.Errorf("unexpected filtered item for 'k': %s", item.Syntax)
		}
	}

	// Prefix "!" matches !, !!, !p
	items = FilterComponents(cat, "4d6!", 4)
	if len(items) < 3 {
		t.Fatalf("expected at least 3 exploding components for prefix '!', got %d", len(items))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/repl -v -run "TestDefaultCatalog|TestFilterComponents"`
Expected: FAIL due to undefined `DefaultCatalog` and `FilterComponents`.

- [ ] **Step 3: Implement catalog and filtering logic**

Create `cmd/repl/catalog.go`:

```go
package main

import (
	"strings"
	"unicode"
)

type Component struct {
	Name        string
	Syntax      string
	Template    string
	Category    string
	Description string
	Example     string
}

func DefaultCatalog() []Component {
	return []Component{
		{
			Name:        "Standard Die",
			Syntax:      "d",
			Template:    "d6",
			Category:    "Dice",
			Description: "Rolls N dice with S sides (e.g., 3d6 rolls three 6-sided dice).",
			Example:     "3d6",
		},
		{
			Name:        "Percentile Die",
			Syntax:      "d%",
			Template:    "d%",
			Category:    "Dice",
			Description: "Rolls a 100-sided percentile die.",
			Example:     "d%",
		},
		{
			Name:        "Fate / Fudge Die",
			Syntax:      "dF",
			Template:    "dF",
			Category:    "Dice",
			Description: "Rolls Fate/Fudge dice with +, -, and blank sides (-1, 0, +1).",
			Example:     "4dF",
		},
		{
			Name:        "Exploding Dice",
			Syntax:      "!",
			Template:    "!",
			Category:    "Exploding",
			Description: "Re-rolls a die whenever it lands on maximum value and adds it to total.",
			Example:     "3d6!",
		},
		{
			Name:        "Compounded Exploding Dice",
			Syntax:      "!!",
			Template:    "!!",
			Category:    "Exploding",
			Description: "Re-rolls maximum values and accumulates results onto a single die.",
			Example:     "3d6!!",
		},
		{
			Name:        "Penetrating Exploding Dice",
			Syntax:      "!p",
			Template:    "!p",
			Category:    "Exploding",
			Description: "Explodes max rolls, but subtracts 1 from each additional rolled die.",
			Example:     "3d6!p",
		},
		{
			Name:        "Keep Highest",
			Syntax:      "kh",
			Template:    "kh1",
			Category:    "Keep/Drop",
			Description: "Keeps only the highest N dice from the roll.",
			Example:     "4d6kh3",
		},
		{
			Name:        "Keep Lowest",
			Syntax:      "kl",
			Template:    "kl1",
			Category:    "Keep/Drop",
			Description: "Keeps only the lowest N dice from the roll.",
			Example:     "4d6kl1",
		},
		{
			Name:        "Drop Highest",
			Syntax:      "dh",
			Template:    "dh1",
			Category:    "Keep/Drop",
			Description: "Drops the highest N dice from the roll.",
			Example:     "4d6dh1",
		},
		{
			Name:        "Drop Lowest",
			Syntax:      "dl",
			Template:    "dl1",
			Category:    "Keep/Drop",
			Description: "Drops the lowest N dice from the roll.",
			Example:     "4d6dl1",
		},
		{
			Name:        "Reroll Once",
			Syntax:      "ro",
			Template:    "ro<2",
			Category:    "Reroll",
			Description: "Rerolls dice matching condition once.",
			Example:     "2d20ro<2",
		},
		{
			Name:        "Reroll Recursive",
			Syntax:      "r",
			Template:    "r<2",
			Category:    "Reroll",
			Description: "Continuously rerolls dice matching condition until condition fails.",
			Example:     "3d6r1",
		},
		{
			Name:        "Target Greater Than",
			Syntax:      ">",
			Template:    ">5",
			Category:    "Target/Success",
			Description: "Counts dice meeting or exceeding target number as successes.",
			Example:     "5d6>4",
		},
		{
			Name:        "Target Less Than",
			Syntax:      "<",
			Template:    "<2",
			Category:    "Target/Success",
			Description: "Counts dice meeting or under target number as successes.",
			Example:     "5d6<3",
		},
		{
			Name:        "Target Equal To",
			Syntax:      "=",
			Template:    "=6",
			Category:    "Target/Success",
			Description: "Counts dice equal to target value as successes.",
			Example:     "5d6=6",
		},
		{
			Name:        "Failure Threshold",
			Syntax:      "f",
			Template:    "f1",
			Category:    "Target/Success",
			Description: "Subtracts 1 success for dice landing on or below failure threshold.",
			Example:     "5d6>4f1",
		},
		{
			Name:        "Sort Ascending",
			Syntax:      "s",
			Template:    "s",
			Category:    "Sorting",
			Description: "Sorts dice results in ascending order.",
			Example:     "4d6s",
		},
		{
			Name:        "Sort Descending",
			Syntax:      "sd",
			Template:    "sd",
			Category:    "Sorting",
			Description: "Sorts dice results in descending order.",
			Example:     "4d6sd",
		},
		{
			Name:        "Grouped Roll",
			Syntax:      "{}",
			Template:    "{d6, d8}",
			Category:    "Grouping",
			Description: "Groups multiple sub-rolls separated by commas.",
			Example:     "{2d6+2, 1d8}",
		},
	}
}

func ExtractTargetToken(input string, cursor int) string {
	if cursor > len(input) {
		cursor = len(input)
	}
	runes := []rune(input[:cursor])
	if len(runes) == 0 {
		return ""
	}

	start := len(runes) - 1
	for start >= 0 {
		ch := runes[start]
		if unicode.IsSpace(ch) || ch == '+' || ch == '-' || ch == ',' {
			start++
			break
		}
		if start == 0 {
			break
		}
		start--
	}
	if start < 0 {
		start = 0
	}
	return string(runes[start:])
}

func FilterComponents(catalog []Component, input string, cursor int) []Component {
	token := ExtractTargetToken(input, cursor)
	if token == "" {
		return catalog
	}

	lowerToken := strings.ToLower(token)
	var matched []Component
	for _, item := range catalog {
		if strings.HasPrefix(strings.ToLower(item.Syntax), lowerToken) ||
			strings.HasPrefix(strings.ToLower(item.Name), lowerToken) ||
			strings.HasPrefix(strings.ToLower(item.Template), lowerToken) {
			matched = append(matched, item)
		}
	}

	if len(matched) == 0 {
		return catalog
	}
	return matched
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/repl -v -run "TestDefaultCatalog|TestFilterComponents"`
Expected: PASS.

- [ ] **Step 5: Commit catalog module**

```bash
git add cmd/repl/catalog.go cmd/repl/catalog_test.go
git commit -m "feat(repl): add syntax component catalog and prefix filtering"
```

---

### Task 2: Autocomplete Model Integration & Keybindings (`cmd/repl/model.go`)

**Files:**
- Modify: `cmd/repl/model.go`
- Modify: `cmd/repl/model_test.go`

- [ ] **Step 1: Write failing autocomplete integration tests**

Add to `cmd/repl/model_test.go`:

```go
func TestAutocompleteTabTogglesAndNavigates(t *testing.T) {
	m := newModel()
	m.input = []rune("3d6!")
	m.cursor = len(m.input)

	// Pressing Tab activates completion
	updatedModel, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	updated := updatedModel.(model)
	if !updated.completionActive {
		t.Fatal("expected completion to be active after Tab")
	}
	if len(updated.completionItems) == 0 {
		t.Fatal("expected matching completion items")
	}
	if updated.completionIndex != 0 {
		t.Fatalf("expected completion index 0, got %d", updated.completionIndex)
	}

	// Pressing Tab again cycles completion index
	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	updated = updatedModel.(model)
	if updated.completionIndex != 1 {
		t.Fatalf("expected completion index 1 after second Tab, got %d", updated.completionIndex)
	}

	// Pressing Esc closes completion
	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	updated = updatedModel.(model)
	if updated.completionActive {
		t.Fatal("expected completion to close on Esc")
	}
}

func TestAutocompleteEnterInsertsComponent(t *testing.T) {
	m := newModel()
	m.input = []rune("4d6k")
	m.cursor = len(m.input)

	// Activate completion
	updatedModel, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	updated := updatedModel.(model)

	// Press Enter to insert selected component template ("kh1")
	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated = updatedModel.(model)

	if updated.completionActive {
		t.Fatal("expected completion to close after selection")
	}
	if string(updated.input) != "4d6kh1" {
		t.Fatalf("expected inserted template '4d6kh1', got %q", string(updated.input))
	}
	if updated.cursor != len(updated.input) {
		t.Fatalf("expected cursor at end of inserted text, got %d", updated.cursor)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/repl -v -run "TestAutocomplete"`
Expected: FAIL due to missing completion fields & key handling logic in `model.go`.

- [ ] **Step 3: Implement autocomplete fields and key handlers in `model.go`**

Update `cmd/repl/model.go`:
1. Add `catalog`, `completionActive`, `completionItems`, `completionIndex` to `model` struct.
2. Initialize `catalog: DefaultCatalog()` in `newModel()`.
3. In `Update()` keypress handler:
   - `tab` / `ctrl+space`: toggle completion or advance `completionIndex`.
   - `esc`: if `completionActive`, set `completionActive = false`; else quit.
   - `down`: if `completionActive`, advance `completionIndex`; else navigate history.
   - `up`: if `completionActive`, decrement `completionIndex`; else navigate history.
   - `enter`: if `completionActive`, apply completion; else submit roll.
   - On character insertion (`insertRunes`, `backspace`, `delete`), recalculate `completionItems` if active.
4. Implement `applyCompletion()` helper to replace target token with `Component.Template`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/repl -v -run "TestAutocomplete"`
Expected: PASS.

- [ ] **Step 5: Commit autocomplete keybinding integration**

```bash
git add cmd/repl/model.go cmd/repl/model_test.go
git commit -m "feat(repl): integrate autocomplete menu navigation and template insertion"
```

---

### Task 3: Render Autocomplete Dropdown & Syntax Help UI (`cmd/repl/model.go`)

**Files:**
- Modify: `cmd/repl/model.go`
- Modify: `cmd/repl/model_test.go`

- [ ] **Step 1: Write failing UI render test**

Add to `cmd/repl/model_test.go`:

```go
func TestAutocompleteViewRendersHelpBox(t *testing.T) {
	m := newModel()
	m.input = []rune("3d6!")
	m.cursor = len(m.input)
	m.completionActive = true
	m.completionItems = FilterComponents(m.catalog, "3d6!", 4)
	m.completionIndex = 0

	viewOutput := m.View().Content
	if viewOutput == "" {
		t.Fatal("expected view output")
	}

	if !strings.Contains(viewOutput, "Completion options") {
		t.Errorf("expected view to contain completion header: %s", viewOutput)
	}
	if !strings.Contains(viewOutput, "Help:") {
		t.Errorf("expected view to contain Help box header: %s", viewOutput)
	}
	if !strings.Contains(viewOutput, "Exploding Dice") {
		t.Errorf("expected view to contain component name: %s", viewOutput)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/repl -v -run "TestAutocompleteViewRendersHelpBox"`
Expected: FAIL because `View()` does not yet render completion menu and help box.

- [ ] **Step 3: Implement `renderCompletion()` in `model.go`**

Add helper methods to `model.go`:
```go
func (m model) renderCompletion() string {
	if !m.completionActive || len(m.completionItems) == 0 {
		return ""
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
	builder.WriteString("\n┌─ Help: " + selected.Name + " " + strings.Repeat("─", max(2, 50-len(selected.Name))) + "┐\n")
	builder.WriteString(fmt.Sprintf("│ Syntax:      %-50s │\n", selected.Syntax))
	builder.WriteString(fmt.Sprintf("│ Description: %-50s │\n", selected.Description))
	builder.WriteString(fmt.Sprintf("│ Example:     %-50s │\n", selected.Example))
	builder.WriteString("└" + strings.Repeat("─", 66) + "┘\n")

	return builder.String()
}
```

Include `m.renderCompletion()` inside `View()`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/repl -v`
Expected: PASS.

- [ ] **Step 5: Commit UI rendering changes**

```bash
git add cmd/repl/model.go cmd/repl/model_test.go
git commit -m "feat(repl): render autocomplete dropdown menu and syntax help box"
```

---

### Task 4: End-to-End Verification & Manual Sanity Check

- [ ] **Step 1: Run all test suites**

Run: `go test ./... -v`
Expected: ALL PASS.

- [ ] **Step 2: Build REPL binary**

Run: `go build -o /tmp/repl ./cmd/repl`
Expected: Clean build, exit code 0.

- [ ] **Step 3: Commit completion**

```bash
git add .
git commit -m "chore: complete REPL autocomplete & help documentation feature"
```
