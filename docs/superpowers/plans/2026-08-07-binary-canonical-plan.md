# Compact Binary Representation & Canonicalisation System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a compact binary marshaling/unmarshaling system for compiled roll `Program` instances and a deterministic canonical string generator (`Canonical()`) for dice roll normalization.

**Architecture:** Add `binary.go` implementing `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` with varint bitmask serialization, and `canonical.go` implementing AST modifier and term canonicalisation. Add `binary_canonical_test.go` for test coverage.

**Tech Stack:** Go 1.18+, `encoding/binary`, standard `testing` package.

---

### Task 1: Implement Compact Binary Marshaling & Unmarshaling (`binary.go`)

**Files:**
- Create: `binary.go`
- Test: `binary_canonical_test.go`

- [ ] **Step 1: Write failing binary round-trip test in `binary_canonical_test.go`**

```go
package roll

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBinaryMarshalUnmarshal(t *testing.T) {
	expressions := []string{
		"3d6+4",
		"6d6!!5kh3sd+3",
		"{3d6+4, 2d8}dl=1f>5",
		"{3d6 + 2d8 - {4d4-1}dl}kh3<4f>3",
		"4dF+2",
		"d%",
	}

	for _, expr := range expressions {
		t.Run(expr, func(t *testing.T) {
			prog1, err := CompileString(expr)
			if err != nil {
				t.Fatalf("compile error: %v", err)
			}

			data, err := prog1.MarshalBinary()
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			prog2, err := UnmarshalProgram(data)
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if !reflect.DeepEqual(prog1.Code, prog2.Code) {
				t.Errorf("Code mismatch:\nexp: %#v\ngot: %#v", prog1.Code, prog2.Code)
			}
			if !reflect.DeepEqual(prog1.DiceTerms, prog2.DiceTerms) {
				t.Errorf("DiceTerms mismatch:\nexp: %#v\ngot: %#v", prog1.DiceTerms, prog2.DiceTerms)
			}
			if !reflect.DeepEqual(prog1.GroupTerms, prog2.GroupTerms) {
				t.Errorf("GroupTerms mismatch:\nexp: %#v\ngot: %#v", prog1.GroupTerms, prog2.GroupTerms)
			}
			if prog1.MaxDepth != prog2.MaxDepth {
				t.Errorf("MaxDepth mismatch: exp %d, got %d", prog1.MaxDepth, prog2.MaxDepth)
			}

			// Verify execution output matches original
			withTestSeed(0, func() {
				res1, err1 := EvaluateProgram(prog1)
				res2, err2 := EvaluateProgram(prog2)
				if err1 != nil || err2 != nil {
					t.Fatalf("eval errors: %v / %v", err1, err2)
				}
				if !reflect.DeepEqual(res1, res2) {
					t.Fatalf("eval result mismatch:\nexp %#v\ngot %#v", res1, res2)
				}
			})
		})
	}
}

func TestBinaryUnmarshalCorrupted(t *testing.T) {
	t.Run("invalid magic header", func(t *testing.T) {
		p := &Program{}
		err := p.UnmarshalBinary([]byte{0x00, 0x00, 0x01})
		if err == nil {
			t.Fatal("expected error on invalid magic header")
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		p := &Program{}
		err := p.UnmarshalBinary([]byte{0x52, 0x4F, 0x99})
		if err == nil {
			t.Fatal("expected error on unsupported version")
		}
	})
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v -run "TestBinaryMarshalUnmarshal|TestBinaryUnmarshalCorrupted"`  
Expected: FAIL (`MarshalBinary` not defined)

- [ ] **Step 3: Implement `binary.go` with compact Varint encoding**

Create `/home/darkliquid/Projects/roll/binary.go`:

