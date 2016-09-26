package roll

import (
	"math/rand"
	"reflect"
	"testing"
)

// Ensure the dice roll statement can get results correctly
func TestDiceRoll_Roll(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		mux  int
		mod  int
		exp  *ExplodingOp
		lmt  *LimitOp
		succ *ComparisonOp
		fail *ComparisonOp
		roll []RerollOp
		sort SortType
		res  []int
		scnt int
		totl int
	}{
		// Fate rolls
		{seed: 0, die: FateDie(0), mux: 4, mod: 2, res: []int{-1, -1, 0, 0}, totl: 0},

		// Normal Die (3)
		{seed: 1, die: NormalDie(3), mux: 3, mod: 8, res: []int{3, 1, 3}, totl: 15},

		// Normal Die (6)
		{seed: 2, die: NormalDie(6), mux: 2, mod: -2, res: []int{5, 1}, totl: 4},

		// Exploding Die (6) on 5s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res:  []int{5, 1, 1},
			totl: 7,
		},

		// Exploding Die (6) on >1s
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res:  []int{6, 4, 6, 6, 2, 1, 2, 3, 5, 1},
			totl: 36,
		},

		// Exploding Die (6) on <2s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res:  []int{5, 1, 1, 3},
			totl: 10,
		},

		// Compounded Die (6) on 5s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res:  []int{5, 1, 5},
			totl: 11,
		},

		// Compounded Die (6) on >1s
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res:  []int{6, 4, 34},
			totl: 44,
		},

		// Compounded Die (6) on <2s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res:  []int{5, 1, 2},
			totl: 8,
		},

		// Penetrating Die (6) on 5s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res:  []int{5, 1, 0},
			totl: 6,
		},

		// Penetrating Die (6) on >1s
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res:  []int{6, 4, 5, 5, 1, 0, 1, 2, 4, 0},
			totl: 28,
		},

		// Penetrating Die (6) on <2s
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res:  []int{5, 1, 0, 2},
			totl: 8,
		},

		// Limit (6) to the highest 3
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			lmt: &LimitOp{
				Type:   KeepHighest,
				Amount: 3,
			},
			res:  []int{5, 3, 1},
			totl: 9,
		},

		// Limit (6) to the lowest 3
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			lmt: &LimitOp{
				Type:   KeepLowest,
				Amount: 3,
			},
			res:  []int{3, 1, 1},
			totl: 5,
		},

		// Limit (6) and drop lowest 3
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			lmt: &LimitOp{
				Type:   DropLowest,
				Amount: 3,
			},
			res:  []int{5},
			totl: 5,
		},

		// Limit (6) and drop highest 3
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			lmt: &LimitOp{
				Type:   DropHighest,
				Amount: 3,
			},
			res:  []int{1},
			totl: 1,
		},

		// Reroll (6) on a 1
		{
			seed: 0,
			die:  NormalDie(6),
			mux:  4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 1,
					},
					Once: false,
				},
			},
			res:  []int{6, 5, 2, 5},
			totl: 18,
		},

		// Reroll (6) on values > 4
		{
			seed: 0,
			die:  NormalDie(6),
			mux:  4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  GreaterThan,
						Value: 4,
					},
					Once: false,
				},
			},
			res:  []int{1, 1, 2, 2},
			totl: 6,
		},

		// Reroll (6) on values < 3
		{
			seed: 0,
			die:  NormalDie(6),
			mux:  4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  LessThan,
						Value: 3,
					},
					Once: false,
				},
			},
			res:  []int{6, 5, 6, 5},
			totl: 22,
		},

		// Reroll (6) on values < 4, but once only
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  LessThan,
						Value: 4,
					},
					Once: true,
				},
			},
			res:  []int{5, 3, 3, 5},
			totl: 16,
		},

		// Reroll (6) on 1s and 2s, but once only
		{
			seed: 0,
			die:  NormalDie(6),
			mux:  4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 1,
					},
					Once: true,
				},
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 2,
					},
					Once: true,
				},
			},
			res:  []int{6, 5, 2, 5},
			totl: 18,
		},

		// Normal Die (3) sorted ascending
		{seed: 1, die: NormalDie(3), mux: 3, res: []int{1, 3, 3}, totl: 7, sort: Ascending},

		// Normal Die (3) sorted descending
		{seed: 1, die: NormalDie(3), mux: 3, res: []int{3, 3, 1}, totl: 7, sort: Descending},

		// Normal Die (6) successes on 6s
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: 2,
			succ: &ComparisonOp{
				Type:  Equals,
				Value: 6,
			},
			scnt: 2,
		},

		// Normal Die (6) successes on <5
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: 1,
			succ: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
			scnt: 1,
		},

		// Normal Die (6) successes on >3
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: 3,
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 3,
			},
			scnt: 3,
		},

		// Normal Die (6) successes on 6s, failures on 4s
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: 1,
			succ: &ComparisonOp{
				Type:  Equals,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 4,
			},
			scnt: 1,
		},

		// Normal Die (6) successes on <5, failures on >5
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: -1,
			succ: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
			fail: &ComparisonOp{
				Type:  GreaterThan,
				Value: 5,
			},
			scnt: -1,
		},

		// Normal Die (6) successes on >4, failures <5
		{
			seed: 1,
			die:  NormalDie(6),
			mux:  3,
			res:  []int{6, 4, 6},
			totl: 1,
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 4,
			},
			fail: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
			scnt: 1,
		},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &DiceRoll{
			Multiplier: tt.mux,
			Die:        tt.die,
			Modifier:   tt.mod,
			Exploding:  tt.exp,
			Limit:      tt.lmt,
			Success:    tt.succ,
			Failure:    tt.fail,
			Rerolls:    tt.roll,
			Sort:       tt.sort,
		}
		rolls := stmt.Roll()
		results := make([]int, len(rolls.Results))
		for i, v := range rolls.Results {
			results[i] = v.Result
		}

		if !reflect.DeepEqual(tt.res, results) {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, results)
		}

		if tt.scnt != rolls.Successes {
			t.Errorf("%d. successes mismatch: exp=%v got=%v", i, tt.scnt, rolls.Successes)
		}

		if tt.totl != rolls.Total {
			t.Errorf("%d. total mismatch: exp=%v got=%v", i, tt.totl, rolls.Total)
		}
	}
}

