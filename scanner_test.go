package roll

import (
	"strings"
	"testing"
)

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s   string
		tok Token
		lit string
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: tEOF},
		{s: `#`, tok: tILLEGAL, lit: `#`},
		{s: ` `, tok: tWS, lit: " "},
		{s: "\t", tok: tWS, lit: "\t"},
		{s: "\n", tok: tWS, lit: "\n"},

		// Numbers
		{s: `3`, tok: tNUM, lit: `3`},
		{s: `123`, tok: tNUM, lit: `123`},

		// Dice
		{s: `d200`, tok: tDIE, lit: "d200"},
		{s: `dF`, tok: tDIE, lit: "dF"},
		{s: `d4d5`, tok: tDIE, lit: "d4"},
		{s: `d6!!`, tok: tDIE, lit: "d6"},
		{s: `d6r`, tok: tDIE, lit: "d6"},

		// Modifiers
		{s: `+`, tok: tPLUS, lit: "+"},
		{s: `-`, tok: tMINUS, lit: "-"},

		// Extra rules
		{s: `f`, tok: tFAILURES, lit: "f"},
		{s: `!`, tok: tEXPLODE, lit: "!"},
		{s: `!!`, tok: tCOMPOUND, lit: "!!"},
		{s: `!p`, tok: tPENETRATE, lit: "!p"},
		{s: `kh`, tok: tKEEPHIGH, lit: "kh"},
		{s: `kh3`, tok: tKEEPHIGH, lit: "kh3"},
		{s: `kl`, tok: tKEEPLOW, lit: "kl"},
		{s: `kl3`, tok: tKEEPLOW, lit: "kl3"},
		{s: `kx`, tok: tILLEGAL, lit: "kx"},
		{s: `dh`, tok: tDROPHIGH, lit: "dh"},
		{s: `dh3`, tok: tDROPHIGH, lit: "dh3"},
		{s: `dl`, tok: tDROPLOW, lit: "dl"},
		{s: `dl3`, tok: tDROPLOW, lit: "dl3"},
		{s: `r`, tok: tREROLL, lit: "r"},
		{s: `ro`, tok: tREROLL, lit: "ro"},
		{s: `s`, tok: tSORT, lit: "s"},
		{s: `sd`, tok: tSORT, lit: "sd"},

		// Tests
		{s: `>`, tok: tGREATER, lit: ">"},
		{s: `<`, tok: tLESS, lit: "<"},
		{s: `=`, tok: tEQUAL, lit: "="},

		// Grouping
		{s: `{`, tok: tGROUPSTART, lit: "{"},
		{s: `}`, tok: tGROUPEND, lit: "}"},
		{s: `,`, tok: tGROUPSEP, lit: ","},
	}

	for i, tt := range tests {
		s := NewScanner(strings.NewReader(tt.s))
		tok, lit := s.Scan()
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.tok, tok, lit)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.lit, lit)
		}
	}
}
