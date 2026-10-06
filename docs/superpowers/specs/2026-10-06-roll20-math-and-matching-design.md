# Design Spec: Roll20 Math Operators, Math Functions & Dice Matching

**Date**: 2026-10-06  
**Status**: Approved  
**Target Files**: `tokens.go`, `scanner.go`, `parser.go`, `ast.go`, `canonical.go`, `binary.go`, `roll20_spec_test.go`, `scanner_test.go`

## 1. Objective

Close the remaining gaps in the Roll20 Dice Reference implementation tracked by issue #6:

1. Math operators: multiplication (`*`), division (`/`), modulus (`%`), exponentiation (`**`).
2. Math functions: `floor(x)`, `round(x)`, `ceil(x)`, `abs(x)`.
3. Dice matching: `mt` (Yahtzee-style match counting).

The work follows the order of operations defined by the Roll20 spec.

---

## 2. Grammar

The current parser is a term parser: a roll is a sum of dice/group terms where `+`/`-`
attach to a term as a *modifier* (`4d6kh3+2` renders as `4d6+2kh3`). That representation is
preserved. A math expression layer is layered on top of the existing term parser:

```
expression      := additive
additive        := multiplicative (('+' | '-') multiplicative)*
multiplicative  := unary (('*' | '/' | '%') unary)*
power           := unary ('**' power)?          // right associative
unary           := ('+' | '-') unary | primary
primary         := NUMBER
                 | FUNCTION '(' expression ')'
                 | '(' expression ')'
                 | rollTerm
rollTerm        := [count] 'd' sides postfix*
```

Precedence (low to high): additive, multiplicative, exponentiation, unary, parentheses/functions.

### Modifier folding

To preserve existing normalisation, when an additive operator joins a *bare* dice or group
node on the left with a *constant* on the right, the constant is folded into that node's
`Modifier` (for `+`) or subtracted (for `-`). Otherwise an `add`/`sub` binary node is built.
This keeps `4d6kh3+2` rendering as `4d6+2kh3` and `3d6+2>3` behaving as a target-number
modifier, while `2d6+4*3` correctly evaluates as `2d6 + (4*3)`.

Postfix modifiers (`kh`, `kl`, `dh`, `dl`, `r`/`ro`, `!`/`!!`/`!p`, `s`/`sa`/`sd`, success
and failure comparisons, `mt`) stay attached to the dice/group term. After a constant is
folded, any further postfix modifiers are applied to the now-updated term.

Group internals (`{...}`) are unchanged: `+`/`-` remain separators/negation and math is not
parsed inside groups.

---

## 3. Value model

Arithmetic is evaluated with `float64` so that `floor(11/2)` observes `5.5`. The VM stack
element gains a numeric channel:

```go
type vmValue struct {
    Result   Result
    Modifier int
    Num      float64
    IsNum    bool
}
```

- Roll/group results carry `Result`; their numeric value is `float64(Result.Total)`.
- Number literals, binary results and function results are numeric (`IsNum`), keeping the
  accumulated `Num` for chained operations. `Result.Total` is the truncated integer, and
  any dice rolled by operands are carried in `Result.Results` so they still display.

New opcodes (encoded with their operand inline in `Instruction.Arg`, so the bytecode section
of the binary format is unchanged):

| Opcode | Arg | Effect |
| --- | --- | --- |
| `OpPushNumber` | integer value | push a numeric literal |
| `OpBinary` | `BinaryOpType` | pop two values, apply operator, push |
| `OpFunc` | `FuncType` | pop one value, apply function, push |

Division or modulus by zero returns `ErrDivisionByZero`.

---

## 4. Functions

| Function | Behaviour |
| --- | --- |
| `floor(x)` | rounds toward negative infinity |
| `ceil(x)` | rounds toward positive infinity |
| `round(x)` | rounds toward zero when the fraction is `< 0.5`, otherwise toward positive infinity (`math.Floor(x + 0.5)`) |
| `abs(x)` | absolute value |

Function names are lower-case keywords; the scanner distinguishes `floor`/`round` from the
`f` (failure) and `r` (reroll) modifiers by look-ahead.

---

## 5. Dice matching (`mt`)

`mt` is a postfix modifier on a dice term:

```
mt[count][comparison]
```

- `count` defaults to 2: the minimum number of identical faces required for a match.
- `comparison` optionally filters which face value qualifies (`5d6mt3>4`).

The total becomes the number of distinct face values that appear at least `count` times and
satisfy the comparison (e.g. `20d6mt` yields 0-6). `m` (visual-only) is accepted and has no
effect on the result. Because matching is stored on `DiceTerm`, the binary format is bumped.

---

## 6. Compatibility

- Existing normalisation, canonical ordering and binary round-tripping are preserved for all
  pre-existing notation; the binary version is bumped when the `DiceTerm` schema changes.
- `Canonical` remains deterministic: commutative operators (`+`, `*`) sort their operands.
- No changes to the public API surface.
