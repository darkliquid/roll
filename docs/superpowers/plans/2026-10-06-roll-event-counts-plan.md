# Implementation Plan: Roll Event Counts (rerolls, explosions, drops)

**Date**: 2026-10-06  
**Spec**: `docs/superpowers/specs/2026-10-06-roll-event-counts-design.md`  
**Issue**: #10

Each task is TDD: add a failing test, implement, make it pass, commit. Every commit leaves
the tree green (`mise run ci`).

## Task 1 - Counter fields and reroll counting

- [ ] Add `Rerolls`, `Explosions`, `Drops` to `Result`.
- [ ] Increment `Rerolls` in `applyRerolls` once per reroll performed.
- [ ] Tests: `4d6ro1`, `4d6r1` (recursive counts every reroll), `4d6r<2`.

## Task 2 - Explosion counting

- [ ] Increment `Explosions` in `applyExplosions` per contributing explosion: each extra
      `!`/`!p` die and each `!!` contribution.
- [ ] Test that a penetrating die elided at zero is **not** counted.
- [ ] Tests: `2d6!5`, `2d6!!5` (per contributing explosion), `2d6!p5` (elided zero),
      `2d6!p6` (kept reduced die).

## Task 3 - Drop counting

- [ ] Have `applyLimit` add the number of removed dice to `Drops`.
- [ ] Tests: `4d6kh3`, `4d6kl2`, `4d6dh1`, `4d6dl1`, `4d6kh1`.

## Task 4 - Aggregation

- [ ] Sum counters in the VM `OpBinary` case.
- [ ] Sum child counters in `evalGroupTerm`.
- [ ] Tests: `{3d6!5, 2d8}` and `{3d6!5 + 2d8}`, nested groups, a combined roll plus a
      separate term.

## Task 5 - Exclude dynamic die faces

- [ ] Confirm `resolveDie` discards the side program's counters (only the numeric total is
      used).
- [ ] Test: `3d(4d(d8)!kh3+2)` counts only the outer term's events, not the side roll's.

## Task 6 - Website

- [ ] Add `Rerolls`/`Explosions`/`Drops` and `Has*` flags to `RollResponse`; derive `Has*`
      from the program's dice/group terms without descending into `Sides`.
- [ ] Engine test for the new fields.
- [ ] Render a stats line in `docs/index.html`, shown only when a `Has*` flag is set, with
      present-but-zero modifiers displayed as `0`.
- [ ] Rebuild `docs/roll.wasm`.

## Task 7 - Wrap-up

- [ ] Run `mise run ci`; commit, push and open a PR with `Fixes #10`.
