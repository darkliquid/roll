package main

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestModelSubmitSuccess(t *testing.T) {
	m := newModel()
	m.evaluator = func(expression string) (string, error) {
		if expression != "3d6+2" {
			t.Fatalf("unexpected expression: %q", expression)
		}
		return `Rolled "3d6+2" and got 1, 2, 3 for a total of 8`, nil
	}
	m.input = []rune("3d6+2")
	m.cursor = len(m.input)

	updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := updatedModel.(model)
	if cmd == nil {
		t.Fatal("expected submit command")
	}
	if string(updated.input) != "" {
		t.Fatalf("expected cleared input, got %q", string(updated.input))
	}
	if updated.cursor != 0 {
		t.Fatalf("expected cursor reset, got %d", updated.cursor)
	}

	msg := cmd()
	updatedModel, _ = updated.Update(msg)
	updated = updatedModel.(model)
	if len(updated.history) != 1 {
		t.Fatalf("expected one history entry, got %d", len(updated.history))
	}
	if updated.history[0].failed {
		t.Fatal("expected success entry")
	}
	if updated.history[0].output != `Rolled "3d6+2" and got 1, 2, 3 for a total of 8` {
		t.Fatalf("unexpected output: %q", updated.history[0].output)
	}
}

func TestModelSubmitError(t *testing.T) {
	m := newModel()
	m.evaluator = func(expression string) (string, error) {
		if expression != "oops" {
			t.Fatalf("unexpected expression: %q", expression)
		}
		return "", errors.New(`found unexpected token "o"`)
	}
	m.input = []rune("oops")
	m.cursor = len(m.input)

	updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := updatedModel.(model)
	msg := cmd()
	updatedModel, _ = updated.Update(msg)
	updated = updatedModel.(model)

	if len(updated.history) != 1 {
		t.Fatalf("expected one history entry, got %d", len(updated.history))
	}
	if !updated.history[0].failed {
		t.Fatal("expected error entry")
	}
	if updated.history[0].output != `found unexpected token "o"` {
		t.Fatalf("unexpected error output: %q", updated.history[0].output)
	}
}

func TestModelIgnoresBlankSubmit(t *testing.T) {
	m := newModel()
	m.input = []rune("   ")
	m.cursor = len(m.input)

	updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := updatedModel.(model)
	if cmd != nil {
		t.Fatal("expected no command for blank input")
	}
	if len(updated.history) != 0 {
		t.Fatalf("expected no history entries, got %d", len(updated.history))
	}
}

func TestModelHistoryNavigationRestoresDraft(t *testing.T) {
	m := newModel()
	m.historyInputs = []string{"4d6kh3", "3d6+2"}
	m.input = []rune("draft")
	m.cursor = len(m.input)
	m.historyIndex = -1

	updatedModel, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	updated := updatedModel.(model)
	if string(updated.input) != "4d6kh3" {
		t.Fatalf("expected latest history item, got %q", string(updated.input))
	}

	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	updated = updatedModel.(model)
	if string(updated.input) != "3d6+2" {
		t.Fatalf("expected older history item, got %q", string(updated.input))
	}

	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	updated = updatedModel.(model)
	if string(updated.input) != "4d6kh3" {
		t.Fatalf("expected newer history item, got %q", string(updated.input))
	}

	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	updated = updatedModel.(model)
	if string(updated.input) != "draft" {
		t.Fatalf("expected restored draft, got %q", string(updated.input))
	}
}

func TestModelPasteMsgInsertsAtCursor(t *testing.T) {
	m := newModel()
	m.input = []rune("3d6")
	m.cursor = 1

	updatedModel, _ := m.Update(tea.PasteMsg{Content: "00"})
	updated := updatedModel.(model)
	if string(updated.input) != "300d6" {
		t.Fatalf("unexpected input after paste: %q", string(updated.input))
	}
	if updated.cursor != 3 {
		t.Fatalf("unexpected cursor after paste: %d", updated.cursor)
	}
}

func TestModelClipboardShortcutReadsClipboard(t *testing.T) {
	m := newModel()
	m.clipboardReady = func() tea.Msg {
		return tea.ClipboardMsg{Content: "d20"}
	}

	updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected clipboard command")
	}
	updated := updatedModel.(model)
	msg := cmd()
	updatedModel, _ = updated.Update(msg)
	updated = updatedModel.(model)

	if string(updated.input) != "d20" {
		t.Fatalf("unexpected clipboard input: %q", string(updated.input))
	}
	if updated.cursor != 3 {
		t.Fatalf("unexpected cursor: %d", updated.cursor)
	}
}

func TestModelSubmitStoresHistoryWithoutImmediateDuplicate(t *testing.T) {
	m := newModel()
	m.evaluator = func(expression string) (string, error) {
		return expression, nil
	}
	m.input = []rune("2d8")
	m.cursor = len(m.input)

	updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := updatedModel.(model)
	if len(updated.historyInputs) != 1 || updated.historyInputs[0] != "2d8" {
		t.Fatalf("expected submitted expression in history inputs, got %#v", updated.historyInputs)
	}

	_, _ = updated.Update(cmd())

	updated.input = []rune("2d8")
	updated.cursor = len(updated.input)
	updatedModel, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated = updatedModel.(model)
	if len(updated.historyInputs) != 1 {
		t.Fatalf("expected duplicate submission to be coalesced, got %#v", updated.historyInputs)
	}
}
