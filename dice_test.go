package roll

import (
	"math/rand"
	"testing"
)

// Ensure the die can roll results correctly.
func TestDie_Roll(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		res  int
		sym  string
	}{
		// Fate rolls
		{seed: 2, die: FateDie(0), res: 0, sym: FATE_BLANK},
		{seed: 1, die: FateDie(0), res: 1, sym: FATE_PLUS},
		{seed: 0, die: FateDie(0), res: -1, sym: FATE_MINUS},

		// Normal Die (3)
		{seed: 0, die: NormalDie(3), res: 1, sym: "1"},
		{seed: 2, die: NormalDie(3), res: 2, sym: "2"},
		{seed: 1, die: NormalDie(3), res: 3, sym: "3"},

		// Normal Die (6)
		{seed: 0, die: NormalDie(6), res: 1, sym: "1"},
		{seed: 4, die: NormalDie(6), res: 2, sym: "2"},
		{seed: 7, die: NormalDie(6), res: 3, sym: "3"},
		{seed: 12, die: NormalDie(6), res: 4, sym: "4"},
		{seed: 3, die: NormalDie(6), res: 5, sym: "5"},
		{seed: 1, die: NormalDie(6), res: 6, sym: "6"},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		result := tt.die.Roll()
		if tt.res != result.Result {
			t.Errorf("%d. result mismatch: exp=%d got=%d", i, tt.res, result.Result)
		} else if tt.sym != result.Symbol {
			t.Errorf("%d. symbol mismatch: exp=%q got=%q", i, tt.sym, result.Symbol)
		}
	}
}
