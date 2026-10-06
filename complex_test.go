package roll

import "testing"

// TestRoll20Spec_ComplexConstructions covers unusual but valid constructions that
// combine several grammar features, including computed die sizes.
func TestRoll20Spec_ComplexConstructions(t *testing.T) {
	tests := []roll20TestCase{
		{name: "percentile dice modulus 4d%%2", seed: 0, input: "4d%%2", wantString: "4d%%2", wantRolls: []int{75, 15, 54, 7}, wantTotal: 1},
		{name: "percentile dice modulus even 4d%%2 seed1", seed: 1, input: "4d%%2", wantString: "4d%%2", wantRolls: []int{82, 88, 48, 60}, wantTotal: 0},
		{name: "percentile dice kept 2d%kh1", seed: 0, input: "2d%kh1", wantString: "2d%kh", wantRolls: []int{75}, wantTotal: 75},
		{name: "dynamic die size from floor 3d(floor(6/2))", seed: 0, input: "3d(floor(6/2))", wantString: "3d3", wantRolls: []int{1, 1, 2}, wantTotal: 4},
		{name: "dynamic die size from addition 2d(2+2)", seed: 0, input: "2d(2+2)", wantString: "2d4", wantRolls: []int{3, 3}, wantTotal: 6},
		{name: "dynamic die size from ceil 1d(ceil(5/2))", seed: 0, input: "1d(ceil(5/2))", wantString: "d3", wantRolls: []int{1}, wantTotal: 1},
		{name: "dynamic die size with modifier 1d(floor(6/2))+1", seed: 0, input: "1d(floor(6/2))+1", wantString: "d3+1", wantRolls: []int{1}, wantTotal: 2},
		{name: "dynamic die size multiplied 2d(2+2)*3", seed: 0, input: "2d(2+2)*3", wantString: "2d4*3", wantRolls: []int{3, 3}, wantTotal: 18},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_InvalidDieSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{name: "unsafe computed die size", input: "9d(floor(3/2))", err: `unsafe die type "d1"`},
		{name: "fractional computed die size", input: "9d(3/2)", err: "die size 1.5 is not a whole number"},
		{name: "non-constant die size", input: "9d(1d6)", err: "die size must be a constant expression"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileString(tt.input)
			if err == nil {
				t.Fatalf("expected compile error for %q", tt.input)
			}
			if err.Error() != tt.err {
				t.Fatalf("error mismatch: exp=%q got=%q", tt.err, err.Error())
			}
		})
	}
}
