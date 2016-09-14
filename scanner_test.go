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
		{s: `d200`, tok: tDIE, lit: "200"},
		{s: `dF`, tok: tDIE, lit: "F"},
		{s: `d4d5`, tok: tDIE, lit: "4"},

		// Modifiers
		{s: `+`, tok: tPLUS, lit: "+"},
		{s: `-`, tok: tMINUS, lit: "-"},
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
