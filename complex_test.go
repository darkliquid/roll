package roll

import "testing"

// TestRoll20Spec_ComplexConstructions covers unusual but valid constructions that
// combine several grammar features, including computed die sizes.
func TestRoll20Spec_ComplexConstructions(t *testing.T) {
	tests := []roll20TestCase{
		{name: "percentile dice modulus 4d%%2", seed: 0, input: "4d%%2", wantString: "4d%%2", wantRolls: []int{75, 15, 54, 7}, wantTotal: 1},
		{name: "percentile dice modulus even 4d%%2 seed1", seed: 1, input: "4d%%2", wantString: "4d%%2", wantRolls: []int{82, 88, 48, 60}, wantTotal: 0},
		{name: "percentile dice kept 2d%kh1", seed: 0, input: "2d%kh1", wantString: "2d%kh", wantRolls: []int{75}, wantTotal: 75},
		{name: "constant die size from floor 3d(floor(6/2))", seed: 0, input: "3d(floor(6/2))", wantString: "3d3", wantRolls: []int{1, 1, 2}, wantTotal: 4},
		{name: "constant die size from addition 2d(2+2)", seed: 0, input: "2d(2+2)", wantString: "2d4", wantRolls: []int{3, 3}, wantTotal: 6},
		{name: "constant die size from ceil 1d(ceil(5/2))", seed: 0, input: "1d(ceil(5/2))", wantString: "d3", wantRolls: []int{1}, wantTotal: 1},
		{name: "constant die size with modifier 1d(floor(6/2))+1", seed: 0, input: "1d(floor(6/2))+1", wantString: "d3+1", wantRolls: []int{1}, wantTotal: 2},
		{name: "constant die size multiplied 2d(2+2)*3", seed: 0, input: "2d(2+2)*3", wantString: "2d4*3", wantRolls: []int{3, 3}, wantTotal: 18},
	}
	runRoll20TestCases(t, tests)
}

// TestRoll20Spec_RuntimeDieSize covers side expressions that roll dice before the
// outer dice are rolled.
func TestRoll20Spec_RuntimeDieSize(t *testing.T) {
	tests := []roll20TestCase{
		{name: "sides from a single die 9d(1d6)", seed: 1, input: "9d(1d6)", wantString: "9d(d6)", wantRolls: []int{4, 6, 6, 2, 1, 2, 3, 5, 1}, wantTotal: 30},
		{name: "sides from a kept pool 3d(2d6kh1)", seed: 1, input: "3d(2d6kh1)", wantString: "3d(2d6kh)", wantRolls: []int{6, 6, 2}, wantTotal: 14},
		{name: "sides from addition 2d(d4+1)", seed: 0, input: "2d(d4+1)", wantString: "2d(d4+1)", wantRolls: []int{3, 2}, wantTotal: 5},
		{name: "nested computed sizes 3d(4d(d8)kh3+2)", seed: 0, input: "3d(4d(d8)kh3+2)", wantString: "3d(4d(d8)kh3+2)", wantRolls: []int{8, 8, 9}, wantTotal: 25},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_InvalidDieSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{name: "zero constant die size", input: "9d(1-1)", err: `unsafe die type "d0"`},
		{name: "negative constant die size", input: "9d(0-5)", err: `unsafe die type "d-5"`},
		{name: "fractional constant die size", input: "9d(3/2)", err: "die size 1.5 is not a whole number"},
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

func TestRoll20Spec_RuntimeDieSizeErrors(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		err   string
	}{
		{name: "zero resolved die size", seed: 0, input: "9d(1d6-1)", err: `unsafe die type "d0"`},
		{name: "fractional resolved die size", seed: 0, input: "9d(1d6/2)", err: "die size 0.5 is not a whole number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := CompileString(tt.input)
			if err != nil {
				t.Fatalf("compile %q: %v", tt.input, err)
			}
			var evalErr error
			withTestSeed(tt.seed, func() {
				_, evalErr = EvaluateProgram(program)
			})
			if evalErr == nil {
				t.Fatalf("expected evaluation error for %q", tt.input)
			}
			if evalErr.Error() != tt.err {
				t.Fatalf("error mismatch: exp=%q got=%q", tt.err, evalErr.Error())
			}
		})
	}
}
