# Design Spec: GitHub Pages Site with WebAssembly Dice Engine

**Date**: 2026-10-05  
**Status**: Approved  
**Target Files**:
- `cmd/wasm/main.go`
- `mise.toml`
- `docs/index.html`
- `docs/wasm_exec.js` (copied from Go runtime)
- `docs/roll.wasm` (compiled binary)

---

## 1. Objective

Create a standalone, zero-external-dependency GitHub Pages static site for `roll`, featuring:
1. A clean, minimalist layout reminiscent of the Google homepage (centered input pill, two action buttons: "Roll" and "I'm feeling lucky").
2. A distinct red-and-white d20 color scheme with full dark mode support using standard CSS media queries (`@media (prefers-color-scheme: dark)`).
3. In-browser dice execution via WebAssembly compiled from the existing Go `roll` library.
4. Dynamic result display beneath the buttons, replaced on each roll, showing grand totals, individual dice breakdown, and canonical notation.
5. An accessible `<dialog>` help modal documenting dice syntax and grammar with interactive auto-fill examples.
6. Links strictly limited to the GitHub repository and the help modal.
7. Build orchestration pinned and managed via `mise`.

---

## 2. WebAssembly Bridge (`cmd/wasm/main.go`)

### 2.1 API Contract
The WebAssembly module runs in the browser via Go's `syscall/js` runtime, exporting a global function:
`window.rollDice(expression: string) -> object`

#### Success Output Structure:
```json
{
  "ok": true,
  "input": "4d6kh3",
  "rendered": "4d6kh3",
  "total": 15,
  "successes": 0,
  "hasSuccesses": false,
  "results": [
    {"result": 5, "symbol": "5"},
    {"result": 6, "symbol": "6"},
    {"result": 4, "symbol": "4"}
  ],
  "text": "Rolled \"4d6kh3\" and got 5, 6, 4 for a total of 15"
}
```

#### Error Output Structure:
```json
{
  "ok": false,
  "error": "syntax error description"
}
```

### 2.2 Go Implementation Architecture
- Entry point in `cmd/wasm/main.go` using package `main`.
- Calls `roll.CompileString(input)` to compile and generate `program.String()`.
- Calls `roll.EvaluateProgram(program)` to obtain `roll.Result` (die rolls, total, successes).
- Calls `roll.ParseString(input)` to generate the canonical text summary.
- Converts structures to `map[string]any` and returns `js.ValueOf(...)`.
- Blocks runtime exit using `select {}` channel.

---

## 3. Tooling & Build (`mise.toml`)

