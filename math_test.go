package roll

import "testing"

func TestRoll20Spec_Multiplication(t *testing.T) {
	tests := []roll20TestCase{
		{name: "dice times constant 2d6*3", seed: 2, input: "2d6*3", wantString: "2d6*3", wantRolls: []int{5, 1}, wantTotal: 18},
		{name: "constant times constant 5*3", seed: 0, input: "5*3", wantString: "5*3", wantRolls: []int{}, wantTotal: 15},
		{name: "parenthesised roll (2d6+3)*4", seed: 2, input: "(2d6+3)*4", wantString: "(2d6+3)*4", wantRolls: []int{5, 1}, wantTotal: 36},
		{name: "multiplication before addition 2d6*3+4", seed: 2, input: "2d6*3+4", wantString: "2d6*3+4", wantRolls: []int{5, 1}, wantTotal: 22},
		{name: "product added to roll 2d6+4*3", seed: 2, input: "2d6+4*3", wantString: "2d6+4*3", wantRolls: []int{5, 1}, wantTotal: 18},
		{name: "chained multiplication 2d6*2*3", seed: 2, input: "2d6*2*3", wantString: "2d6*2*3", wantRolls: []int{5, 1}, wantTotal: 36},
	}
	runRoll20TestCases(t, tests)
}
