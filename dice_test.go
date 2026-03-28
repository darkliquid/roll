package roll

import (
	"math/rand"
	"testing"
)

func withTestSeed(seed int64, fn func()) {
	prev := randomIntn
	r := rand.New(rand.NewSource(seed))
	randomIntn = r.Intn
	defer func() {
		randomIntn = prev
	}()
	fn()
}

// Ensure the die can roll results correctly.
func TestDie_Roll(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		res  int
		sym  string
	}{
		// Fate rolls
		{seed: 2, die: FateDie(0), res: 0, sym: FateBlank},
		{seed: 1, die: FateDie(0), res: 1, sym: FatePlus},
		{seed: 0, die: FateDie(0), res: -1, sym: FateMinus},

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

		// Percentile Die
		{seed: 0, die: PercentileDie(0), res: 75, sym: "75"},
	}

	for i, tt := range tests {
		withTestSeed(tt.seed, func() {
			result := tt.die.Roll()
			if tt.res != result.Result {
				t.Errorf("%d. result mismatch: exp=%d got=%d", i, tt.res, result.Result)
			} else if tt.sym != result.Symbol {
				t.Errorf("%d. symbol mismatch: exp=%q got=%q", i, tt.sym, result.Symbol)
			}
		})
	}
}
