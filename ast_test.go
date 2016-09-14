package roll

import (
	"math/rand"
	"reflect"
	"testing"
)

// Ensure the dice roll statement can get results correctly
func TestDiceRollStmt_Roll(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		mux  int
		mod  int
		res  []int
	}{
		// Fate rolls
		{seed: 0, die: FateDie(0), mux: 4, mod: 2, res: []int{-1, -1, 0, 0}},

		// Normal Die (3)
		{seed: 1, die: NormalDie(3), mux: 3, mod: 8, res: []int{3, 1, 3}},

		// Normal Die (6)
		{seed: 2, die: NormalDie(6), mux: 2, mod: -2, res: []int{5, 1}},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &DiceRollStmt{
			Multiplier: tt.mux,
			Die:        tt.die,
			Modifier:   tt.mod,
		}
		rolls := stmt.Roll()
		results := make([]int, len(rolls))
		for i, v := range rolls {
			results[i] = v.Result
		}

		if !reflect.DeepEqual(tt.res, results) {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, results)
		}
	}
}

// Ensure the dice roll statement can be represented as a string
func TestDiceRollStmt_String(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		mux  int
		mod  int
		res  string
	}{
		// Fate rolls
		{seed: 0, die: FateDie(0), mux: 4, mod: 2, res: `4dF+2`},

		// Normal Die (3)
		{seed: 1, die: NormalDie(3), mux: 3, mod: 8, res: `3d3+8`},

		// Normal Die (6)
		{seed: 2, die: NormalDie(6), mux: 2, mod: -2, res: `2d6-2`},

		// Normal Die (6), no modifier
		{seed: 2, die: NormalDie(6), mux: 2, mod: 0, res: `2d6`},

		// Normal Die (20), no multiplier
		{seed: 2, die: NormalDie(20), mux: 1, mod: 2, res: `d20+2`},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &DiceRollStmt{
			Multiplier: tt.mux,
			Die:        tt.die,
			Modifier:   tt.mod,
		}
		if stmt.String() != tt.res {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, stmt.String())
		}
	}
}
