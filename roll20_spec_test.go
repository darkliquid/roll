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

func TestRoll20Spec_ExplodingDice(t *testing.T) {
	tests := []roll20TestCase{
		{name: "exploding dice 2d6!5", seed: 2, input: "2d6!5", wantString: "2d6!5", wantRolls: []int{5, 1, 1}, wantTotal: 7},
		{name: "exploding greater than 2d6!>4", seed: 2, input: "2d6!>4", wantString: "2d6!>4", wantRolls: []int{5, 1, 1}, wantTotal: 7},
		{name: "compounded exploding 2d6!!5", seed: 2, input: "2d6!!5", wantString: "2d6!!5", wantRolls: []int{5, 1, 5}, wantTotal: 11},
		{name: "compounded greater than 2d6!!>4", seed: 2, input: "2d6!!>4", wantString: "2d6!!>4", wantRolls: []int{5, 1, 5}, wantTotal: 11},
		{name: "penetrating exploding 2d6!p5", seed: 2, input: "2d6!p5", wantString: "2d6!p5", wantRolls: []int{5, 1, 0}, wantTotal: 6},
		{name: "penetrating greater than 2d6!p>4", seed: 2, input: "2d6!p>4", wantString: "2d6!p>4", wantRolls: []int{5, 1, 0}, wantTotal: 6},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_KeepDrop(t *testing.T) {
	tests := []roll20TestCase{
		{name: "keep highest 4d6kh3", seed: 2, input: "4d6kh3", wantString: "4d6kh3", wantRolls: []int{5, 3, 1}, wantTotal: 9},
		{name: "keep lowest 4d6kl2", seed: 2, input: "4d6kl2", wantString: "4d6kl2", wantRolls: []int{1, 1}, wantTotal: 2},
		{name: "drop highest 4d6dh1", seed: 2, input: "4d6dh1", wantString: "4d6dh", wantRolls: []int{3, 1, 1}, wantTotal: 5},
		{name: "drop lowest 4d6dl1", seed: 2, input: "4d6dl1", wantString: "4d6dl", wantRolls: []int{5, 3, 1}, wantTotal: 9},
		{name: "keep highest with modifier 4d6kh3+2", seed: 2, input: "4d6kh3+2", wantString: "4d6+2kh3", wantRolls: []int{5, 3, 1}, wantTotal: 11},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Reroll(t *testing.T) {
	tests := []roll20TestCase{
		{name: "reroll once 4d6ro<4", seed: 2, input: "4d6ro<4", wantString: "4d6ro<4", wantRolls: []int{5, 3, 3, 5}, wantTotal: 16},
		{name: "reroll once exact 4d6ro1", seed: 2, input: "4d6ro1", wantString: "4d6ro1", wantRolls: []int{5, 3, 3, 3}, wantTotal: 14},
		{name: "recursive reroll 4d6r1", seed: 2, input: "4d6r1", wantString: "4d6r1", wantRolls: []int{5, 3, 3, 3}, wantTotal: 14},
		{name: "recursive reroll comparison 4d6r<2", seed: 2, input: "4d6r<2", wantString: "4d6r<2", wantRolls: []int{5, 3, 3, 3}, wantTotal: 14},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Sorting(t *testing.T) {
	tests := []roll20TestCase{
		{name: "ascending sort 3d3s", seed: 1, input: "3d3s", wantString: "3d3s", wantRolls: []int{1, 3, 3}, wantTotal: 7},
		{name: "ascending sort alias 3d3sa", seed: 1, input: "3d3sa", wantString: "3d3s", wantRolls: []int{1, 3, 3}, wantTotal: 7},
		{name: "descending sort 3d3sd", seed: 1, input: "3d3sd", wantString: "3d3sd", wantRolls: []int{3, 3, 1}, wantTotal: 7},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_GroupedRolls(t *testing.T) {
	tests := []roll20TestCase{
		{name: "single grouped roll {3d6+4}", seed: 0, input: "{3d6+4}", wantString: "{3d6+4}", wantRolls: []int{5, 5, 6}, wantTotal: 16},
		{name: "separated grouped roll {3d6, 2d8}", seed: 0, input: "{3d6, 2d8}", wantString: "{3d6, 2d8}", wantRolls: []int{4, 7}, wantTotal: 11},
		{name: "combined grouped roll {3d6 + 2d8}", seed: 0, input: "{3d6 + 2d8}", wantString: "{3d6 + 2d8}", wantRolls: []int{1, 1, 2, 3, 4}, wantTotal: 11},
		{name: "group addition modifier {3d6, 2d8}+5", seed: 0, input: "{3d6, 2d8}+5", wantString: "{3d6, 2d8}+5", wantRolls: []int{4, 7}, wantTotal: 16},
		{name: "group subtraction modifier {3d6 + 2d8}-3", seed: 0, input: "{3d6 + 2d8}-3", wantString: "{3d6 + 2d8}-3", wantRolls: []int{1, 1, 2, 3, 4}, wantTotal: 8},
		{name: "negated grouped roll -{3d6, 2d8}", seed: 0, input: "-{3d6, 2d8}", wantString: "-{3d6, 2d8}", wantRolls: []int{-4, -7}, wantTotal: -11},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_GroupingModifiers(t *testing.T) {
	tests := []roll20TestCase{
		{name: "group keep highest {3d6, 2d8}kh1", seed: 0, input: "{3d6, 2d8}kh1", wantString: "{3d6, 2d8}kh", wantRolls: []int{7}, wantTotal: 7},
		{name: "group drop lowest {3d6, 2d8}dl1", seed: 0, input: "{3d6, 2d8}dl1", wantString: "{3d6, 2d8}dl", wantRolls: []int{7}, wantTotal: 7},
		{name: "grouped successes {3d6 + 2d8}>3", seed: 0, input: "{3d6 + 2d8}>3", wantString: "{3d6 + 2d8}>3", wantRolls: []int{1, 1, 2, 3, 4}, wantScnt: 1, wantTotal: 1},
		{name: "grouped success and failures {3d6 + 2d8}>2f=1", seed: 0, input: "{3d6 + 2d8}>2f=1", wantString: "{3d6 + 2d8}>2f=1", wantRolls: []int{1, 1, 2, 3, 4}, wantScnt: 0, wantTotal: 0},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_ComplexPermutations(t *testing.T) {
	tests := []roll20TestCase{
		{name: "multi modifier dice roll 6d6!!5kh3sd+3", seed: 2, input: "6d6!!5kh3sd+3", wantString: "6d6+3!!5kh3sd", wantRolls: []int{10, 5, 3}, wantTotal: 21},
		{name: "complex nested combined groups", seed: 0, input: "{3d6+2d8-{4d4-1}dl}kh3<4f>3", wantString: "{3d6 + 2d8 - {4d4-1}dl}kh3<4f>3", wantRolls: []int{4, 3, 2}, wantScnt: 1, wantTotal: 1},
		{name: "separated group with limit and failure {3d6+4, 2d8}dl=1f>5", seed: 0, input: "{3d6+4, 2d8}dl=1f>5", wantString: "{3d6+4, 2d8}dl=1f>5", wantRolls: []int{8}, wantScnt: -1, wantTotal: -1},
	}
	runRoll20TestCases(t, tests)
}
