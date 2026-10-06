package roll

import "testing"

func TestResultCounts_Aggregation(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		input      string
		rerolls    int
		explosions int
		drops      int
	}{
		{name: "separate terms sum explosions", seed: 6, input: "2d6!5+2d6!5", explosions: 2},
		{name: "separated group sums child explosions", seed: 2, input: "{2d6!5, 1d8}", explosions: 1},
		{name: "combined group sums child explosions", seed: 2, input: "{2d6!5 + 1d8}", explosions: 1},
		{name: "nested groups sum events", seed: 2, input: "{{2d6!5}, 1d8}", explosions: 1},
		{name: "group limit counts drops", seed: 0, input: "{3d6, 2d8}dl1", drops: 1},
		{name: "rerolls and drops across a term", seed: 2, input: "4d6ro1kh3", rerolls: 2, drops: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			if result.Rerolls != tt.rerolls {
				t.Fatalf("rerolls: got %d want %d", result.Rerolls, tt.rerolls)
			}
			if result.Explosions != tt.explosions {
				t.Fatalf("explosions: got %d want %d", result.Explosions, tt.explosions)
			}
			if result.Drops != tt.drops {
				t.Fatalf("drops: got %d want %d", result.Drops, tt.drops)
			}
		})
	}
}

func TestResultCounts_Drops(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		want  int
	}{
		{name: "no limit modifier", seed: 2, input: "4d6", want: 0},
		{name: "keep highest drops one", seed: 2, input: "4d6kh3", want: 1},
		{name: "keep highest drops three", seed: 2, input: "4d6kh1", want: 3},
		{name: "keep lowest drops two", seed: 2, input: "4d6kl2", want: 2},
		{name: "drop highest drops one", seed: 2, input: "4d6dh1", want: 1},
		{name: "drop lowest drops one", seed: 2, input: "4d6dl1", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			if result.Drops != tt.want {
				t.Fatalf("drops: got %d want %d", result.Drops, tt.want)
			}
		})
	}
}

func TestResultCounts_Explosions(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		input string
		want  int
	}{
		{name: "no explosion modifier", seed: 2, input: "2d6", want: 0},
		{name: "exploding adds a die", seed: 2, input: "2d6!5", want: 1},
		{name: "penetrating elided zero is not counted", seed: 2, input: "2d6!p5", want: 0},
		{name: "penetrating kept die is counted", seed: 9, input: "2d6!p6", want: 1},
		{name: "compounded counts contributing explosions", seed: 2, input: "2d6!!5", want: 1},
		{name: "compounded counts each contributing explosion", seed: 1, input: "2d6!!6", want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateProgram(t, tt.seed, tt.input)
			if result.Explosions != tt.want {
				t.Fatalf("explosions: got %d want %d", result.Explosions, tt.want)
			}
		})
	}
}

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
