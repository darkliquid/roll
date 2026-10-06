package roll

import "testing"

// TestRoll20Spec_OneSidedDie covers deterministic one-sided dice, which are folded
// to their fixed value instead of being rolled so that rerolls and explosions
// cannot loop forever.
func TestRoll20Spec_OneSidedDie(t *testing.T) {
	ones := []int{1, 1, 1, 1}
	nines := []int{1, 1, 1, 1, 1, 1, 1, 1, 1}
	tests := []roll20TestCase{
		{name: "single one-sided die", seed: 0, input: "d1", wantString: "d1", wantRolls: []int{1}, wantTotal: 1},
		{name: "multiple one-sided dice", seed: 0, input: "4d1", wantString: "4d1", wantRolls: ones, wantTotal: 4},
		{name: "one-sided die with modifier", seed: 0, input: "2d1+5", wantString: "2d1+5", wantRolls: []int{1, 1}, wantTotal: 7},
		{name: "one-sided die with success", seed: 0, input: "4d1>0", wantString: "4d1>0", wantRolls: ones, wantScnt: 4, wantTotal: 4},
		{name: "one-sided die with keep", seed: 0, input: "4d1kh3", wantString: "4d1kh3", wantRolls: []int{1, 1, 1}, wantTotal: 3},
		{name: "exploding one-sided die does not loop", seed: 0, input: "4d1!", wantString: "4d1!", wantRolls: ones, wantTotal: 4},
		{name: "rerolling one-sided die does not loop", seed: 0, input: "4d1r1", wantString: "4d1r1", wantRolls: ones, wantTotal: 4},
		{name: "computed one-sided die", seed: 0, input: "9d(floor(3/2))", wantString: "9d1", wantRolls: nines, wantTotal: 9},
		{name: "runtime one-sided die", seed: 0, input: "9d(1d6)", wantString: "9d(d6)", wantRolls: nines, wantTotal: 9},
	}
	runRoll20TestCases(t, tests)
}
