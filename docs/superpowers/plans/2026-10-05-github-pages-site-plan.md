# GitHub Pages Site with WebAssembly Dice Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a zero-external-dependency static GitHub Pages site reminiscent of Google's homepage in `docs/`, powered by an in-browser WebAssembly dice engine compiled from the `roll` library with d20 red/white styling, dark mode support, and an interactive grammar help modal.

**Architecture:** A standalone Go CLI package `cmd/wasm` separates pure evaluation logic (`engine.go`) from JS bindings (`main.go`), allowing unit testing via `go test`. The static site (`docs/index.html`) loads Go's `wasm_exec.js` and `roll.wasm`, providing a Google-style single-input interface, dice breakdown cards, and a native `<dialog>` help reference. Build automation and tool pinning is managed via `mise.toml`.

**Tech Stack:** Go 1.27.1, WebAssembly (`GOOS=js GOARCH=wasm`), `syscall/js`, vanilla HTML5 (`<dialog>`, CSS Grid/Flexbox), CSS Custom Properties with `@media (prefers-color-scheme: dark)`, and `mise`.

---

### Task 1: WASM Engine Logic and Unit Tests

**Files:**
- Create: `cmd/wasm/engine.go`
- Create: `cmd/wasm/engine_test.go`

- [ ] **Step 1: Write the failing unit tests for `cmd/wasm/engine.go`**

```go
package main

import (
	"testing"
)

func TestEvaluateRoll_Success(t *testing.T) {
	resp := EvaluateRoll("3d6+2")
	if !resp.OK {
		t.Fatalf("expected OK=true, got error: %s", resp.Error)
	}
	if resp.Input != "3d6+2" {
		t.Errorf("expected input '3d6+2', got '%s'", resp.Input)
	}
	if resp.Rendered != "3d6+2" {
		t.Errorf("expected rendered '3d6+2', got '%s'", resp.Rendered)
	}
	if len(resp.Results) != 3 {
		t.Fatalf("expected 3 dice results, got %d", len(resp.Results))
	}
	if resp.Total < 5 || resp.Total > 20 {
		t.Errorf("expected total between 5 and 20, got %d", resp.Total)
	}
	if resp.Text == "" {
		t.Errorf("expected non-empty text representation")
	}
}

func TestEvaluateRoll_FateDie(t *testing.T) {
	resp := EvaluateRoll("4dF")
	if !resp.OK {
		t.Fatalf("expected OK=true, got error: %s", resp.Error)
	}
	if len(resp.Results) != 4 {
		t.Fatalf("expected 4 dice results, got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if r.Symbol != "⊞" && r.Symbol != "⊟" && r.Symbol != "☐" {
			t.Errorf("unexpected fate symbol: %q", r.Symbol)
		}
	}
}

func TestEvaluateRoll_Error(t *testing.T) {
	resp := EvaluateRoll("invalid_dice")
	if resp.OK {
		t.Fatalf("expected OK=false for invalid syntax")
	}
	if resp.Error == "" {
		t.Errorf("expected descriptive error message")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./cmd/wasm`  
Expected: Compilation failure / `EvaluateRoll` undefined.

- [ ] **Step 3: Implement `cmd/wasm/engine.go`**

```go
package main

import (
	"github.com/darkliquid/roll"
)

// DieResult contains the individual die evaluation outcome.
type DieResult struct {
	Result int    `json:"result"`
	Symbol string `json:"symbol"`
}

// RollResponse represents the structured evaluation result sent to the browser.
type RollResponse struct {
	OK           bool        `json:"ok"`
	Input        string      `json:"input,omitempty"`
	Rendered     string      `json:"rendered,omitempty"`
	Total        int         `json:"total,omitempty"`
	Successes    int         `json:"successes,omitempty"`
	HasSuccesses bool        `json:"hasSuccesses,omitempty"`
	Results      []DieResult `json:"results,omitempty"`
	Text         string      `json:"text,omitempty"`
	Error        string      `json:"error,omitempty"`
}

// EvaluateRoll parses and evaluates a dice expression string.
func EvaluateRoll(expr string) RollResponse {
	program, err := roll.CompileString(expr)
	if err != nil {
		return RollResponse{
			OK:    false,
			Error: err.Error(),
		}
	}

	result, err := roll.EvaluateProgram(program)
	if err != nil {
		return RollResponse{
			OK:    false,
			Error: err.Error(),
		}
	}

	text, _ := roll.ParseString(expr)

	var diceResults []DieResult
	for _, r := range result.Results {
		diceResults = append(diceResults, DieResult{
			Result: r.Result,
			Symbol: r.Symbol,
		})
	}

	hasSuccesses := false
	for _, term := range program.DiceTerms {
		if term.Success != nil || term.Failure != nil {
			hasSuccesses = true
			break
		}
	}
	if !hasSuccesses {
		for _, term := range program.GroupTerms {
			if term.Success != nil || term.Failure != nil {
				hasSuccesses = true
				break
			}
		}
	}

	return RollResponse{
		OK:           true,
		Input:        expr,
		Rendered:     program.String(),
		Total:        result.Total,
		Successes:    result.Successes,
		HasSuccesses: hasSuccesses,
		Results:      diceResults,
		Text:         text,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./cmd/wasm`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/wasm/engine.go cmd/wasm/engine_test.go
git commit -m "feat(wasm): add dice evaluation engine and unit tests"
```

---

### Task 2: WebAssembly Entrypoint (`cmd/wasm/main.go`) and `mise.toml` Automation

**Files:**
- Create: `cmd/wasm/main.go`
- Create: `mise.toml`

- [ ] **Step 1: Write `cmd/wasm/main.go` with JS bridge**

```go
//go:build js && wasm

