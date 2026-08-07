# Design Specification: REPL Auto-Completion & Component Help

Date: 2026-08-07  
Status: Approved  

## Overview
This specification details the addition of auto-completion, component selection, and syntax help documentation to the `roll` REPL terminal interface (`cmd/repl`).

## Objectives
1. Provide interactive auto-complete suggestions as users type or trigger via hotkeys (`Tab` / `Ctrl+Space`).
2. Offer a comprehensive catalog of all supported `roll` syntax components (dice types, explosion modifiers, keep/drop limits, rerolls, target successes, sorting, grouping).
3. Display clear help text, descriptions, and examples for each syntax option alongside the completion dropdown.
4. Maintain clean keyboard navigation separating prompt editing, history navigation, and completion selection.

## Component Catalog Architecture (`cmd/repl/catalog.go`)

### Data Structure
```go
package main

type Component struct {
    Name        string // Human-readable component name (e.g. "Exploding Dice")
    Syntax      string // Syntax prefix/token (e.g. "!")
    Template    string // Text inserted into prompt upon selection
    Category    string // Category ("Dice", "Exploding", "Keep/Drop", "Reroll", "Target/Success", "Sort", "Grouping")
    Description string // Explanation of what the syntax component does
    Example     string // Usage example (e.g. "3d6! - roll 3d6, explode on 6")
}
```

### Catalog Items
- **Dice**: Standard (`d6`), Percentile (`d%`), Fate/Fudge (`dF`)
- **Exploding**: Exploding (`!`), Compounded (`!!`), Penetrating (`!p`)
- **Keep/Drop**: Keep High (`kh`), Keep Low (`kl`), Drop High (`dh`), Drop Low (`dl`)
- **Reroll**: Reroll Recursive (`r`), Reroll Once (`ro`)
- **Target/Success**: Greater Than (`>`), Less Than (`<`), Equal (`=`), Failures (`f`)
- **Sorting**: Ascending (`s`), Descending (`sd`)
- **Grouping & Math**: Grouped Roll (`{d6, d8}`), Basic Operators (`+`, `-`)

## Model Extensions (`cmd/repl/model.go`)

### Model Fields
```go
type model struct {
    // Existing fields
    input          []rune
    cursor         int
    history        []historyEntry
    historyInputs  []string
    historyIndex   int
    historyDraft   string
    quitting       bool
    evaluator      func(string) (string, error)
    clipboardReady func() tea.Msg

    // Autocomplete fields
    catalog          []Component
    completionActive bool
    completionItems  []Component
    completionIndex  int
}
```

### Completion Logic & Filtering
- **Activation**:
  - Automatically triggered when typing syntax trigger characters (`d`, `!`, `k`, `r`, `s`, `>`, `<`, `{`, etc.) or when pressing `Tab` / `Ctrl+Space`.
  - Can be dismissed with `Esc`.
- **Filtering**:
  - Determines current token word around/before cursor.
  - Matches token against `Component.Syntax`, `Component.Name`, and `Component.Category`.
  - When input is blank or after operator, displays all catalog items.
- **Insertion**:
  - Replaces current target token at cursor with `Component.Template`.
  - Cursor is positioned after the inserted template.

## Terminal UI & Layout

The REPL output will dynamically render completion menu and syntax help box below the prompt:

```text
Dice rolling REPL
Enter a dice expression and press Enter.
Keys: Tab/ctrl+space complete, up/down history/navigate, q/esc/ctrl+c quit.

roll> 3d6!█

Completion options (Tab/Down select, Enter insert, Esc cancel):
  [!]  Exploding Dice  (Re-rolls die on max value and adds to roll)
> [!!] Compounded Exploding Dice
  [!p] Penetrating Exploding Dice

┌─ Help: Compounded Exploding Dice ────────────────────────────────────┐
│ Syntax: !!                                                           │
│ Category: Exploding                                                  │
│ Description: Re-rolls maximum values and accumulates onto single die.│
│ Example: 3d6!!                                                       │
└──────────────────────────────────────────────────────────────────────┘

Recent rolls:
  (none yet)
```

## Keybindings Matrix

| Key | Completion Active | Completion Inactive |
| :--- | :--- | :--- |
| `Tab` | Next completion item (cycle) | Open completion menu |
| `Shift+Tab` / `Up` | Previous completion item | History forward (`Ctrl+P`) |
| `Down` | Next completion item | History backward (`Ctrl+N`) |
| `Enter` | Insert selected component | Submit roll expression |
| `Esc` | Close completion menu | Quit REPL |
| `Ctrl+Space` | Toggle completion menu | Toggle completion menu |

## Testing Plan (`cmd/repl/model_test.go`)
- Unit tests for completion catalog filtering by prefix.
- Integration tests for key handling (`Tab` activation, `Down`/`Up` navigation, `Enter` insertion).
- Verification of prompt cursor positioning after template insertion.
- Edge case tests: empty input, token replacement at start/middle/end of string.
