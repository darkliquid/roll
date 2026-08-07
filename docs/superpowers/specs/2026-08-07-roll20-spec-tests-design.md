# Design Spec: Roll20 Dice Rolling Language Specification Test Suite

**Date**: 2026-08-07  
**Status**: Approved  
**Target File**: `roll20_spec_test.go`  

## 1. Objective

Validate the `roll` Go library against the Roll20 Dice Rolling Language Specification by establishing a dedicated, comprehensive test suite (`roll20_spec_test.go`). The suite ensures that every operation type and combination of operations—including dice types, target numbers, exploding dice, keep/drop, rerolls, sorting, grouped rolls, and grouping modifiers—are thoroughly tested for both AST compilation and VM evaluation.

---

## 2. Validation Findings & Scope

### Supported Operations Verified in Codebase (`scanner.go`, `parser.go`, `ast.go`):
1. **Basic Dice & Types**: Standard (`XdY`, `dY`), Percentile (`d%`), Fate/Fudge (`dF`).
2. **Comparison & Target Numbers**: Equals (`=`), Greater Than (`>`), Less Than (`<`), Greater/Equal (`>=`), Less/Equal (`<=`), Failure (`f`).
3. **Exploding Dice**: Exploding (`!`), Compounded (`!!`), Penetrating (`!p`).
4. **Keep / Drop Modifiers**: Keep Highest (`kh`), Keep Lowest (`kl`), Drop Highest (`dh`), Drop Lowest (`dl`).
5. **Reroll Modifiers**: Recursive Reroll (`r`), Reroll Once (`ro`).
6. **Sorting**: Ascending (`s`, `sa`), Descending (`sd`).
7. **Groupings & Group Modifiers**:
   - Separated groups (`{a, b}`) vs. Combined groups (`{a + b}`).
   - Group Keep/Drop (`{...}kh1`, `{...}dl1`).
   - Group Success/Failure (`{...}>10`, `{...}f<2`).
   - Group Math Modifiers (`{...}+4`, `-{...}`).
   - Nested Groups (`{{a, b}, c}`).

---

## 3. Test Architecture & Structure

The tests will be contained entirely in `roll20_spec_test.go`.

### Deterministic Test Setup
Tests utilize `withTestSeed(seed, ...)` to ensure pseudo-random dice rolls produce deterministic outcomes across runs.

### Test Matrix Structure
Each test case uses the following struct:

```go
type roll20TestCase struct {
    name       string
    seed       int64
    input      string
    wantString string
    wantRolls  []int
    wantTotal  int
    wantScnt   int
}
```

### Test Sub-Suites:

1. `TestRoll20Spec_BasicDice`
   - `d20`, `3d6`, `d%`, `2d%`, `4dF`, `4dF+2`, `+3d6`, `-2d8`, `3d6+4`, `2d8-2`

2. `TestRoll20Spec_TargetNumbers`
   - Strict and inclusive comparisons (`>4`, `>=4`, `<3`, `<=3`, `=6`)
   - Failure checks (`f=1`, `f<2`, `f>5`)
   - Modifier combined with target numbers (`3d6+2>5`)

3. `TestRoll20Spec_ExplodingDice`
   - Exploding (`!6`, `!>4`)
   - Compounded (`!!6`, `!!>4`)
   - Penetrating (`!p6`, `!p>4`)

4. `TestRoll20Spec_KeepDrop`
   - Keep highest/lowest (`kh3`, `kl1`)
   - Drop highest/lowest (`dh1`, `dl2`)
   - Limit combined with math (`4d6kh3+2`)

5. `TestRoll20Spec_Reroll`
   - Reroll once (`ro1`, `ro<3`)
   - Recursive reroll (`r1`, `r<3`)

6. `TestRoll20Spec_Sorting`
   - Ascending (`s`, `sa`)
   - Descending (`sd`)

7. `TestRoll20Spec_GroupedRolls`
   - Single item group (`{3d6}`)
   - Separated group (`{3d6, 2d8}`)
   - Combined group (`{3d6 + 2d8}`)
   - Group math (`{3d6, 2d8}+5`)
   - Group negation (`-{3d6, 2d8}`)

8. `TestRoll20Spec_GroupingModifiers`
   - Group keep/drop (`{3d6, 2d8}kh1`, `{3d6, 2d8}dl1`)
   - Group target numbers (`{3d6, 2d8}>10`)
   - Group failure (`{3d6, 2d8}>8f<5`)

9. `TestRoll20Spec_ComplexPermutations`
   - Stacked modifiers (`6d6!!5kh3sd+3`, `4d6!>5ro<2kh3s`)
   - Nested groupings (`{{3d6, 2d8}kh1, 4d4}`, `{3d6+2d8-{4d4-1}dl}kh3<4f>3`)

---

## 4. Verification

Run `go test -v ./...` to ensure all existing and new spec tests pass cleanly.
