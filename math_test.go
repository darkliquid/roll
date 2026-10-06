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