`mise.toml` pins the Go compiler version and declares the build automation task:

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
"""
```

---

## 4. Webpage Design & Styling (`docs/index.html`)

### 4.1 Theme & CSS Variables
Standard CSS custom properties with automatic dark-mode switching:

- **Light Mode (Default)**:
  - `--bg`: `#ffffff`
  - `--surface`: `#f8f9fa`
  - `--border`: `#dadce0`
  - `--border-focus`: `#c5221f`
  - `--text`: `#202124`
  - `--text-muted`: `#5f6368`
  - `--d20-red`: `#c5221f`
  - `--d20-red-hover`: `#a31b18`
  - `--d20-badge-bg`: `#c5221f`
  - `--d20-badge-text`: `#ffffff`
  - `--btn-bg`: `#f8f9fa`
  - `--btn-border`: `#f8f9fa`
  - `--btn-text`: `#3c4043`
  - `--btn-hover-bg`: `#f1f3f4`
  - `--btn-hover-border`: `#dadce0`
  - `--chip-bg`: `#f1f3f4`
  - `--chip-text`: `#202124`
  - `--modal-backdrop`: `rgba(32, 33, 36, 0.6)`

- **Dark Mode (`@media (prefers-color-scheme: dark)`)**:
  - `--bg`: `#121214`
  - `--surface`: `#1e1f20`
  - `--border`: `#3c4043`
  - `--border-focus`: `#e04845`
  - `--text`: `#e8eaed`
  - `--text-muted`: `#9aa0a6`
  - `--d20-red`: `#e04845`
  - `--d20-red-hover`: `#f2615e`
  - `--d20-badge-bg`: `#c5221f`
  - `--d20-badge-text`: `#ffffff`
  - `--btn-bg`: `#282a2d`
  - `--btn-border`: `#282a2d`
  - `--btn-text`: `#e8eaed`
  - `--btn-hover-bg`: `#35373b`
  - `--btn-hover-border`: `#5f6368`
  - `--chip-bg`: `#282a2d`
  - `--chip-text`: `#e8eaed`
  - `--modal-backdrop`: `rgba(0, 0, 0, 0.75)`

### 4.2 Page Components & Layout
1. **Top Navigation**:
   - Right-aligned links:
     - Help button: `"Grammar & Help"` (opens `<dialog id="help-modal">`).
     - GitHub button with inline SVG Octocat linking to `https://github.com/darkliquid/roll`.
2. **Main Google-Style Centered Container**:
   - **Hero Logo**:
     - Inline SVG depicting a 20-sided icosahedron in d20 crimson red with white numerals ("20") and modern typography: **Roll**.
   - **Search / Input Box**:
     - Large rounded pill input (`border-radius: 24px`, shadow, smooth focus transition).
     - Placeholder: `e.g. 4d6kh3, 1d20+5, 3d6!`
     - Keyboard: <kbd>Enter</kbd> triggers roll.
   - **Button Row**:
     - `Roll`: Triggers execution of the input box expression.
     - `I'm feeling lucky`: Picks a random expression from the curated RPG pool, inputs it, and rolls.
   - **Results Area**:
     - Appears below buttons, replaced on each roll.
     - Large summary card featuring:
       - Grand total in a prominent d20 ruby badge.
       - Formula / expression badge.
       - Individual dice roll chips with roll values / symbols (including Fate symbols `⊞`, `⊟`, `☐`).
       - Success / failure metrics if applicable.
       - Clear error notice with suggestions if parsing fails.
3. **Footer**:
   - Clean, centered bottom bar:
     - "Roll Dice Engine" & link to GitHub repository.

### 4.3 Help Modal (`<dialog id="help-modal">`)
- Triggered by "Grammar & Help".
- Dismissible via close button (<kbd>&times;</kbd>), outside backdrop click, or <kbd>Esc</kbd>.
- Contains categorized syntax explanations:
  - **Basic Dice**: `NdX` (e.g., `4d6`), `d%` (percentile), `dF` (Fate/Fudge).
  - **Exploding**: `!` (explode), `!!` (compound), `!p` (penetrating).
  - **Keep & Drop**: `khN` (keep highest), `klN` (keep lowest), `dhN` (drop highest), `dlN` (drop lowest).
  - **Rerolls**: `rN` (reroll indefinitely), `roN` (reroll once).
  - **Success / Failure**: `>N`, `<N`, `=N`, `fN` (failure threshold).
  - **Sorting**: `s` (ascending), `sd` (descending).
  - **Grouping**: `{expr1, expr2}`.
- Every syntax category includes **interactive example chips** that fill the input and roll directly when clicked.

### 4.4 "I'm Feeling Lucky" Generator Pool
Curated expressions covering the gamut of supported features:
- `4d6kh3` (D&D 5e ability score)
- `1d20+7` (Attack roll with +7 modifier)
- `2d20kl1` (Disadvantage roll)
- `3d6!` (Exploding dice)
- `4d6!!` (Compounding dice)
- `2d8!p+2` (Penetrating dice)
- `4dF+1` (Fate Core roll)
- `1d%` (Percentile check)
- `5d10>7` (Storyteller dice pool)
- `6d6>4f1` (Success target with failure cancellation)
- `4d6ro1` (Great Weapon Fighting reroll once)
- `6d6sd` (Sorted descending)
- `{4d6, 2d8}>4` (Grouped dice evaluation)
- `8d6` (Fireball damage)

---

## 5. Verification & Testing

1. **WASM Build**: Execute `mise run build-wasm` and verify `docs/roll.wasm` and `docs/wasm_exec.js` are created.
2. **Go Test Suite**: Ensure all existing tests pass (`go test ./...`) without regression.
3. **Static Server Test**: Run a local HTTP server (`python3 -m http.server` or Go server) serving `docs/`.
4. **Functional Testing**:
   - Verify WASM loads and initializes cleanly in the browser.
   - Verify standard rolls (`1d20`, `4d6kh3`, `4dF`) output correct totals and dice chips.
   - Verify "I'm feeling lucky" generates and rolls valid random expressions.
   - Verify error messages display clearly for invalid syntax (e.g. `d`, `abc`).
   - Verify dark mode renders with proper contrast when system preference is dark.
   - Verify help modal opens, closes, and interactive example chips work.
