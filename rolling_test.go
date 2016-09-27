package roll

import (
	"math/rand"
	"strings"
	"testing"
)

// Ensure the parser works
func TestParse(t *testing.T) {
	var tests = []struct {
		seed int64
		in   string
		out  string
	}{
		// Fate rolls
		{seed: 0, in: "2dF+4", out: `Rolled "2dF+4" and got ⊟, ⊟ for a total of 2`},
		{seed: 0, in: "dF", out: `Rolled "dF" and got ⊟ for a total of -1`},

		// Normal Die (3)
		{seed: 0, in: "3d3-2", out: `Rolled "3d3-2" and got 1, 1, 2 for a total of 2`},
		{seed: 0, in: "d3+1", out: `Rolled "d3+1" and got 1 for a total of 2`},

		// Normal Die (6)
		{seed: 0, in: "3d6+4", out: `Rolled "3d6+4" and got 1, 1, 2 for a total of 8`},
		{seed: 0, in: "d6-1", out: `Rolled "d6-1" and got 1 for a total of 0`},

		// Normal Die (6) Successes
		{seed: 0, in: "3d6+4>4f=6", out: `Rolled "3d6+4>4f=6" and got 1, 1, 2 for a total of 2`},
		{seed: 0, in: "d6-1<3", out: `Rolled "d6-1<3" and got 1 for a total of 1`},

		// Grouped rolls
		{seed: 0, in: "{3d6+4}", out: `Rolled "{3d6+4}" and got 5, 5, 6 for a total of 16`},
		{seed: 0, in: "{3d6, 2d8}", out: `Rolled "{3d6, 2d8}" and got 4, 7 for a total of 11`},
		{seed: 0, in: "{3d6 + 2d8}", out: `Rolled "{3d6 + 2d8}" and got 1, 1, 2, 3, 4 for a total of 11`},

		// Grouped Successes
		{seed: 0, in: "{3d6 + 2d8}>3", out: `Rolled "{3d6 + 2d8}>3" and got 1, 1, 2, 3, 4 for a total of 1`},
		{seed: 0, in: "{3d6 + 2d8}>2f=1", out: `Rolled "{3d6 + 2d8}>2f=1" and got 1, 1, 2, 3, 4 for a total of 0`},

		// Errors
		{seed: 0, in: "3dX-2", out: `unrecognised die type "dX"`},
		{seed: 0, in: "CRAP", out: `found unexpected token "C"`},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		out, err := Parse(strings.NewReader(tt.in))
		if err != nil {
			if err.Error() != tt.out {
				t.Errorf("%d. unexpected parse error: exp=%v got=%v", i, tt.out, err)
			}
		} else if tt.out != out {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.out, out)
		}
	}
}

// Ensure the parser works
func TestParseString(t *testing.T) {
	var tests = []struct {
		seed int64
		in   string
		out  string
	}{
		// Fate rolls
		{seed: 0, in: "2dF+4", out: `Rolled "2dF+4" and got ⊟, ⊟ for a total of 2`},
		{seed: 0, in: "dF", out: `Rolled "dF" and got ⊟ for a total of -1`},

		// Normal Die (3)
		{seed: 0, in: "3d3-2", out: `Rolled "3d3-2" and got 1, 1, 2 for a total of 2`},
		{seed: 0, in: "d3+1", out: `Rolled "d3+1" and got 1 for a total of 2`},

		// Normal Die (6)
		{seed: 0, in: "3d6+4", out: `Rolled "3d6+4" and got 1, 1, 2 for a total of 8`},
		{seed: 0, in: "d6-1", out: `Rolled "d6-1" and got 1 for a total of 0`},

		// Errors
		{seed: 0, in: "3dX-2", out: `unrecognised die type "dX"`},
		{seed: 0, in: "CRAP", out: `found unexpected token "C"`},
	}

	for i, tt := range tests {
		rand.Seed(tt.seed)
		out, err := ParseString(tt.in)
		if err != nil {
			if err.Error() != tt.out {
				t.Errorf("%d. unexpected parse error: exp=%v got=%v", i, tt.out, err)
			}
		} else if tt.out != out {
			t.Errorf("%d. result mismatch: exp=%v got=%v", i, tt.out, out)
		}
	}
}
