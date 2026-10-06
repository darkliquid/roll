# Design Spec: Roll Event Counts (rerolls, explosions, drops)

**Date**: 2026-10-06  
**Status**: Draft  
**Issue**: #10  
**Target Files**: `ast.go`, `cmd/wasm/engine.go`, `cmd/wasm/engine_test.go`, `docs/index.html`, `roll20_spec_test.go`

## 1. Objective

The `Result` value only exposes the dice that survive to the end, so a caller cannot tell
how much work the roll did. Expose, for the whole expression:

1. how many dice were **rerolled** (`r`, `ro`),
2. how many dice were **added** by explosions (`!`, `!!`, `!p`),
3. how many dice were **dropped** by keep/drop limits (`kh`, `kl`, `dh`, `dl`),

and surface those counts on the website when they are applicable.

Counts describe the *final result set* only: rolls made while computing a dynamic die's
face (`DiceTerm.Sides`) are excluded.

### What is counted

Counters record the dice each operation contributes to the final result set:

- `Rerolls` counts **every reroll performed**; a die rerolled recursively counts each time,
  because each reroll replaces its face.
- `Explosions` counts **every explosion that contributes a die**: each extra `!`/`!p` die,
  and each explosion that feeds a `!!` compound (a compound fed by three explosions counts
  three).
- `Drops` counts every die removed by a keep/drop limit.

The one nuance versus "count the dice" is the penetrating elision: a die reduced to zero
never reaches the result set, so it is not counted.

---

## 2. Data model

`Result` gains three exported counters (additive, non-breaking):

```go
type Result struct {
    Results    []DieRoll
    Total      int
    Successes  int
    Rerolls    int
    Explosions int
    Drops      int
}
```

### Semantics

| Counter | Counts | Notes |
| --- | --- | --- |
| `Rerolls` | Every reroll performed. | A die rerolled recursively counts each time; `ro` rerolls a die at most once. |
| `Explosions` | Every explosion that contributes a die. | `!`/`!p` count each extra die (a penetrating die reduced to zero is elided and counts nothing); `!!` counts each explosion feeding the compound. |
| `Drops` | Every die removed by a keep/drop limit. | `4d6kh3` → 1, `4d6kh1` → 3, `4d6dh1` → 1. Applies to dice terms and groups. |

The issue phrased these as counts of `r`/`ro` and `!`/`!!`/`p` triggers. We count rerolls
and contributing explosions as events, but a penetrating die that does not survive is
never counted.

---

## 3. Aggregation

Counts are produced where the event happens and combined as the expression is evaluated:

- `evalDiceTerm` records its own `Rerolls`, `Explosions` and `Drops`.
- `applyRerolls` increments `Rerolls` once per reroll performed; `applyExplosions`
  increments `Explosions` per contributing explosion (each extra `!`/`!p` die, and each
  `!!` contribution; an elided penetrating zero is never appended, so it is not counted);
  `applyLimit` adds the number of dice it removed to `Drops`.
- The VM's `OpBinary` adds the counters of both operands (like it already does for
  `Successes`).
- `evalGroupTerm` adds the counters of its children, then its own limit drops.

The top-level `Result` returned by `EvaluateProgram` therefore carries the totals for the
entire expression, including nested groups.

### Exclusion of dynamic die faces

`resolveDie` evaluates `DiceTerm.Sides` with `runProgram` and uses only the numeric total;
the sub-result's counters are discarded, so they never reach the final result. Those rolls
still count against safety limits (`MaxRollsTotal`, depth) because the context is shared.

---

## 4. Serialization and canonicalisation

No change. `Result` is a runtime value and is not part of the compiled `Program`, so the
binary wire format and `Canonical` output are untouched. No version bump.

---

## 5. Website

`RollResponse` (`cmd/wasm/engine.go`) gains:

```go
Rerolls, Explosions, Drops       int  `json:"...,omitempty"`
HasRerolls, HasExplosions, HasDrops bool `json:"...,omitempty"`
```

The `Has*` flags mirror the existing `HasSuccesses`: they are derived from the compiled
program's dice/group terms (a term with `Rerolls`/`Exploding`/`Limit`), and deliberately do
not descend into `DiceTerm.Sides`, matching the exclusion rule.

`docs/index.html` shows a stats line under the result only when at least one `Has*` is
true, e.g. `Rerolls: 2 · Explosions: 1 · Drops: 3`. A modifier that is present but did not
fire is shown as `0` (e.g. `4d6!` with no sixes shows `Explosions: 0`); a roll without the
modifier omits the stat entirely.

---

## 6. Testing

- **Unit** (`roll20_spec_test.go` / a new `counts_test.go`): reroll counts for `r`/`ro`
  (including recursive), explosion counts for `!`/`!!`/`!p`, drop counts for
  `kh`/`kl`/`dh`/`dl`, aggregation across groups and nested groups, a penetrating die that
  is elided still counting as an explosion, and a dynamic-size roll proving side-program
  events are excluded.
- **Engine** (`cmd/wasm/engine_test.go`): the response carries the counts and `Has*` flags.
- **Manual**: the website stats line renders, omits non-applicable stats, and shows zeros
  for present modifiers.

---

## 7. Out of scope

- The REPL: its `evaluator` returns a formatted string rather than a `Result`, so exposing
  counts there is a separate change.
- `m` (visual matching) and per-die event attribution.
