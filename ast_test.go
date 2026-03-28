package roll

import (
	"reflect"
	"testing"
)

func compileProgram(t *testing.T, input string) *Program {
	t.Helper()
	program, err := CompileString(input)
	if err != nil {
		t.Fatalf("compile %q: %v", input, err)
	}
	return program
}

func evaluateProgram(t *testing.T, seed int64, input string) Result {
	t.Helper()
	program := compileProgram(t, input)
	var result Result
	withTestSeed(seed, func() {
		var err error
		result, err = EvaluateProgram(program)
		if err != nil {
			t.Fatalf("evaluate %q: %v", input, err)
		}
	})
	return result
}

func TestEvaluateProgram_DiceTerms(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		res   []int
		scnt  int
		totl  int
	}{
		{name: "fate modifier", seed: 0, input: "4dF+2", res: []int{-1, -1, 0, 0}, totl: 0},
		{name: "exploding", seed: 2, input: "2d6!5", res: []int{5, 1, 1}, totl: 7},
		{name: "compounded", seed: 2, input: "2d6!!5", res: []int{5, 1, 5}, totl: 11},
		{name: "penetrating", seed: 2, input: "2d6!p5", res: []int{5, 1, 0}, totl: 6},
		{name: "keep highest", seed: 2, input: "4d6kh3", res: []int{5, 3, 1}, totl: 9},
		{name: "reroll once", seed: 2, input: "4d6ro<4", res: []int{5, 3, 3, 5}, totl: 16},
		{name: "descending sort", seed: 1, input: "3d3sd", res: []int{3, 3, 1}, totl: 7},
		{name: "success failure", seed: 1, input: "3d6=6f=4", res: []int{6, 4, 6}, scnt: 1, totl: 1},
		{name: "inclusive success", seed: 0, input: "d6>=1", res: []int{1}, scnt: 1, totl: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			values := make([]int, len(result.Results))
			for i, roll := range result.Results {
				values[i] = roll.Result
			}

			if !reflect.DeepEqual(tt.res, values) {
				t.Fatalf("results mismatch: exp=%v got=%v", tt.res, values)
			}
			if tt.scnt != result.Successes {
				t.Fatalf("success mismatch: exp=%d got=%d", tt.scnt, result.Successes)
			}
			if tt.totl != result.Total {
				t.Fatalf("total mismatch: exp=%d got=%d", tt.totl, result.Total)
			}
		})
	}
}

func TestEvaluateProgram_GroupTerms(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		res   []int
		scnt  int
		totl  int
	}{
		{name: "single grouped roll", seed: 0, input: "{3d6+4}", res: []int{5, 5, 6}, totl: 16},
		{name: "separated grouped roll", seed: 0, input: "{3d6, 2d8}", res: []int{4, 7}, totl: 11},
		{name: "combined grouped roll", seed: 0, input: "{3d6 + 2d8}", res: []int{1, 1, 2, 3, 4}, totl: 11},
		{name: "grouped successes", seed: 0, input: "{3d6 + 2d8}>3", res: []int{1, 1, 2, 3, 4}, scnt: 1, totl: 1},
		{name: "grouped success failures", seed: 0, input: "{3d6 + 2d8}>2f=1", res: []int{1, 1, 2, 3, 4}, scnt: 0, totl: 0},
		{name: "nested combined", seed: 0, input: "{3d6+2d8-{4d4-1}dl}kh3<4f>3", res: []int{4, 3, 2}, scnt: 1, totl: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			values := make([]int, len(result.Results))
			for i, roll := range result.Results {
				values[i] = roll.Result
			}

			if !reflect.DeepEqual(tt.res, values) {
				t.Fatalf("results mismatch: exp=%v got=%v", tt.res, values)
			}
			if tt.scnt != result.Successes {
				t.Fatalf("success mismatch: exp=%d got=%d", tt.scnt, result.Successes)
			}
			if tt.totl != result.Total {
				t.Fatalf("total mismatch: exp=%d got=%d", tt.totl, result.Total)
			}
		})
	}
}

func TestProgram_String(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "3d6+4", want: "3d6+4"},
		{input: "6d6!!5kh3sd+3", want: "6d6+3!!5kh3sd"},
		{input: "6d6sa>=5", want: "6d6>=5s"},
		{input: "{3d6+4,2d8}dl=1f>5", want: "{3d6+4, 2d8}dl=1f>5"},
		{input: "{3d6+2d8-{4d4-1}dl}kh3<4f>3", want: "{3d6 + 2d8 - {4d4-1}dl}kh3<4f>3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := compileProgram(t, tt.input)
			if got := program.String(); got != tt.want {
				t.Fatalf("program string mismatch: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestComparisonOp_MatchInclusive(t *testing.T) {
	tests := []struct {
		name string
		op   *ComparisonOp
		val  int
		want bool
	}{
		{name: "greater inclusive", op: &ComparisonOp{Type: GreaterThan, Value: 4, Inclusive: true}, val: 4, want: true},
		{name: "less inclusive", op: &ComparisonOp{Type: LessThan, Value: 2, Inclusive: true}, val: 2, want: true},
		{name: "greater strict", op: &ComparisonOp{Type: GreaterThan, Value: 4}, val: 4, want: false},
	}

	for _, tt := range tests {
		if got := tt.op.Match(tt.val); got != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.want, got)
		}
	}
}

func TestEvaluateProgramWithLimits(t *testing.T) {
	t.Run("rejects unsafe die built in bytecode", func(t *testing.T) {
		program := &Program{
			Code:      []Instruction{{Op: OpRollDice, Arg: 0}},
			DiceTerms: []DiceTerm{{Multiplier: 1, Die: NormalDie(1)}},
			Rendered:  "d1",
			MaxDepth:  1,
		}

		_, err := EvaluateProgramWithLimits(program, DefaultLimits)
		if err == nil {
			t.Fatal("expected unsafe die error")
		}
		if err.Error() != `unsafe die type "d1"` {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("enforces evaluation depth", func(t *testing.T) {
		program := &Program{
			Code: []Instruction{
				{Op: OpRollDice, Arg: 0},
				{Op: OpRollGroup, Arg: 0},
				{Op: OpRollGroup, Arg: 1},
			},
			DiceTerms: []DiceTerm{{Multiplier: 1, Die: NormalDie(6)}},
			GroupTerms: []GroupTerm{
				{Combined: true, ChildCount: 1},
				{Combined: true, ChildCount: 1},
			},
			Rendered: "{{d6}}",
			MaxDepth: 3,
		}

		withTestSeed(0, func() {
			_, err := EvaluateProgramWithLimits(program, Limits{
				MaxDieSize:     100,
				MaxRollsPerDie: 10,
				MaxRollsTotal:  10,
				MaxEvalDepth:   2,
			})
			if err == nil {
				t.Fatal("expected depth error")
			}
			if err.Error() != "roll exceeded maximum evaluation depth of 2" {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	})
}
