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
		"2d6!",
		"2d6*3+4",
		"(2d6+3)*4",
		"floor(11/2)+1d6",
		"4d6mt3>5",
		"2d6m",
		"3d(4d(d8)kh3+2)",
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
		{input: "2d6*3", want: "2d6*3"},
		{input: "2d6*3+4", want: "2d6*3+4"},
		{input: "(2d6+3)*4", want: "(2d6+3)*4"},
		{input: "2d6+4*3", want: "3*4+2d6"},
		{input: "floor(11/2)", want: "floor(11/2)"},
		{input: "4d6mt3>5", want: "4d6mt3>5"},
		{input: "2d6m", want: "2d6m"},
		{input: "9d(1d6)", want: "9d(d6)"},
		{input: "3d(4d(d8)kh3+2)", want: "3d(4d(d8)kh3+2)"},
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
