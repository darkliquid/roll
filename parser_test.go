package roll

import (
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *DiceRollStmt
		err  string
	}{
		// Simple roll
		{
			s: `3d6`,
			stmt: &DiceRollStmt{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   0,
			},
		},

		// Fate roll statement
		{
			s: `4dF`,
			stmt: &DiceRollStmt{
				Multiplier: 4,
				Die:        FateDie(0),
				Modifier:   0,
			},
		},

		// Simple roll with modifier
		{
			s: `3d6+4`,
			stmt: &DiceRollStmt{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   4,
			},
		},

		// Fate roll with modifier
		{
			s: `3dF+4`,
			stmt: &DiceRollStmt{
				Multiplier: 3,
				Die:        FateDie(0),
				Modifier:   4,
			},
		},

		// Simple roll with multiple modifiers
		{
			s: `3d6+4-1+6-3`,
			stmt: &DiceRollStmt{
				Multiplier: 3,
				Die:        NormalDie(6),
				Modifier:   6,
			},
		},

		// Simple roll with no multiplier
		{
			s: `d6`,
			stmt: &DiceRollStmt{
				Multiplier: 1,
				Die:        NormalDie(6),
				Modifier:   0,
			},
		},

		// Errors
		{s: `foo`, err: `found unexpected token "f"`},
		{s: `dX`, err: `unrecognised die type "dX"`},
		{s: `d4--`, err: `found unexpected token "-"`},
		{s: `3d4d5`, err: `found unexpected token "d5"`},
	}

	for i, tt := range tests {
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
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
