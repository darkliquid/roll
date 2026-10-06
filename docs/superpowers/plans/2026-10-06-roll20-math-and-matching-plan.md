# Implementation Plan: Roll20 Math Operators, Math Functions & Dice Matching

**Date**: 2026-10-06  
**Spec**: `docs/superpowers/specs/2026-10-06-roll20-math-and-matching-design.md`

Each task is TDD: add a failing test, implement, make it pass, commit. Every commit leaves
the tree green (`go test ./...`, `go vet ./...`, `golangci-lint run`).

## Task 0 - Foundation + multiplication

- [x] Add `tMULT`, `tLPAREN`, `tRPAREN` tokens; scan `*`, `(`, `)`.
- [x] Stop die scanning at math characters.
- [x] Add `numberNode`, `binaryNode`, `parenNode`; `OpPushNumber`, `OpBinary`.
- [x] Add the expression parser (additive/multiplicative/power/unary/primary) and modifier
      folding; route `Parse` through it.
- [x] Evaluate `*` with `float64`; render/canonicalise; bump binary version.
- [x] Tests: `2d6*3`, `(2d6+3)*4`, `2d6*3+4`, `2d6+4*3`, bare `5*3`.

## Task 1 - Division

- [x] Add `tDIV`, scan `/`, evaluate with `float64`, render.
- [x] Test `2d6/2`, `7/2`, `floor(11/2)`-style chaining, divide-by-zero error.

## Task 2 - Modulus

- [x] Add `tMOD`, scan `%` (disambiguated from `d%`), evaluate with `math.Mod`.
- [x] Test `2d6%3`, `7%3`, `d%` still compiles.

## Task 3 - Exponentiation

- [x] Add `tPOW`, scan `**`, evaluate with `math.Pow` (right associative).
- [x] Test `2**3`, `2**3**2`, `3d6**2`.

## Task 4 - `floor(x)`

- [x] Add `tFUNC`, function look-ahead scanning, `funcNode`, `OpFunc`.
- [x] Test `floor(11/2)`, `floor(-3/2)`, `floor(2d6/2)`.

## Task 5 - `round(x)`

- [x] Test `round(11/2)`, `round(-5/2)`, `round(2.4)`.

## Task 6 - `ceil(x)`

- [x] Test `ceil(11/2)`, `ceil(-3/2)`.

## Task 7 - `abs(x)`

- [x] Test `abs(-3)`, `abs(3-10)`, `abs(-2d6)`.

## Task 8 - Dice matching (`mt`)

- [x] Add `tMATCH`, scan `m`/`mt`, `MatchOp` on `DiceTerm`, evaluation, render, canonical,
      binary schema + version bump.
- [x] Test `2d6mt`, `20d6mt`, `6d6mt3`, `5d6mt3>4`, `2d6m`.

## Task 9 - Wrap-up

- [x] Add REPL catalog entries for the new syntax.
- [x] Rebuild `docs/roll.wasm`; run `mise run ci`.

## Task 10 - Dynamic die size

- [x] Parse a bare `d(expr)`; fold constant expressions, otherwise evaluate at roll time.
- [x] Reject fractional and unsafe results (compile time when constant, roll time otherwise).
- [x] Serialize the nested side program (binary version `0x04`).
- [x] Tests: `3d(floor(6/2))`, `2d(2+2)`, `9d(1d6)`, `3d(2d6kh1)`, `3d(4d(d8)kh3+2)`,
      and the runtime fractional/unsafe errors.

## Task 11 - Tooling and docs

- [x] Build the repl binary as part of `mise run build` (CI check).
- [x] Document math operators, functions, matching and dynamic dice on the website.
- [x] Add exhaustive tests for odd constructions (`4d%%2`, `2d%kh1`, ...).

## Task 12 - One-sided dice (#12)

- [x] Allow a one-sided die and fold it to its fixed value without rolling; skip
      rerolls/explosions so they cannot loop.
- [x] Tests: `d1`, `4d1`, `2d1+5`, `4d1>0`, `4d1kh3`, `4d1!`, `4d1r1`, computed `9d1`.

## Task 13 - Penetrating zero (#9)

- [x] Drop a penetrating die whose reduced face reaches zero from the results.
- [x] Tests: `2d6!p5` and `2d6!p6` (kept reduced face).
