package roll

import (
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_Parse(t *testing.T) {
	var tests = []struct {
		s    string
		roll Roll
		err  string
	}{
		// Simple roll
		{
			s: `3d6`,
			roll: &DiceRoll{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   0,
			},
		},

		// Fate roll statement
		{
			s: `4dF`,
			roll: &DiceRoll{
				Multiplier: 4,
				Die:        FateDie(0),
				Modifier:   0,
			},
		},

		// Simple roll with modifier
		{
			s: `3d6+4`,
			roll: &DiceRoll{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   4,
			},
		},

		// Fate roll with modifier
		{
			s: `3dF+4`,
			roll: &DiceRoll{
				Multiplier: 3,
				Die:        FateDie(0),
				Modifier:   4,
			},
		},

		// Simple roll with multiple modifiers
		{
			s: `3d6+4-1+6-3`,
			roll: &DiceRoll{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   6,
			},
		},

		// Simple roll with no multiplier
		{
			s: `d6`,
			roll: &DiceRoll{
				Multiplier: 1,
				Die:        NormalDie(6),
				Modifier:   0,
			},
		},

		// Multi-roll, compounded on 5s, keep top 3, sort descending, +3
		{
			s: `6d6!!5kh3sd+3`,
			roll: &DiceRoll{
				Multiplier: 6,
				Die:        NormalDie(6),
				Modifier:   3,
				Sort:       Descending,
				Limit: &LimitOp{
					Type:   KeepHighest,
					Amount: 3,
				},
				Exploding: &ExplodingOp{
					Type: Compounded,
					ComparisonOp: &ComparisonOp{
						Type:  Equals,
						Value: 5,
					},
				},
			},
		},

		// Multi-roll, reroll 2s, reroll once on 4s, successes > 3, failures on 1s
		{
			s: `6d6r2ro4>3f=1`,
			roll: &DiceRoll{
				Multiplier: 6,
				Die:        NormalDie(6),
				Rerolls: []RerollOp{
					RerollOp{
						ComparisonOp: &ComparisonOp{
							Type:  Equals,
							Value: 2,
						},
					},
					RerollOp{
						ComparisonOp: &ComparisonOp{
							Type:  Equals,
							Value: 4,
						},
						Once: true,
					},
				},
				Success: &ComparisonOp{
					Type:  GreaterThan,
					Value: 3,
				},
				Failure: &ComparisonOp{
					Type:  Equals,
					Value: 1,
				},
			},
		},

		// Grouped multi-roll, drop lowest, successes on 1s, fails > 5
		{
			s: `{3d6+4,2d8}dl=1f>5`,
			roll: &GroupedRoll{
				Rolls: []Roll{
					&DiceRoll{
						Multiplier: 3,
						Die:        NormalDie(6),
						Modifier:   4,
					},
					&DiceRoll{
						Multiplier: 2,
						Die:        NormalDie(8),
					},
				},
				Limit: &LimitOp{
					Amount: 1,
					Type:   DropLowest,
				},
				Success: &ComparisonOp{
					Type:  Equals,
					Value: 1,
				},
				Failure: &ComparisonOp{
					Type:  GreaterThan,
					Value: 5,
				},
				Combined: false,
			},
		},

		// Grouped combined nested multi-roll, keep high 3, succ <4, fail >3
		{
			s: `{3d6+2d8-{4d4-1}dl}kh3<4f>3`,
			roll: &GroupedRoll{
				Rolls: []Roll{
					&DiceRoll{
						Multiplier: 3,
						Die:        NormalDie(6),
					},
					&DiceRoll{
						Multiplier: 2,
						Die:        NormalDie(8),
					},
					&GroupedRoll{
						Rolls: []Roll{
							&DiceRoll{
								Multiplier: 4,
								Die:        NormalDie(4),
								Modifier:   -1,
							},
						},
						Limit: &LimitOp{
							Amount: 1,
							Type:   DropLowest,
						},
						Combined: true,
						Negative: true,
					},
				},
				Limit: &LimitOp{
					Amount: 3,
					Type:   KeepHighest,
				},
				Success: &ComparisonOp{
					Type:  LessThan,
					Value: 4,
				},
				Failure: &ComparisonOp{
					Type:  GreaterThan,
					Value: 3,
				},
				Combined: true,
			},
		},

		// Errors
		{s: `foo`, err: `found unexpected token "f"`},
		{s: `dX`, err: `unrecognised die type "dX"`},
		{s: `d4--`, err: `found unexpected token "-"`},
		{s: `3d4d5`, err: `found unexpected token "d5"`},
	}

	for i, tt := range tests {
		roll, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.roll, roll) {
			t.Errorf("%d. %q\n\nroll mismatch:\n\nexp=%#v (%s)\n\ngot=%#v (%s)\n\n", i, tt.s, tt.roll, tt.roll.String(), roll, roll.String())
		}
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
