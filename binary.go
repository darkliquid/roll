package roll

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	binaryMagic0  byte = 0x52 // 'R'
	binaryMagic1  byte = 0x4F // 'O'
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
	if codeLen > uint64(r.Len()) {
		return fmt.Errorf("codeLen %d exceeds remaining data length %d", codeLen, r.Len())
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
	if diceLen > uint64(r.Len()) {
		return fmt.Errorf("diceLen %d exceeds remaining data length %d", diceLen, r.Len())
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
	if groupLen > uint64(r.Len()) {
		return fmt.Errorf("groupLen %d exceeds remaining data length %d", groupLen, r.Len())
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
	if len(code) == 0 {
		p.Code = nil
	} else {
		p.Code = code
	}
	if len(diceTerms) == 0 {
		p.DiceTerms = nil
	} else {
		p.DiceTerms = diceTerms
	}
	if len(groupTerms) == 0 {
		p.GroupTerms = nil
	} else {
		p.GroupTerms = groupTerms
	}

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
		if rrLen > uint64(r.Len()) {
			return dt, fmt.Errorf("rrLen %d exceeds remaining data length %d", rrLen, r.Len())
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
		return
	}
	buf.WriteByte(1)
	buf.WriteByte(byte(cmp.Type))
	var flags byte
	if cmp.Inclusive {
		flags |= 1
	}
	buf.WriteByte(flags)
	writeVarint(buf, int64(cmp.Value))
}

func unmarshalComparisonOp(r *bytes.Reader) (*ComparisonOp, error) {
	presence, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if presence == 0 {
		return nil, nil
	}
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
