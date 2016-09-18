package roll

import (
	"bufio"
	"bytes"
	"io"
)

var eof = rune(0)

// Return true if ch is a whitespace character
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

// Return true if ch is a number
func isNumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Return true if ch is a weird die character
func isDieChar(ch rune) bool {
	return ch == 'F' || ch == 'f'
}

// Return true if ch is a comparison character
func isCompare(ch rune) bool {
	return ch == '<' || ch == '>' || ch == '='
}

// Return true if ch is a modifier character
func isModifier(ch rune) bool {
	return ch == '+' || ch == '-'
}

// Scanner is our lexical scanner for dice roll strings
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	switch {
	case isWhitespace(ch):
		s.unread()
		return s.scanWhitespace()
	case isNumber(ch):
		s.unread()
		return s.scanNumber()
	case ch == 'd':
		s.unread()
		return s.scanDieOrDrop()
	case ch == 'f':
		return tFAILURES, string(ch)
	case ch == '!':
		s.unread()
		return s.scanExplosions()
	case ch == 'k':
		s.unread()
		return s.scanKeep()
	case ch == 'r':
		s.unread()
		return s.scanReroll()
	case ch == 's':
		s.unread()
		return s.scanSort()
	case ch == '-':
		return tMINUS, string(ch)
	case ch == '+':
		return tPLUS, string(ch)
	case ch == '>':
		return tGREATER, string(ch)
	case ch == '<':
		return tLESS, string(ch)
	case ch == '=':
		return tEQUAL, string(ch)
	case ch == '{':
		return tGROUPSTART, string(ch)
	case ch == '}':
		return tGROUPEND, string(ch)
	case ch == ',':
		return tGROUPSEP, string(ch)
	case ch == eof:
		return tEOF, ""
	}

	return tILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return tWS, buf.String()
}

// scanNumber consumes the current rune and all contiguous number runes.
func (s *Scanner) scanNumber() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent number character into the buffer.
	// Non-number characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isNumber(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier.
	return tNUM, buf.String()
}

// scanDieOrDrop consumes the current rune and all contiguous die/drop runes.
func (s *Scanner) scanDieOrDrop() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// We assume a die token by default and switch based on subsequent chars.
	tok = tDIE
	for {
		ch := s.read()

		if ch == eof {
			break
		} else if tok == tDIE && ch == 'l' {
			tok = tDROPLOW
		} else if tok == tDIE && ch == 'h' {
			tok = tDROPHIGH
		} else if tok == tDIE && !isNumber(ch) && !isDieChar(ch) {
			if !isCompare(ch) && !isModifier(ch) && ch != 'd' && ch != 'D' {
				_, _ = buf.WriteRune(ch)
			}
			s.unread()
			break
		} else if tok != tDIE && !isNumber(ch) {
			if !isCompare(ch) && !isModifier(ch) && ch != 'd' && ch != 'D' {
				_, _ = buf.WriteRune(ch)
			}
			s.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as a regular identifier.
	return tok, buf.String()
}

// scanKeep consumes the current rune and all contiguous keep runes.
func (s *Scanner) scanKeep() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// We assume an illegal token by default and switch based on later chars.
	tok = tILLEGAL
	for {
		ch := s.read()

		if ch == eof {
			break
		} else if tok == tILLEGAL && ch == 'l' {
			tok = tKEEPLOW
		} else if tok == tILLEGAL && ch == 'h' {
			tok = tKEEPHIGH
		} else if tok != tILLEGAL && !isNumber(ch) {
			s.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as a regular identifier.
	return tok, buf.String()
}

// scanExplosions consumes the current rune and all contiguous explode runes.
func (s *Scanner) scanExplosions() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// We assume an explode token by default and switch based on later chars.
	tok = tEXPLODE

	ch := s.read()
	if ch == eof {
		return tok, buf.String()
	}

	if ch == '!' {
		tok = tCOMPOUND
		_, _ = buf.WriteRune(ch)
	} else if ch == 'p' {
		tok = tPENETRATE
		_, _ = buf.WriteRune(ch)
	} else {
		s.unread()
	}

	// Otherwise return as a regular identifier.
	return tok, buf.String()
}

// scanReroll consumes the current rune and all contiguous reroll runes.
func (s *Scanner) scanReroll() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// Rerolls are simple flags with an optional modifier
	tok = tREROLL

	ch := s.read()
	if ch == eof {
		return tok, buf.String()
	}

	if ch == 'o' {
		_, _ = buf.WriteRune(ch)
	} else {
		s.unread()
	}

	// Otherwise return as a regular identifier.
	return tok, buf.String()
}

// scanSort consumes the current rune and all contiguous sort runes.
func (s *Scanner) scanSort() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// Sorts are simple flags with an optional modifier
	tok = tSORT

	ch := s.read()
	if ch == eof {
		return tok, buf.String()
	}

	if ch == 'd' {
		_, _ = buf.WriteRune(ch)
	} else {
		s.unread()
	}

	// Otherwise return as a regular identifier.
	return tok, buf.String()
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }
