package roll

import (
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
			var res1, res2 Result
			var err1, err2 error
			withTestSeed(0, func() {
				res1, err1 = EvaluateProgram(prog1)
			})
			withTestSeed(0, func() {
				res2, err2 = EvaluateProgram(prog2)
			})
			if err1 != nil || err2 != nil {
				t.Fatalf("eval errors: %v / %v", err1, err2)
			}
			if !reflect.DeepEqual(res1, res2) {
				t.Fatalf("eval result mismatch:\nexp %#v\ngot %#v", res1, res2)
			}
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
