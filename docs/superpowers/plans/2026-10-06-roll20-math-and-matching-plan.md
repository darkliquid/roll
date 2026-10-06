# Implementation Plan: Roll20 Math Operators, Math Functions & Dice Matching

**Date**: 2026-10-06  
**Spec**: `docs/superpowers/specs/2026-10-06-roll20-math-and-matching-design.md`

Each task is TDD: add a failing test, implement, make it pass, commit. Every commit leaves
the tree green (`go test ./...`, `go vet ./...`, `golangci-lint run`).

## Task 0 - Foundation + multiplication

- [ ] Add `tMULT`, `tLPAREN`, `tRPAREN` tokens; scan `*`, `(`, `)`.
- [ ] Stop die scanning at math characters.
- [ ] Add `numberNode`, `binaryNode`, `parenNode`; `OpPushNumber`, `OpBinary`.
- [ ] Add the expression parser (additive/multiplicative/power/unary/primary) and modifier
      folding; route `Parse` through it.
- [ ] Evaluate `*` with `float64`; render/canonicalise; bump binary version.
- [ ] Tests: `2d6*3`, `(2d6+3)*4`, `2d6*3+4`, `2d6+4*3`, bare `5*3`.

## Task 1 - Division

- [ ] Add `tDIV`, scan `/`, evaluate with `float64`, render.
- [ ] Test `2d6/2`, `7/2`, `floor(11/2)`-style chaining, divide-by-zero error.

## Task 2 - Modulus

- [ ] Add `tMOD`, scan `%` (disambiguated from `d%`), evaluate with `math.Mod`.
- [ ] Test `2d6%3`, `7%3`, `d%` still compiles.

## Task 3 - Exponentiation

- [ ] Add `tPOW`, scan `**`, evaluate with `math.Pow` (right associative).
- [ ] Test `2**3`, `2**3**2`, `3d6**2`.

## Task 4 - `floor(x)`

- [ ] Add `tFUNC`, function look-ahead scanning, `funcNode`, `OpFunc`.
- [ ] Test `floor(11/2)`, `floor(-3/2)`, `floor(2d6/2)`.

## Task 5 - `round(x)`

- [ ] Test `round(11/2)`, `round(-5/2)`, `round(2.4)`.

## Task 6 - `ceil(x)`

- [ ] Test `ceil(11/2)`, `ceil(-3/2)`.

## Task 7 - `abs(x)`

- [ ] Test `abs(-3)`, `abs(3-10)`, `abs(-2d6)`.

## Task 8 - Dice matching (`mt`)

- [ ] Add `tMATCH`, scan `m`/`mt`, `MatchOp` on `DiceTerm`, evaluation, render, canonical,
      binary schema + version bump.
- [ ] Test `2d6mt`, `20d6mt`, `6d6mt3`, `5d6mt3>4`, `2d6m`.

## Task 9 - Wrap-up

- [ ] Add REPL catalog entries for the new syntax.
- [ ] Rebuild `docs/roll.wasm`; run `mise run ci`.