```go
package roll

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	binaryMagic0 byte = 0x52 // 'R'
	binaryMagic1 byte = 0x4F // 'O'
	binaryVersion byte = 0x01
)

// MarshalBinary serializes Program into a compact binary representation.
func (p *Program) MarshalBinary() ([]byte, error) {
	if p == nil {
		return nil, nil
	}

	buf := &bytes.Buffer{}
	buf.WriteByte(binaryMagic0)
	buf.WriteByte(binaryMagic1)
	buf.WriteByte(binaryVersion)

	writeVarint(buf, int64(p.MaxDepth))
	writeString(buf, p.Rendered)

	writeUvarint(buf, uint64(len(p.Code)))
	for _, inst := range p.Code {
		buf.WriteByte(byte(inst.Op))
		writeVarint(buf, int64(inst.Arg))
	}

	writeUvarint(buf, uint64(len(p.DiceTerms)))
	for _, dt := range p.DiceTerms {
		if err := marshalDiceTerm(buf, dt); err != nil {
			return nil, err
		}
	}

	writeUvarint(buf, uint64(len(p.GroupTerms)))
	for _, gt := range p.GroupTerms {
		if err := marshalGroupTerm(buf, gt); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary deserializes binary data into the Program.
func (p *Program) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		*p = Program{}
		return nil
	}

	r := bytes.NewReader(data)
	b0, err := r.ReadByte()
	if err != nil {
		return err
	}
	b1, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b0 != binaryMagic0 || b1 != binaryMagic1 {
		return fmt.Errorf("invalid binary magic header")
	}

	ver, err := r.ReadByte()
	if err != nil {
		return err
	}
	if ver != binaryVersion {
		return fmt.Errorf("unsupported binary version %d", ver)
	}

	maxDepth, err := readVarint(r)
	if err != nil {
		return err
	}
	rendered, err := readString(r)
	if err != nil {
		return err
	}

	codeLen, err := readUvarint(r)
	if err != nil {
		return err
	}
	code := make([]Instruction, codeLen)
	for i := uint64(0); i < codeLen; i++ {
		opByte, err := r.ReadByte()
		if err != nil {
			return err
		}
		arg, err := readVarint(r)
		if err != nil {
			return err
		}
		code[i] = Instruction{Op: Opcode(opByte), Arg: int(arg)}
	}

	diceLen, err := readUvarint(r)
	if err != nil {
		return err
	}
	diceTerms := make([]DiceTerm, diceLen)
	for i := uint64(0); i < diceLen; i++ {
		dt, err := unmarshalDiceTerm(r)
		if err != nil {
			return err
		}
		diceTerms[i] = dt
	}

	groupLen, err := readUvarint(r)
	if err != nil {
		return err
	}
	groupTerms := make([]GroupTerm, groupLen)
	for i := uint64(0); i < groupLen; i++ {
		gt, err := unmarshalGroupTerm(r)
		if err != nil {
			return err
		}
		groupTerms[i] = gt
	}

	p.MaxDepth = int(maxDepth)
	p.Rendered = rendered
	p.Code = code
	p.DiceTerms = diceTerms
	p.GroupTerms = groupTerms

	return nil
}

// UnmarshalProgram deserializes a compact binary byte slice into a new Program.
func UnmarshalProgram(data []byte) (*Program, error) {
	p := &Program{}
	if err := p.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return p, nil
}

func marshalDiceTerm(buf *bytes.Buffer, dt DiceTerm) error {
	writeVarint(buf, int64(dt.Multiplier))
	if err := marshalDie(buf, dt.Die); err != nil {
		return err
	}
	writeVarint(buf, int64(dt.Modifier))

	var mask byte
	if dt.Exploding != nil {
		mask |= 1 << 0
	}
	if dt.Limit != nil {
		mask |= 1 << 1
	}
	if dt.Success != nil {
		mask |= 1 << 2
	}
	if dt.Failure != nil {
		mask |= 1 << 3
	}
	if len(dt.Rerolls) > 0 {
		mask |= 1 << 4
	}
	buf.WriteByte(mask)

	if dt.Exploding != nil {
		buf.WriteByte(byte(dt.Exploding.Type))
		marshalComparisonOp(buf, dt.Exploding.ComparisonOp)
	}
	if dt.Limit != nil {
		buf.WriteByte(byte(dt.Limit.Type))
		writeVarint(buf, int64(dt.Limit.Amount))
	}
	if dt.Success != nil {
		marshalComparisonOp(buf, dt.Success)
	}
	if dt.Failure != nil {
		marshalComparisonOp(buf, dt.Failure)
	}
	if len(dt.Rerolls) > 0 {
		writeUvarint(buf, uint64(len(dt.Rerolls)))
		for _, rr := range dt.Rerolls {
			var flags byte
			if rr.Once {
				flags |= 1
			}
			buf.WriteByte(flags)
			marshalComparisonOp(buf, rr.ComparisonOp)
		}
	}
	buf.WriteByte(byte(dt.Sort))
	return nil
}

func unmarshalDiceTerm(r *bytes.Reader) (dt DiceTerm, err error) {
	mult, err := readVarint(r)
	if err != nil {
		return dt, err
	}
	dt.Multiplier = int(mult)

	die, err := unmarshalDie(r)
	if err != nil {
		return dt, err
	}
	dt.Die = die

	mod, err := readVarint(r)
	if err != nil {
		return dt, err
	}
	dt.Modifier = int(mod)

	mask, err := r.ReadByte()
	if err != nil {
		return dt, err
	}

	if mask&(1<<0) != 0 {
		expType, err := r.ReadByte()
		if err != nil {
			return dt, err
		}
		cmp, err := unmarshalComparisonOp(r)
		if err != nil {
			return dt, err
		}
		dt.Exploding = &ExplodingOp{Type: ExplodingType(expType), ComparisonOp: cmp}
	}

	if mask&(1<<1) != 0 {
		limitType, err := r.ReadByte()
		if err != nil {
			return dt, err
		}
		amt, err := readVarint(r)
		if err != nil {
			return dt, err
		}
		dt.Limit = &LimitOp{Type: LimitType(limitType), Amount: int(amt)}
	}

	if mask&(1<<2) != 0 {
		cmp, err := unmarshalComparisonOp(r)
		if err != nil {
			return dt, err
		}
		dt.Success = cmp
	}

	if mask&(1<<3) != 0 {
		cmp, err := unmarshalComparisonOp(r)
		if err != nil {
			return dt, err
		}
		dt.Failure = cmp
	}

	if mask&(1<<4) != 0 {
		rrLen, err := readUvarint(r)
		if err != nil {
			return dt, err
		}
		dt.Rerolls = make([]RerollOp, rrLen)
		for i := uint64(0); i < rrLen; i++ {
			flags, err := r.ReadByte()
			if err != nil {
				return dt, err
			}
			cmp, err := unmarshalComparisonOp(r)
			if err != nil {
				return dt, err
			}
			dt.Rerolls[i] = RerollOp{Once: (flags & 1) != 0, ComparisonOp: cmp}
		}
	}

	sortByte, err := r.ReadByte()
	if err != nil {
		return dt, err
	}
	dt.Sort = SortType(sortByte)

	return dt, nil
}

func marshalGroupTerm(buf *bytes.Buffer, gt GroupTerm) error {
	writeVarint(buf, int64(gt.Modifier))
	writeVarint(buf, int64(gt.ChildCount))

	var flags byte
	if gt.Combined {
		flags |= 1 << 0
	}
	if gt.Negative {
		flags |= 1 << 1
	}
	buf.WriteByte(flags)

	var mask byte
	if gt.Limit != nil {
		mask |= 1 << 0
	}
	if gt.Success != nil {
		mask |= 1 << 1
	}
	if gt.Failure != nil {
		mask |= 1 << 2
	}
	buf.WriteByte(mask)

	if gt.Limit != nil {
		buf.WriteByte(byte(gt.Limit.Type))
		writeVarint(buf, int64(gt.Limit.Amount))
	}
	if gt.Success != nil {
		marshalComparisonOp(buf, gt.Success)
	}
	if gt.Failure != nil {
		marshalComparisonOp(buf, gt.Failure)
	}
	return nil
}

func unmarshalGroupTerm(r *bytes.Reader) (gt GroupTerm, err error) {
	mod, err := readVarint(r)
	if err != nil {
		return gt, err
	}
	gt.Modifier = int(mod)

	childCnt, err := readVarint(r)
	if err != nil {
		return gt, err
	}
	gt.ChildCount = int(childCnt)

	flags, err := r.ReadByte()
	if err != nil {
		return gt, err
	}
	gt.Combined = (flags & (1 << 0)) != 0
	gt.Negative = (flags & (1 << 1)) != 0

	mask, err := r.ReadByte()
	if err != nil {
		return gt, err
	}

	if mask&(1<<0) != 0 {
		limitType, err := r.ReadByte()
		if err != nil {
			return gt, err
		}
		amt, err := readVarint(r)
		if err != nil {
			return gt, err
		}
		gt.Limit = &LimitOp{Type: LimitType(limitType), Amount: int(amt)}
	}

	if mask&(1<<1) != 0 {
		cmp, err := unmarshalComparisonOp(r)
		if err != nil {
			return gt, err
		}
		gt.Success = cmp
	}

	if mask&(1<<2) != 0 {
		cmp, err := unmarshalComparisonOp(r)
		if err != nil {
			return gt, err
		}
		gt.Failure = cmp
	}

	return gt, nil
}

func marshalDie(buf *bytes.Buffer, die Die) error {
	switch d := die.(type) {
	case NormalDie:
		buf.WriteByte(1)
		writeVarint(buf, int64(d))
	case PercentileDie:
		buf.WriteByte(2)
	case FateDie:
		buf.WriteByte(3)
	default:
		return fmt.Errorf("unknown die type %T", die)
	}
	return nil
}

func unmarshalDie(r *bytes.Reader) (Die, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case 1:
		sides, err := readVarint(r)
		if err != nil {
			return nil, err
		}
		return NormalDie(sides), nil
	case 2:
		return PercentileDie(0), nil
	case 3:
		return FateDie(0), nil
	default:
		return nil, fmt.Errorf("unknown die type tag %d", tag)
	}
}

func marshalComparisonOp(buf *bytes.Buffer, cmp *ComparisonOp) {
	if cmp == nil {
		buf.WriteByte(0)
		buf.WriteByte(0)
		writeVarint(buf, 0)
		return
	}
	buf.WriteByte(byte(cmp.Type))
	var flags byte
	if cmp.Inclusive {
		flags |= 1
	}
	buf.WriteByte(flags)
	writeVarint(buf, int64(cmp.Value))
}

func unmarshalComparisonOp(r *bytes.Reader) (*ComparisonOp, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	flags, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	val, err := readVarint(r)
	if err != nil {
		return nil, err
	}
	return &ComparisonOp{
		Type:      ComparisonType(typ),
		Inclusive: (flags & 1) != 0,
		Value:     int(val),
	}, nil
}

func writeVarint(buf *bytes.Buffer, v int64) {
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(b, v)
	buf.Write(b[:n])
}

func readVarint(r *bytes.Reader) (int64, error) {
	return binary.ReadVarint(r)
}

func writeUvarint(buf *bytes.Buffer, v uint64) {
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, v)
	buf.Write(b[:n])
}

func readUvarint(r *bytes.Reader) (uint64, error) {
	return binary.ReadUvarint(r)
}

func writeString(buf *bytes.Buffer, s string) {
	writeUvarint(buf, uint64(len(s)))
	buf.WriteString(s)
}

func readString(r *bytes.Reader) (string, error) {
	n, err := readUvarint(r)
	if err != nil {
		return "", err
	}
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run "TestBinaryMarshalUnmarshal|TestBinaryUnmarshalCorrupted"`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add binary.go binary_canonical_test.go
git commit -m "feat: implement compact binary serialization for compiled roll programs"
```

---

### Task 2: Implement Canonicalisation System (`canonical.go`)

**Files:**
- Create: `canonical.go`
- Modify: `binary_canonical_test.go`

- [ ] **Step 1: Write canonicalisation test cases in `binary_canonical_test.go`**

Append to `binary_canonical_test.go`:

```go
func TestCanonicalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "1d20", want: "d20"},
		{input: "3d6+4", want: "3d6+4"},
		{input: "1d%", want: "d%"},
		{input: "4dF+2", want: "4dF+2"},
		{input: "4d6s>4kh3", want: "4d6kh3>4s"},
		{input: "6d6sa>=5", want: "6d6>=5s"},
		{input: "{2d8 + 3d6}", want: "{3d6 + 2d8}"},
		{input: "{3d6, 2d8}", want: "{3d6, 2d8}"},
		{input: "{3d6+4, 2d8}dl=1f>5", want: "{3d6+4, 2d8}dl=1f>5"},
		{input: "{3d6+2d8-{4d4-1}dl}kh3<4f>3", want: "{3d6 + 2d8 - {4d4-1}dl}kh3<4f>3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Canonicalize(tt.input)
			if err != nil {
				t.Fatalf("Canonicalize error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Canonicalize mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}

			// Verify Canonical() method on Program produces identical output
			prog, err := CompileString(tt.input)
			if err != nil {
				t.Fatalf("CompileString error: %v", err)
			}
			if pGot := prog.Canonical(); pGot != tt.want {
				t.Errorf("Program.Canonical mismatch:\ngot:  %q\nwant: %q", pGot, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v -run "TestCanonicalize"`  
Expected: FAIL (`Canonicalize` undefined)

- [ ] **Step 3: Implement `canonical.go`**

Create `/home/darkliquid/Projects/roll/canonical.go`:

```go
package roll

import (
	"fmt"
	"sort"
	"strings"
)

// Canonical returns the deterministic, normalized dice notation for the compiled program.
func (p *Program) Canonical() string {
	if p == nil || len(p.Code) == 0 {
		return ""
	}

	// Build AST string representation canonically from AST nodes
	// We can reconstruct the root node string representation using canonical rules.
	return canonicalizeProgram(p)
}

// Canonicalize parses a dice roll string and returns its canonical representation.
func Canonicalize(input string) (string, error) {
	program, err := CompileString(input)
	if err != nil {
		return "", err
	}
	return program.Canonical(), nil
}

func canonicalizeProgram(p *Program) string {
	type stackNode struct {
		rendered string
		sortKey  string
	}

	stack := make([]stackNode, 0, len(p.Code))

	for _, inst := range p.Code {
		switch inst.Op {
		case OpRollDice:
			if inst.Arg < 0 || inst.Arg >= len(p.DiceTerms) {
				return p.Rendered
			}
			term := p.DiceTerms[inst.Arg]
			rendered := canonicalizeDiceTerm(term)
			stack = append(stack, stackNode{rendered: rendered, sortKey: rendered})

		case OpRollGroup:
			if inst.Arg < 0 || inst.Arg >= len(p.GroupTerms) {
				return p.Rendered
			}
			term := p.GroupTerms[inst.Arg]
			if term.ChildCount > len(stack) {
				return p.Rendered
			}

			children := stack[len(stack)-term.ChildCount:]
			stack = stack[:len(stack)-term.ChildCount]

			rendered := canonicalizeGroupTerm(term, children)
			stack = append(stack, stackNode{rendered: rendered, sortKey: rendered})
		}
	}

	if len(stack) != 1 {
		return p.Rendered
	}

	return strings.TrimPrefix(stack[0].rendered, "+")
}

func canonicalizeDiceTerm(term DiceTerm) string {
	var output strings.Builder

	// Multiplier & Die
	if term.Multiplier > 1 || term.Multiplier < -1 {
		output.WriteString(fmt.Sprintf("%+d", term.Multiplier))
	} else if term.Multiplier == -1 {
		output.WriteString("-")
	} else if term.Multiplier == 1 {
		output.WriteString("+")
	}

	// Die representation (canonical lowercase/uppercase symbols, omit count 1)
	output.WriteString(canonicalDieString(term.Die))

	// Term modifier
	if term.Modifier != 0 {
		output.WriteString(fmt.Sprintf("%+d", term.Modifier))
	}

	// Rerolls (sorted lexicographically by string)
	if len(term.Rerolls) > 0 {
		rrStrs := make([]string, len(term.Rerolls))
		for i, rr := range term.Rerolls {
			rrStrs[i] = rr.String()
		}
		sort.Strings(rrStrs)
		for _, s := range rrStrs {
			output.WriteString(s)
		}
	}

	// Exploding
	if term.Exploding != nil {
		output.WriteString(term.Exploding.String())
	}

	// Limit (kh, kl, dh, dl)
	if term.Limit != nil {
		output.WriteString(term.Limit.String())
	}

	// Success
	if term.Success != nil {
		output.WriteString(term.Success.String())
	}

	// Failure
	if term.Failure != nil {
		output.WriteString("f" + term.Failure.String())
	}

	// Sort
	output.WriteString(term.Sort.String())

	return output.String()
}

func canonicalDieString(die Die) string {
	switch d := die.(type) {
	case NormalDie:
		return fmt.Sprintf("d%d", d)
	case PercentileDie:
		return "d%"
	case FateDie:
		return "dF"
	default:
		return die.String()
	}
}

type stackItem struct {
	rendered string
	sortKey  string
}

func canonicalizeGroupTerm(term GroupTerm, children []stackItem) string {
	parts := make([]stackItem, len(children))
	copy(parts, children)

	if term.Combined {
		// Sort combined children to make commutative additions deterministic
		sort.SliceStable(parts, func(i, j int) bool {
			return parts[i].sortKey < parts[j].sortKey
		})
	}

	rawParts := make([]string, len(parts))
	for i, p := range parts {
		rawParts[i] = p.rendered
	}

	sep := ", "
	if term.Combined {
		sep = " + "
	}

	output := strings.Join(rawParts, sep)
	if term.Combined {
		output = strings.ReplaceAll(output, "+-", "-")
	} else if len(parts) == 1 {
		output += ","
	}

	output = "{" + output + "}"
	output = strings.ReplaceAll(output, "{+", "{")
	output = strings.ReplaceAll(output, "{-", "{")
	output = strings.ReplaceAll(output, ", +", ", ")
	output = strings.ReplaceAll(output, ", -", ", ")
	output = strings.ReplaceAll(output, "+ +", "+ ")
	output = strings.ReplaceAll(output, "+ -", "- ")

	if term.Limit != nil {
		output += term.Limit.String()
	}
	if term.Success != nil {
		output += term.Success.String()
	}
	if term.Failure != nil {
		output += "f" + term.Failure.String()
	}
	if term.Modifier != 0 {
		output += fmt.Sprintf("%+d", term.Modifier)
	}
	if term.Negative {
		output = "-" + output
	}

	return output
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run "TestCanonicalize"`  
Expected: PASS

- [ ] **Step 5: Run all package tests**

Run: `go test -v ./...`  
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add canonical.go binary_canonical_test.go
git commit -m "feat: implement canonical dice notation formatting and Canonicalize API"
```
