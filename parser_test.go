package roll

import (
	"strings"
	"testing"
)

func TestParser_ParseSimpleDiceProgram(t *testing.T) {
	program, err := NewParser(strings.NewReader("3d6+4")).Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if got, want := program.String(), "3d6+4"; got != want {
		t.Fatalf("program string mismatch: got %q want %q", got, want)
	}

	if got, want := len(program.Code), 1; got != want {
		t.Fatalf("instruction count mismatch: got %d want %d", got, want)
	}
	if got, want := program.Code[0].Op, OpRollDice; got != want {
		t.Fatalf("opcode mismatch: got %v want %v", got, want)
	}
	if got, want := len(program.DiceTerms), 1; got != want {
		t.Fatalf("dice term count mismatch: got %d want %d", got, want)
	}

	term := program.DiceTerms[0]
	if got, want := term.Multiplier, 3; got != want {
		t.Fatalf("multiplier mismatch: got %d want %d", got, want)
	}
	if got, want := term.Die.String(), "d6"; got != want {
		t.Fatalf("die mismatch: got %q want %q", got, want)
	}
	if got, want := term.Modifier, 4; got != want {
		t.Fatalf("modifier mismatch: got %d want %d", got, want)
	}
}

func TestParser_ParseGroupedProgram(t *testing.T) {
	program, err := NewParser(strings.NewReader("{3d6+4,2d8}dl=1f>5")).Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if got, want := program.String(), "{3d6+4, 2d8}dlf=1f>5"; got == want {
		t.Fatal("unexpected legacy-normalized string; test guard should not pass")
	}

	if got, want := program.String(), "{3d6+4, 2d8}dlf=1f>5"; got == want {
		t.Fatal("duplicate guard")
	}

	if got, want := program.String(), "{3d6+4, 2d8}dl=1f>5"; got != want {
		t.Fatalf("program string mismatch: got %q want %q", got, want)
	}

	if got, want := len(program.Code), 3; got != want {
		t.Fatalf("instruction count mismatch: got %d want %d", got, want)
	}
	if program.Code[0].Op != OpRollDice || program.Code[1].Op != OpRollDice || program.Code[2].Op != OpRollGroup {
		t.Fatalf("unexpected opcode sequence: %#v", program.Code)
	}
	if got, want := len(program.GroupTerms), 1; got != want {
		t.Fatalf("group term count mismatch: got %d want %d", got, want)
	}

	group := program.GroupTerms[0]
	if group.Combined {
		t.Fatal("expected separated group")
	}
	if got, want := group.ChildCount, 2; got != want {
		t.Fatalf("child count mismatch: got %d want %d", got, want)
	}
	if group.Limit == nil || group.Limit.Type != DropLowest || group.Limit.Amount != 1 {
		t.Fatalf("unexpected limit: %#v", group.Limit)
	}
	if group.Success == nil || group.Success.Type != Equals || group.Success.Value != 1 {
		t.Fatalf("unexpected success comparison: %#v", group.Success)
	}
	if group.Failure == nil || group.Failure.Type != GreaterThan || group.Failure.Value != 5 {
		t.Fatalf("unexpected failure comparison: %#v", group.Failure)
	}
}

func TestParser_ParseAscendingAliasAndInclusiveComparison(t *testing.T) {
	program, err := NewParser(strings.NewReader("6d6sa>=5")).Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if got, want := program.String(), "6d6>=5s"; got != want {
		t.Fatalf("program string mismatch: got %q want %q", got, want)
	}

	term := program.DiceTerms[0]
	if term.Sort != Ascending {
		t.Fatalf("expected ascending sort, got %v", term.Sort)
	}
	if term.Success == nil {
		t.Fatal("expected success comparison")
	}
	if term.Success.Type != GreaterThan || term.Success.Value != 5 || !term.Success.Inclusive {
		t.Fatalf("unexpected success comparison: %#v", term.Success)
	}
}

func TestParser_ParseRejectsUnsafeDie(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		limits Limits
		err    string
	}{
		{
			name:   "reject d1",
			input:  "d1",
			limits: DefaultLimits,
			err:    `unsafe die type "d1"`,
		},
		{
			name:   "reject oversized die",
			input:  "d1001",
			limits: Limits{MaxDieSize: 1000},
			err:    "die size 1001 exceeds maximum 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParserWithLimits(strings.NewReader(tt.input), tt.limits).Parse()
			if err == nil {
				t.Fatal("expected parse error")
			}
			if err.Error() != tt.err {
				t.Fatalf("unexpected parse error: exp=%q got=%q", tt.err, err.Error())
			}
		})
	}
}

func TestParser_ParseErrors(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		{input: "foo", err: `found unexpected token "f"`},
		{input: "dX", err: `unrecognised die type "dX"`},
		{input: "d4--", err: `found unexpected token "-"`},
		{input: "3d4d5", err: `found unexpected token "d5"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := NewParser(strings.NewReader(tt.input)).Parse()
			if err == nil {
				t.Fatal("expected parse error")
			}
			if err.Error() != tt.err {
				t.Fatalf("unexpected parse error: exp=%q got=%q", tt.err, err.Error())
			}
		})
	}
}