// Ensure the dice roll statement can be represented as a string
func TestDiceRoll_String(t *testing.T) {
	var tests = []struct {
		seed int64
		die  Die
		mux  int
		mod  int
		exp  *ExplodingOp
		lmt  *LimitOp
		succ *ComparisonOp
		fail *ComparisonOp
		roll []RerollOp
		sort SortType
		res  string
	}{
		// Fate rolls
		{seed: 0, die: FateDie(0), mux: 4, mod: 2, res: `+4dF+2`},

		// Normal Die (3)
		{seed: 1, die: NormalDie(3), mux: 3, mod: 8, res: `+3d3+8`},

		// Normal Die (6)
		{seed: 2, die: NormalDie(6), mux: 2, mod: -2, res: `+2d6-2`},

		// Normal Die (6), no modifier
		{seed: 2, die: NormalDie(6), mux: 2, mod: 0, res: `+2d6`},

		// Normal Die (20), no multiplier
		{seed: 2, die: NormalDie(20), mux: 1, mod: 2, res: `+d20+2`},

		// Normal Die (20), negative 1x multiplier
		{seed: 2, die: NormalDie(20), mux: -1, mod: 2, res: `-d20+2`},

		// Normal Die (6), negative multiplier
		{seed: 2, die: NormalDie(6), mux: -2, mod: -2, res: `-2d6-2`},

		// Exploding Die (6) on 5s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res: "+2d6!5",
		},

		// Exploding Die (6) on >1s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res: "+2d6!>1",
		},

		// Exploding Die (6) on <2s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Exploding,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res: "+2d6!<2",
		},

		// Compounded Die (6) on 5s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res: "+2d6!!5",
		},

		// Compounded Die (6) on >1s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res: "+2d6!!>1",
		},

		// Compounded Die (6) on <2s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Compounded,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res: "+2d6!!<2",
		},

		// Penetrating Die (6) on 5s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  Equals,
					Value: 5,
				},
			},
			res: "+2d6!p5",
		},

		// Penetrating Die (6) on >1s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  GreaterThan,
					Value: 1,
				},
			},
			res: "+2d6!p>1",
		},

		// Penetrating Die (6) on <2s
		{
			die: NormalDie(6),
			mux: 2,
			exp: &ExplodingOp{
				Type: Penetrating,
				ComparisonOp: &ComparisonOp{
					Type:  LessThan,
					Value: 2,
				},
			},
			res: "+2d6!p<2",
		},

		// Limit (6) to the highest 3
		{
			seed: 2,
			die:  NormalDie(6),
			mux:  4,
			lmt: &LimitOp{
				Type:   KeepHighest,
				Amount: 3,
			},
			res: "+4d6kh3",
		},

		// Limit (6) to the lowest 3
		{
			die: NormalDie(6),
			mux: 4,
			lmt: &LimitOp{
				Type:   KeepLowest,
				Amount: 3,
			},
			res: "+4d6kl3",
		},

		// Limit (6) and drop lowest 3
		{
			die: NormalDie(6),
			mux: 4,
			lmt: &LimitOp{
				Type:   DropLowest,
				Amount: 3,
			},
			res: "+4d6dl3",
		},

		// Limit (6) and drop highest 3
		{
			die: NormalDie(6),
			mux: 4,
			lmt: &LimitOp{
				Type:   DropHighest,
				Amount: 3,
			},
			res: "+4d6dh3",
		},

		// Reroll (6) on a 1
		{
			die: NormalDie(6),
			mux: 4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 1,
					},
					Once: false,
				},
			},
			res: "+4d6r1",
		},

		// Reroll (6) on values > 4
		{
			die: NormalDie(6),
			mux: 4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  GreaterThan,
						Value: 4,
					},
					Once: false,
				},
			},
			res: "+4d6r>4",
		},

		// Reroll (6) on values < 3
		{
			die: NormalDie(6),
			mux: 4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  LessThan,
						Value: 3,
					},
					Once: false,
				},
			},
			res: "+4d6r<3",
		},

		// Reroll (6) on values < 4, but once only
		{
			die: NormalDie(6),
			mux: 4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  LessThan,
						Value: 4,
					},
					Once: true,
				},
			},
			res: "+4d6ro<4",
		},

		// Reroll (6) on 1s and 2s, but once only
		{
			die: NormalDie(6),
			mux: 4,
			roll: []RerollOp{
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 1,
					},
					Once: true,
				},
				RerollOp{
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 2,
					},
					Once: true,
				},
			},
			res: "+4d6ro1ro2",
		},

		// Normal Die (3) sorted ascending
		{die: NormalDie(3), mux: 3, res: "+3d3s", sort: Ascending},

		// Normal Die (3) sorted descending
		{die: NormalDie(3), mux: 3, res: "+3d3sd", sort: Descending},

		// Normal Die (6) successes on 6s
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6=6",
			succ: &ComparisonOp{
				Type:  Equals,
				Value: 6,
			},
		},

		// Normal Die (6) successes on <5
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6<5",
			succ: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
		},

		// Normal Die (6) successes on >3
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6>3",
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 3,
			},
		},

		// Normal Die (6) successes on 6s, failures on 4s
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6=6f=4",
			succ: &ComparisonOp{
				Type:  Equals,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 4,
			},
		},

		// Normal Die (6) successes on <5, failures on >5
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6<5f>5",
			succ: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
			fail: &ComparisonOp{
				Type:  GreaterThan,
				Value: 5,
			},
		},

		// Normal Die (6) successes on >4, failures <5
		{
			die: NormalDie(6),
			mux: 3,
			res: "+3d6>4f<5",
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 4,
			},
			fail: &ComparisonOp{
				Type:  LessThan,
				Value: 5,
			},
		},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &DiceRoll{
			Multiplier: tt.mux,
			Die:        tt.die,
			Modifier:   tt.mod,
			Exploding:  tt.exp,
			Rerolls:    tt.roll,
			Success:    tt.succ,
			Failure:    tt.fail,
			Sort:       tt.sort,
			Limit:      tt.lmt,
		}
		if stmt.String() != tt.res {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, stmt.String())
		}
	}
}

