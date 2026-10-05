# AGENTS.md

Guidance for agents working in this repository.

## What this is

`github.com/darkliquid/roll` is a dice-expression **compiler and stack-based VM** for Go, implementing (most of) the [Roll20 Dice Rolling Language Specification](https://wiki.roll20.net/Dice_Reference). The root package `roll` is a library; `cmd/repl` is an interactive terminal front-end.

## Commands

```bash
mise run ci             # build + vet + lint + vulncheck + test (what CI runs)
mise run build          # compile all packages, including the js/wasm target
mise run lint           # golangci-lint
mise run vulncheck      # govulncheck
mise run build-wasm     # compile docs/roll.wasm + sync wasm_exec.js
go test ./...           # run all tests directly
go test -v ./cmd/repl   # run one package verbosely
```

`mise.toml` pins the toolchain (Go 1.27.1, golangci-lint 2.14.0, govulncheck 1.8.0); run `mise install` first. CI is GitHub Actions (`.github/workflows/ci.yml`) using `jdx/mise-action` with caching, and just runs `mise run ci`.

### Lint and CI gotchas

- `.golangci.yml` uses the golangci-lint **v2** config format and enables the standard linter set plus `gofmt`/`goimports` formatters, so `golangci-lint run` also fails on formatting; fix with `golangci-lint fmt`.
- `golangci-lint run` hides repeated messages by default (`max-same-issues: 3`). To see every occurrence use `golangci-lint run --max-same-issues=0 --max-issues-per-linter=0`.
- On a non-js host `go build ./...` skips `cmd/wasm/main.go` (build-tagged). `mise run build` additionally compiles the js/wasm target so the wasm entrypoint is type-checked.

## Architecture and data flow

The pipeline is: **source string → Scanner → Parser (compiler) → `Program` bytecode → VM evaluator → `Result`**.

| File | Role |
| --- | --- |
| `scanner.go` | Lexer. Produces `Token` + literal. Uses a single-rune `bufio.Reader` with `read`/`unread`. |
| `tokens.go` | `Token` enum (`tNUM`, `tDIE`, `tEXPLODE`, ...). All unexported. |
| `parser.go` | Compiler. Builds an internal `compiledNode` tree, then `emit`s a `Program`. |
| `ast.go` | Core types (`Program`, `DiceTerm`, `GroupTerm`, `ComparisonOp`, ...) **and** the VM (`EvaluateProgram`). |
| `dice.go` | `Die` interface and `NormalDie` / `FateDie` / `PercentileDie`. |
| `canonical.go` | Deterministic normalized notation (`Canonical` / `Canonicalize`). |
| `binary.go` | Custom binary serialization (`MarshalBinary` / `UnmarshalBinary`). |
| `rolling.go` | Public top-level API: `Compile`, `CompileString`, `Parse`, `ParseString`, each with a `...WithLimits` variant. |
| `cmd/repl/` | Bubbletea TUI (`model.go`), syntax catalog (`catalog.go`), entrypoint (`main.go`). |
| `cmd/wasm/` | WebAssembly engine (`engine.go`) + js/wasm entrypoint (`main.go`) exposing `window.rollDice`. `main_nonwasm.go` supplies an empty host `main` so the package compiles under `go build ./...`. |

### Key mechanics

- `Program` is a linear instruction list (`OpRollDice`, `OpRollGroup`) plus parallel `DiceTerms` / `GroupTerms` slices. Each instruction's `Arg` indexes into those slices. The VM is a stack of `vmValue`; `OpRollGroup` pops `ChildCount` children off the stack, aggregates them, and pushes the group result.
- `Program.String()` returns `Rendered` (the parser's normalized notation). `Program.Canonical()` returns a **different**, deterministic form with commutative terms sorted and defaults omitted. Do not conflate them.
- The parser has only a **one-token unscan buffer** (`Parser.buf`), not arbitrary lookahead. `scanIgnoreWhitespace` skips at most one whitespace token.

## Testing

- Tests are **table-driven** and **deterministic**. Randomness flows through the package-level var `randomIntn` (`dice.go:9`). Tests override it with the `withTestSeed(seed, func(){...})` helper in `dice_test.go`.
- **Any test that exercises a roll must wrap it in `withTestSeed`** or it will be flaky. Existing expectations are tied to exact seeded values (e.g. seed 0 → `3d6` rolls `1,1,2`).
- `roll20_spec_test.go` is the spec-conformance suite; add cases there via the `runRoll20TestCase`/`runRoll20TestCases` helper (checks `String()`, individual rolls, total, successes).
- `cmd/repl` tests inject a fake `evaluator func(string)(string,error)` into `model` rather than running the real engine.

## Conventions and gotchas

- **Safety limits are mandatory.** `Limits` (`MaxDieSize`, `MaxRollsPerDie`, `MaxRollsTotal`, `MaxEvalDepth`) guard against unsafe input; `DefaultLimits` applies when zero. New dice types **must** be added to the `validateDieLimits` switch (`ast.go`) or evaluation returns `unsupported die type %T`. Dice with fewer than 2 sides are rejected (`ErrUnsafeDie`).
- Error types are **string-backed** (`ErrUnexpectedToken`, `ErrUnknownDie`, `ErrEndOfRoll`, `ErrAmbiguousModifier`, `ErrLimitExceeded`). Tests compare `err.Error()` strings exactly, so changing message text breaks tests.
- **Binary format changes require a version bump.** The wire format is magic `0x52 0x4F` ("RO") + version byte (`binaryVersion = 0x01`) + varints/bitmasks. See `docs/superpowers/specs/2026-08-07-binary-canonical-design.md` for the schema.
- **Canonical ordering is load-bearing.** Combined group children are sorted lexicographically by canonical string; rerolls are sorted. Canonical output is intended as a stable hash/cache key.
- Group rendering normalizes operator spacing with chained `strings.ReplaceAll` calls (e.g. `"+-"` → `"-"`); group string formatting is subtle.
- `cmd/repl` imports `charm.land/bubbletea/v2` (not `github.com/charmbracelet/bubbletea`).
- `go.mod` lists `charm.land/bubbletea/v2` and `github.com/charmbracelet/colorprofile` with incorrect `// indirect` markers, so `gopls` reports "should be direct/indirect" warnings. This is a `go mod tidy` artifact, not a real error.

## Documentation workflow (non-obvious)

Non-trivial features follow a spec → plan → implement flow under `docs/superpowers/`:

- `specs/YYYY-MM-DD-<topic>-design.md` — design spec, marked `Status: Approved`.
- `plans/YYYY-MM-DD-<topic>-plan.md` — implementation plan with checkbox tasks and TDD steps.

Read the relevant spec before touching a subsystem. The binary/canonical and REPL autocomplete subsystems are documented there.

## WebAssembly and GitHub Pages

The browser build is implemented; see `docs/superpowers/specs/2026-10-05-github-pages-site-design.md` and its plan.

- `cmd/wasm/` holds the engine (`engine.go`, host-testable) and the js/wasm entrypoint (`main.go`). `main_nonwasm.go` (build tag `!(js && wasm)`) exists only so `go build ./...` succeeds off the wasm target; do not delete it.
- `mise run build-wasm` compiles `cmd/wasm` to `docs/roll.wasm` and copies Go's `wasm_exec.js` into `docs/`. Re-run it after changing the engine and commit the refreshed `docs/roll.wasm` / `docs/wasm_exec.js` artifacts.
- `docs/` is intentionally the GitHub Pages root and contains the committed wasm artifacts alongside `index.html`.
- `index.html` calls `window.rollDice(expr)`. The returned `text` field comes from a second, independent roll (`roll.ParseString`), so it can disagree with `results`; the UI deliberately does not display it.

## Commit style

Conventional Commits with scopes, e.g. `feat(repl): ...`, `fix(repl): ...`, `feat: ...`, `test: ...`, `docs: ...`, `refactor: ...`.
