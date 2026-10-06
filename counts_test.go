package roll

import "testing"

func TestResultCounts_Rerolls(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		want  int
	}{
		{name: "no reroll modifier", seed: 2, input: "4d6", want: 0},
		{name: "reroll once fires twice", seed: 2, input: "4d6ro1", want: 2},
		{name: "reroll once no match", seed: 2, input: "4d6ro6", want: 0},
		{name: "recursive reroll counts every reroll", seed: 52, input: "4d6r1", want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			if result.Rerolls != tt.want {
				t.Fatalf("rerolls: got %d want %d", result.Rerolls, tt.want)
			}
		})
	}
}
