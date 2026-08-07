package roll

import (
	"reflect"
	"testing"
)

type roll20TestCase struct {
	name       string
	seed       int64
	input      string
	wantString string
	wantRolls  []int
	wantTotal  int
	wantScnt   int
}

func runRoll20TestCases(t *testing.T, tests []roll20TestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := CompileString(tt.input)
			if err != nil {
				t.Fatalf("compile %q error: %v", tt.input, err)
			}

			if got := program.String(); got != tt.wantString {
				t.Errorf("string normalization mismatch for %q: got %q, want %q", tt.input, got, tt.wantString)
			}

			var result Result
			withTestSeed(tt.seed, func() {
				result, err = EvaluateProgram(program)
			})
			if err != nil {
				t.Fatalf("evaluate %q error: %v", tt.input, err)
			}

			values := make([]int, len(result.Results))
			for i, roll := range result.Results {
				values[i] = roll.Result
			}

			if !reflect.DeepEqual(tt.wantRolls, values) {
				t.Errorf("%q rolls mismatch: got %v, want %v", tt.input, values, tt.wantRolls)
			}
			if result.Successes != tt.wantScnt {
				t.Errorf("%q success count mismatch: got %d, want %d", tt.input, result.Successes, tt.wantScnt)
			}
			if result.Total != tt.wantTotal {
				t.Errorf("%q total mismatch: got %d, want %d", tt.input, result.Total, tt.wantTotal)
			}
		})
	}
}

func TestRoll20Spec_BasicDice(t *testing.T) {
	tests := []roll20TestCase{
		{name: "standard d20", seed: 0, input: "d20", wantString: "d20", wantRolls: []int{15}, wantTotal: 15},
		{name: "3d6 standard", seed: 0, input: "3d6", wantString: "3d6", wantRolls: []int{1, 1, 2}, wantTotal: 4},
		{name: "percentile d%", seed: 0, input: "d%", wantString: "d%", wantRolls: []int{75}, wantTotal: 75},
		{name: "multiple percentile 2d%", seed: 0, input: "2d%", wantString: "2d%", wantRolls: []int{75, 15}, wantTotal: 90},
		{name: "fate dice 4dF", seed: 0, input: "4dF", wantString: "4dF", wantRolls: []int{-1, -1, 0, 0}, wantTotal: -2},
		{name: "fate dice with positive modifier 4dF+2", seed: 0, input: "4dF+2", wantString: "4dF+2", wantRolls: []int{-1, -1, 0, 0}, wantTotal: 0},
		{name: "dice with addition 3d6+4", seed: 0, input: "3d6+4", wantString: "3d6+4", wantRolls: []int{1, 1, 2}, wantTotal: 8},
		{name: "dice with subtraction 2d8-2", seed: 0, input: "2d8-2", wantString: "2d8-2", wantRolls: []int{3, 3}, wantTotal: 4},
		{name: "dice without modifier 2d8", seed: 0, input: "2d8", wantString: "2d8", wantRolls: []int{3, 3}, wantTotal: 6},
		{name: "positive multiplier +3d6", seed: 0, input: "+3d6", wantString: "3d6", wantRolls: []int{1, 1, 2}, wantTotal: 4},
		{name: "negative multiplier -2d8", seed: 0, input: "-2d8", wantString: "-2d8", wantRolls: []int{3, 3}, wantTotal: -6},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_TargetNumbers(t *testing.T) {
	tests := []roll20TestCase{
		{name: "strict greater than success 3d6>4", seed: 0, input: "3d6>4", wantString: "3d6>4", wantRolls: []int{1, 1, 2}, wantScnt: 0, wantTotal: 0},
		{name: "inclusive greater than success 3d6>=1", seed: 0, input: "3d6>=1", wantString: "3d6>=1", wantRolls: []int{1, 1, 2}, wantScnt: 3, wantTotal: 3},
		{name: "strict less than success 3d6<3", seed: 0, input: "3d6<3", wantString: "3d6<3", wantRolls: []int{1, 1, 2}, wantScnt: 3, wantTotal: 3},
		{name: "inclusive less than success 3d6<=1", seed: 0, input: "3d6<=1", wantString: "3d6<=1", wantRolls: []int{1, 1, 2}, wantScnt: 2, wantTotal: 2},
		{name: "exact match success 3d6=2", seed: 0, input: "3d6=2", wantString: "3d6=2", wantRolls: []int{1, 1, 2}, wantScnt: 1, wantTotal: 1},
		{name: "success and failure checks 3d6=6f=1", seed: 0, input: "3d6=6f=1", wantString: "3d6=6f=1", wantRolls: []int{1, 1, 2}, wantScnt: -2, wantTotal: -2},
		{name: "failure range 5d6>4f<2", seed: 0, input: "5d6>4f<2", wantString: "5d6>4f<2", wantRolls: []int{1, 1, 2, 5, 6}, wantScnt: 0, wantTotal: 0},
		{name: "target number with term modifier 3d6+2>3", seed: 0, input: "3d6+2>3", wantString: "3d6+2>3", wantRolls: []int{1, 1, 2}, wantScnt: 1, wantTotal: 1},
	}
	runRoll20TestCases(t, tests)
}