// Ensure the group dice roll statement can get results correctly
func TestGroupedRoll_Roll(t *testing.T) {
	var tests = []struct {
		seed  int64
		rolls []Roll
		mod   int
		lmit  *LimitOp
		succ  *ComparisonOp
		fail  *ComparisonOp
		comb  bool
		scnt  int
		totl  int
		res   []int
	}{
		// Single roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
			},
			mod:  2,
			res:  []int{8},
			totl: 10,
		},
		// Combined single roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
			},
			mod:  2,
			res:  []int{5, 5, 6},
			totl: 18,
			comb: true,
		},
		// Combined multi-roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(4),
					Modifier:   0,
				},
			},
			mod:  0,
			res:  []int{5, 5, 6, 3, 4},
			totl: 23,
			comb: true,
		},
		// Multi-roll with limit
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  2,
			res:  []int{15},
			totl: 17,
			lmit: &LimitOp{
				Type:   KeepHighest,
				Amount: 1,
			},
		},
		// Multi-roll combined with limit
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  2,
			res:  []int{14, 11, 6, 5},
			totl: 38,
			comb: true,
			lmit: &LimitOp{
				Type:   KeepHighest,
				Amount: 4,
			},
		},
		// Multi-roll with succ/fail
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  0,
			res:  []int{8, 3, 15},
			totl: 1,
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 3,
			},
			scnt: 1,
		},
		// Multi-roll combined with succ/fail
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  0,
			res:  []int{5, 5, 6, 3, 14, 11},
			totl: 1,
			comb: true,
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 3,
			},
			scnt: 1,
		},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &GroupedRoll{
			Rolls:    tt.rolls,
			Modifier: tt.mod,
			Limit:    tt.lmit,
			Success:  tt.succ,
			Failure:  tt.fail,
			Combined: tt.comb,
		}
		rolls := stmt.Roll()
		results := make([]int, len(rolls.Results))
		for i, v := range rolls.Results {
			results[i] = v.Result
		}

		if !reflect.DeepEqual(tt.res, results) {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, results)
		}

		if tt.scnt != rolls.Successes {
			t.Errorf("%d. successes mismatch: exp=%v got=%v", i, tt.scnt, rolls.Successes)
		}

		if tt.totl != rolls.Total {
			t.Errorf("%d. total mismatch: exp=%v got=%v", i, tt.totl, rolls.Total)
		}
	}
}

