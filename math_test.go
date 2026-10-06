package roll

import "testing"

func TestRoll20Spec_Multiplication(t *testing.T) {
	tests := []roll20TestCase{
		{name: "dice times constant 2d6*3", seed: 2, input: "2d6*3", wantString: "2d6*3", wantRolls: []int{5, 1}, wantTotal: 18},
		{name: "constant times constant 5*3", seed: 0, input: "5*3", wantString: "5*3", wantRolls: []int{}, wantTotal: 15},
		{name: "parenthesised roll (2d6+3)*4", seed: 2, input: "(2d6+3)*4", wantString: "(2d6+3)*4", wantRolls: []int{5, 1}, wantTotal: 36},
		{name: "multiplication before addition 2d6*3+4", seed: 2, input: "2d6*3+4", wantString: "2d6*3+4", wantRolls: []int{5, 1}, wantTotal: 22},
		{name: "product added to roll 2d6+4*3", seed: 2, input: "2d6+4*3", wantString: "2d6+4*3", wantRolls: []int{5, 1}, wantTotal: 18},
		{name: "chained multiplication 2d6*2*3", seed: 2, input: "2d6*2*3", wantString: "2d6*2*3", wantRolls: []int{5, 1}, wantTotal: 36},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Division(t *testing.T) {
	tests := []roll20TestCase{
		{name: "dice divided by constant 2d6/2", seed: 2, input: "2d6/2", wantString: "2d6/2", wantRolls: []int{5, 1}, wantTotal: 3},
		{name: "constant divided by constant 7/2", seed: 0, input: "7/2", wantString: "7/2", wantRolls: []int{}, wantTotal: 3},
		{name: "exact division 10/2", seed: 0, input: "10/2", wantString: "10/2", wantRolls: []int{}, wantTotal: 5},
		{name: "division before addition 2d6/2+1", seed: 2, input: "2d6/2+1", wantString: "2d6/2+1", wantRolls: []int{5, 1}, wantTotal: 4},
		{name: "chained division and multiplication 2d6/2*3", seed: 2, input: "2d6/2*3", wantString: "2d6/2*3", wantRolls: []int{5, 1}, wantTotal: 9},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_DiceMatching(t *testing.T) {
	tests := []roll20TestCase{
		{name: "no matches 2d6mt", seed: 2, input: "2d6mt", wantString: "2d6mt", wantRolls: []int{5, 1}, wantTotal: 0},
		{name: "visual match keeps total 2d6m", seed: 2, input: "2d6m", wantString: "2d6m", wantRolls: []int{5, 1}, wantTotal: 6},
		{name: "three of a kind 4d6mt3", seed: 1, input: "4d6mt3", wantString: "4d6mt3", wantRolls: []int{6, 4, 6, 6}, wantTotal: 1},
		{name: "default two of a kind 4d6mt", seed: 1, input: "4d6mt", wantString: "4d6mt", wantRolls: []int{6, 4, 6, 6}, wantTotal: 1},
		{name: "visual three of a kind 4d6m3", seed: 1, input: "4d6m3", wantString: "4d6m3", wantRolls: []int{6, 4, 6, 6}, wantTotal: 22},
		{name: "matching value filter 4d6mt3>5", seed: 1, input: "4d6mt3>5", wantString: "4d6mt3>5", wantRolls: []int{6, 4, 6, 6}, wantTotal: 1},
		{name: "matching value filter excludes 4d6mt3>6", seed: 1, input: "4d6mt3>6", wantString: "4d6mt3>6", wantRolls: []int{6, 4, 6, 6}, wantTotal: 0},
		{name: "insufficient matches 4d6mt4", seed: 1, input: "4d6mt4", wantString: "4d6mt4", wantRolls: []int{6, 4, 6, 6}, wantTotal: 0},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Abs(t *testing.T) {
	tests := []roll20TestCase{
		{name: "abs of negative constant abs(-3)", seed: 0, input: "abs(-3)", wantString: "abs(-3)", wantRolls: []int{}, wantTotal: 3},
		{name: "abs of positive constant abs(3)", seed: 0, input: "abs(3)", wantString: "abs(3)", wantRolls: []int{}, wantTotal: 3},
		{name: "abs of subtraction abs(3-10)", seed: 0, input: "abs(3-10)", wantString: "abs(3-10)", wantRolls: []int{}, wantTotal: 7},
		{name: "abs of negative dice abs(-2d6)", seed: 2, input: "abs(-2d6)", wantString: "abs(-2d6)", wantRolls: []int{5, 1}, wantTotal: 6},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Ceil(t *testing.T) {
	tests := []roll20TestCase{
		{name: "ceil rounds up ceil(11/2)", seed: 0, input: "ceil(11/2)", wantString: "ceil(11/2)", wantRolls: []int{}, wantTotal: 6},
		{name: "ceil of exact ceil(4/2)", seed: 0, input: "ceil(4/2)", wantString: "ceil(4/2)", wantRolls: []int{}, wantTotal: 2},
		{name: "ceil towards positive infinity ceil(-3/2)", seed: 0, input: "ceil(-3/2)", wantString: "ceil(-3/2)", wantRolls: []int{}, wantTotal: -1},
		{name: "ceil of dice division ceil(2d6/4)", seed: 2, input: "ceil(2d6/4)", wantString: "ceil(2d6/4)", wantRolls: []int{5, 1}, wantTotal: 2},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Round(t *testing.T) {
	tests := []roll20TestCase{
		{name: "round half up round(11/2)", seed: 0, input: "round(11/2)", wantString: "round(11/2)", wantRolls: []int{}, wantTotal: 6},
		{name: "round half up round(5/2)", seed: 0, input: "round(5/2)", wantString: "round(5/2)", wantRolls: []int{}, wantTotal: 3},
		{name: "round toward zero round(1/3)", seed: 0, input: "round(1/3)", wantString: "round(1/3)", wantRolls: []int{}, wantTotal: 0},
		{name: "round toward positive infinity round(-5/2)", seed: 0, input: "round(-5/2)", wantString: "round(-5/2)", wantRolls: []int{}, wantTotal: -2},
		{name: "round of dice division round(2d6/4)", seed: 2, input: "round(2d6/4)", wantString: "round(2d6/4)", wantRolls: []int{5, 1}, wantTotal: 2},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Floor(t *testing.T) {
	tests := []roll20TestCase{
		{name: "floor of division floor(11/2)", seed: 0, input: "floor(11/2)", wantString: "floor(11/2)", wantRolls: []int{}, wantTotal: 5},
		{name: "floor towards negative infinity floor(-3/2)", seed: 0, input: "floor(-3/2)", wantString: "floor(-3/2)", wantRolls: []int{}, wantTotal: -2},
		{name: "floor of dice division floor(2d6/2)", seed: 2, input: "floor(2d6/2)", wantString: "floor(2d6/2)", wantRolls: []int{5, 1}, wantTotal: 3},
		{name: "floor added to dice floor(11/2)+1d6", seed: 0, input: "floor(11/2)+1d6", wantString: "floor(11/2)+d6", wantRolls: []int{1}, wantTotal: 6},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Exponentiation(t *testing.T) {
	tests := []roll20TestCase{
		{name: "constant power 2**3", seed: 0, input: "2**3", wantString: "2**3", wantRolls: []int{}, wantTotal: 8},
		{name: "right associative 2**3**2", seed: 0, input: "2**3**2", wantString: "2**3**2", wantRolls: []int{}, wantTotal: 512},
		{name: "dice power 3d6**2", seed: 2, input: "3d6**2", wantString: "3d6**2", wantRolls: []int{5, 1, 1}, wantTotal: 49},
		{name: "power before multiplication 2*3**2", seed: 0, input: "2*3**2", wantString: "2*3**2", wantRolls: []int{}, wantTotal: 18},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_Modulus(t *testing.T) {
	tests := []roll20TestCase{
		{name: "dice modulus constant 2d6%3", seed: 2, input: "2d6%3", wantString: "2d6%3", wantRolls: []int{5, 1}, wantTotal: 0},
		{name: "constant modulus constant 7%3", seed: 0, input: "7%3", wantString: "7%3", wantRolls: []int{}, wantTotal: 1},
		{name: "constant modulus constant 10%4", seed: 0, input: "10%4", wantString: "10%4", wantRolls: []int{}, wantTotal: 2},
		{name: "modulus before addition 2d6%4+1", seed: 2, input: "2d6%4+1", wantString: "2d6%4+1", wantRolls: []int{5, 1}, wantTotal: 3},
		{name: "percentile die still works d%", seed: 0, input: "d%", wantString: "d%", wantRolls: []int{75}, wantTotal: 75},
	}
	runRoll20TestCases(t, tests)
}

func TestRoll20Spec_ModulusByZero(t *testing.T) {
	program, err := CompileString("5%0")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err := EvaluateProgram(program); err == nil {
		t.Fatal("expected modulus by zero error")
	} else if err.Error() != "modulus by zero" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoll20Spec_DivisionByZero(t *testing.T) {
	program, err := CompileString("5/0")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err := EvaluateProgram(program); err == nil {
		t.Fatal("expected division by zero error")
	} else if err.Error() != "division by zero" {
		t.Fatalf("unexpected error: %v", err)
	}
}
