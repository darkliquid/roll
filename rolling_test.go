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

		// Errors
		{seed: 0, in: "3dX-2", out: `unrecognised die type "X"`},
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
		{seed: 0, in: "3dX-2", out: `unrecognised die type "X"`},
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