// Ensure the group dice roll statement string representation is correct
func TestGroupedRoll_String(t *testing.T) {
	var tests = []struct {
		seed  int64
		rolls []Roll
		mod   int
		lmit  *LimitOp
		succ  *ComparisonOp
		fail  *ComparisonOp
		comb  bool
		res   string
	}{
		// Single roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
			},
			mod: 2,
			res: "{3d6+4,}+2",
		},
		// Combined single roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
			},
			mod:  2,
			res:  "{3d6+4}+2",
			comb: true,
		},
		// Combined multi-roll
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(4),
					Modifier:   0,
				},
			},
			mod:  0,
			res:  "{3d6+4 + 2d4}",
			comb: true,
		},
		// Multi-roll with limit
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod: 2,
			res: "{3d6+4, d4, 2d8+10}kh1+2",
			lmit: &LimitOp{
				Type:   KeepHighest,
				Amount: 1,
			},
		},
		// Multi-roll combined with limit
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  2,
			res:  "{3d6+4 + d4 + 2d8+10}kh4+2",
			comb: true,
			lmit: &LimitOp{
				Type:   KeepHighest,
				Amount: 4,
			},
		},
		// Multi-roll with succ/fail
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod: 0,
			res: "{3d6+4, d4, 2d8+10}>6f=3",
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 3,
			},
		},
		// Multi-roll combined with succ/fail
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&DiceRoll{
					Multiplier: 1,
					Die:        NormalDie(4),
					Modifier:   0,
				},
				&DiceRoll{
					Multiplier: 2,
					Die:        NormalDie(8),
					Modifier:   10,
				},
			},
			mod:  0,
			res:  "{3d6+4 + d4 + 2d8+10}>6f=3",
			comb: true,
			succ: &ComparisonOp{
				Type:  GreaterThan,
				Value: 6,
			},
			fail: &ComparisonOp{
				Type:  Equals,
				Value: 3,
			},
		},
		// Nested grouped roll combined
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&GroupedRoll{
					Rolls: []Roll{
						&DiceRoll{
							Multiplier: 3,
							Die:        NormalDie(6),
							Modifier:   4,
						},
					},
				},
			},
			mod:  2,
			res:  "{3d6+4 + {3d6+4,}}+2",
			comb: true,
		},
		// Nested grouped roll non-combined
		{
			seed: 0,
			rolls: []Roll{
				&DiceRoll{
					Multiplier: 3,
					Die:        NormalDie(6),
					Modifier:   4,
				},
				&GroupedRoll{
					Rolls: []Roll{
						&DiceRoll{
							Multiplier: 3,
							Die:        NormalDie(6),
							Modifier:   4,
						},
					},
				},
			},
			mod: 2,
			res: "{3d6+4, {3d6+4,}}+2",
		},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		stmt := &GroupedRoll{
			Rolls:    tt.rolls,
			Modifier: tt.mod,
			Limit:    tt.lmit,
			Success:  tt.succ,
			Failure:  tt.fail,
			Combined: tt.comb,
		}

		if !reflect.DeepEqual(tt.res, stmt.String()) {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.res, stmt.String())
		}
	}
}
