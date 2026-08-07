# Design Spec: Compact Binary Program Representation & Canonicalisation System

**Date**: 2026-08-07  
**Status**: Approved  
**Target Files**: `binary.go`, `canonical.go`, `binary_canonical_test.go`  

## 1. Objective

Design and implement:
1. A compact, fast, and versioned binary encoding protocol for compiled dice VM `Program` instances (`MarshalBinary` / `UnmarshalBinary`).
2. A deterministic `Canonical` formatting system that standardizes dice notation strings (modifier ordering, casing, default omit rules, commutative term ordering) so equivalent dice rolls produce identical canonical string keys suitable for hashing and caching.

---

## 2. Compact Binary Specification

`*Program` implements Go's standard `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`.

### Wire Format Header:
- `Magic` (2 bytes): `0x52 0x4F` (ASCII `'R'`, `'O'`)
- `Version` (1 byte): `0x01`

### Binary Schema Details (Varint / Bitmask):
1. **Program Metadata**:
   - `MaxDepth` (Varint)
   - `Rendered` (Uvarint len + UTF-8 string bytes)

2. **Bytecode Instructions**:
   - `len(Code)` (Uvarint)
   - For each instruction: `Opcode` (uint8), `Arg` (Varint)

3. **Dice Terms**:
   - `len(DiceTerms)` (Uvarint)
   - For each `DiceTerm`:
     - `Multiplier` (Varint)
     - `Die`: `TypeTag` (uint8: 1=Normal, 2=Percentile, 3=Fate) + `Sides` (Varint if Normal)
     - `Modifier` (Varint)
     - `PresenceMask` (uint8 bitflags):
       - Bit 0: Exploding present
       - Bit 1: Limit present
       - Bit 2: Success present
       - Bit 3: Failure present
       - Bit 4: Rerolls present
     - `Exploding` (if Bit 0): `Type` (uint8) + `ComparisonOp`
     - `Limit` (if Bit 1): `Type` (uint8) + `Amount` (Varint)
     - `Success` (if Bit 2): `ComparisonOp`
     - `Failure` (if Bit 3): `ComparisonOp`
     - `Rerolls` (if Bit 4): `len(Rerolls)` (Uvarint) + list of `RerollOp` (`Once` bool + `ComparisonOp`)
     - `Sort` (uint8 enum)

4. **Group Terms**:
   - `len(GroupTerms)` (Uvarint)
   - For each `GroupTerm`:
     - `Modifier` (Varint)
     - `ChildCount` (Varint)
     - `Flags` (uint8): `Combined` (bit 0), `Negative` (bit 1)
     - `PresenceMask` (uint8): Limit (bit 0), Success (bit 1), Failure (bit 2)
     - Sub-fields for `Limit`, `Success`, `Failure` if present

5. **ComparisonOp Sub-encoding**:
   - `Type` (uint8: 0=Equals, 1=GreaterThan, 2=LessThan)
   - `Flags` (uint8: bit 0 = Inclusive)
   - `Value` (Varint)

---

## 3. Canonicalisation Rules

A `Canonical() string` method on `*Program` and top-level `Canonicalize(rollString string) (string, error)` function will enforce:

1. **Symbol Normalization**:
   - `1d20` -> `d20` (omit default count 1)
   - `d%` (percentile), `dF` (Fate)
   - `kh1` -> `kh`, `dl1` -> `dl` (omit default count 1)

2. **Strict Modifier Ordering per Dice Term**:
   - `[Multiplier][Die][Modifier][Rerolls (sorted)][Exploding][Limit][Success][Failure][Sort]`

3. **Strict Group Term Canonicalisation**:
   - `{ [Children] }[Limit][Success][Failure][Modifier]`
   - Commutative combined children sorted lexicographically by canonical string (e.g. `{2d8 + 3d6}` -> `{3d6 + 2d8}`).

---

## 4. API & Verification

New APIs:
- `(p *Program) MarshalBinary() ([]byte, error)`
- `(p *Program) UnmarshalBinary(data []byte) error`
- `UnmarshalProgram(data []byte) (*Program, error)`
- `(p *Program) Canonical() string`
- `Canonicalize(rollString string) (string, error)`

Test suite in `binary_canonical_test.go` will test round-trip binary encoding, corruption error handling, and canonical string equivalencies.
