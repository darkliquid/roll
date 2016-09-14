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
	case ch == '-':
		return tMINUS, string(ch)
	case ch == '+':
		return tPLUS, string(ch)
	case ch == eof:
		return tEOF, ""
	case ch == 'd' || ch == 'D':
		return s.scanDie()
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

// scanDie consumes the current rune and all contiguous die runes.
func (s *Scanner) scanDie() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent die character into the buffer.
	// Non-die characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isNumber(ch) && !isDieChar(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier.
	return tDIE, buf.String()
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