package main

import (
	"syscall/js"
)

func rollDiceWrapper(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{
			"ok":    false,
			"error": "missing roll expression",
		}
	}

	expr := args[0].String()
	resp := EvaluateRoll(expr)

	if !resp.OK {
		return map[string]any{
			"ok":    false,
			"error": resp.Error,
		}
	}

	resultsSlice := make([]any, len(resp.Results))
	for i, r := range resp.Results {
		resultsSlice[i] = map[string]any{
			"result": r.Result,
			"symbol": r.Symbol,
		}
	}

	return map[string]any{
		"ok":           true,
		"input":        resp.Input,
		"rendered":     resp.Rendered,
		"total":        resp.Total,
		"successes":    resp.Successes,
		"hasSuccesses": resp.HasSuccesses,
		"results":      resultsSlice,
		"text":         resp.Text,
	}
}

func main() {
	js.Global().Set("rollDice", js.FuncOf(rollDiceWrapper))
	// Prevent Go program from exiting so JS callbacks remain active
	select {}
}
```

- [ ] **Step 2: Create `mise.toml`**

```toml
[tools]
go = "1.27.1"

[tasks.build-wasm]
description = "Build WebAssembly binary and copy wasm_exec.js for GitHub Pages"
run = """
set -euo pipefail
mkdir -p docs
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o docs/roll.wasm ./cmd/wasm
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/lib/wasm/wasm_exec.js" docs/wasm_exec.js
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/misc/wasm/wasm_exec.js" docs/wasm_exec.js
else
  echo "Error: wasm_exec.js not found in GOROOT ($GOROOT)" >&2
  exit 1
fi
echo "Successfully built docs/roll.wasm and synced docs/wasm_exec.js"
"""
```

- [ ] **Step 3: Run `mise run build-wasm` and verify output artifacts**

Run: `mise run build-wasm`  
Expected output:
- `docs/roll.wasm` created (size > 1MB)
- `docs/wasm_exec.js` copied
- Zero build errors

- [ ] **Step 4: Commit**

```bash
git add cmd/wasm/main.go mise.toml docs/roll.wasm docs/wasm_exec.js
git commit -m "feat(wasm): add browser wasm entrypoint and mise build task"
```

---

### Task 3: Static Webpage (`docs/index.html`)

**Files:**
- Create: `docs/index.html`

- [ ] **Step 1: Write `docs/index.html`**
Create `docs/index.html` containing:
- Semantic HTML5 structure with Google homepage aesthetics.
- Embedded CSS variables for light and dark modes (`@media (prefers-color-scheme: dark)`).
- D20 crimson red palette (`#c5221f` light / `#e04845` dark), white contrast numerals, and sleek neutral cards.
- Centered icosahedron SVG logo with "Roll".
- Search input pill with autofocus and <kbd>Enter</kbd> key submission.
- Two action buttons: "Roll" and "I'm feeling lucky".
- Curated lucky pool covering varied notations (`4d6kh3`, `1d20+7`, `2d20kl1`, `3d6!`, `4d6!!`, `2d8!p+2`, `4dF+1`, `1d%`, `5d10>7`, `6d6>4f1`, `4d6ro1`, `6d6sd`, `{4d6, 2d8}>4`, `8d6`).
- Dynamic result section below the buttons showing:
  - Grand total badge
  - Rendered formula
  - Individual dice chips (with proper handling for Fate dice symbols `⊞`, `⊟`, `☐`)
  - Success/failure callouts when applicable
  - Clear error notice when parsing fails
- Native `<dialog id="help-modal">`:
  - Categories: Basic Dice, Exploding, Keep & Drop, Rerolls, Targets & Failures, Sorting, Grouping.
  - Interactive clickable chips that populate the input, close the dialog, and roll immediately.
  - Link to GitHub repository `https://github.com/darkliquid/roll`.
- Top-bar links: Only "Grammar & Help" and the GitHub repo link (with inline Octocat SVG).
- Footer links: Only the GitHub repository link.
- Embedded WebAssembly loader using `wasm_exec.js` and `WebAssembly.instantiateStreaming` with fallback to `ArrayBuffer`.

- [ ] **Step 2: Commit**

```bash
git add docs/index.html
git commit -m "feat(web): add Google-style d20 GitHub Pages interface"
```

---

### Task 4: Verification, Browser Testing & Documentation Update

**Files:**
- Modify: `README.md`
- Verify: Full test suite and browser execution

- [ ] **Step 1: Run Go test suite across the entire repository**

Run: `go test -v ./...`  
Expected: All tests pass.

- [ ] **Step 2: Start a local test server and verify browser behavior**

Run: A local background HTTP server (e.g. `python3 -m http.server 8085 --directory docs`)  
Use Chrome DevTools / headless evaluation to verify:
1. WASM initializes cleanly with `window.rollDice`.
2. Enter `4d6kh3` and click Roll: result card appears with 3 dice chips, total, and rendered formula.
3. Click "I'm feeling lucky": input fills with a random dice expression and rolls immediately.
4. Enter an invalid roll `abc` and click Roll: error alert shows syntax error message.
5. Click "Grammar & Help": modal opens, click an interactive example (e.g. `4dF`), modal closes and rolls Fate dice.
6. Verify Dark Mode toggle via Chrome DevTools emulation (`emulate({ colorScheme: 'dark' })`).

- [ ] **Step 3: Update `README.md`**
Add a short section to `README.md` documenting the live GitHub Pages site (`https://darkliquid.github.io/roll/`) and how to build the wasm site via `mise run build-wasm`.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document GitHub Pages site and mise build task"
```
