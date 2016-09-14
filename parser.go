package roll

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrUnexpectedToken is raised on unexpected tokens
type ErrUnexpectedToken string

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf("found unexpected token %q", string(e))
}

// ErrUnknownDie is raised on unrecognised die types
type ErrUnknownDie string

func (e ErrUnknownDie) Error() string {
	return fmt.Sprintf("unrecognised die type %q", string(e))
}

// Parser is our dice rolling parser
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token
		lit string
		n   int
	}
}

// NewParser returns a Parser instance
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a SQL SELECT statement.
func (p *Parser) Parse() (stmt *DiceRollStmt, err error) {
	stmt = &DiceRollStmt{}

	// First token should be a NUM or a DIE
	tok, lit := p.scanIgnoreWhitespace()
	if tok != tNUM && tok != tDIE {
		return nil, ErrUnexpectedToken(lit)
	}

	// If NUM, we store it as the multiplier, else we use 1
	if tok == tNUM {
		stmt.Multiplier, _ = strconv.Atoi(lit)
		tok, lit = p.scanIgnoreWhitespace()
		if tok != tDIE {
			return nil, ErrUnexpectedToken(lit)
		}
	} else {
		stmt.Multiplier = 1
	}

	// We will have a DIE token here, so parse it
	if stmt.Die, err = p.parseDie(lit); err != nil {
		return nil, err
	}

	// Next we should loop over all our modifiers and total them up
	for {
		// Read a modifier
		tok, lit := p.scanIgnoreWhitespace()

		// Handle modifier or EOF
		mult := 1
		switch tok {
		case tPLUS:
		case tMINUS:
			mult = -1
		case tEOF:
			return
		default:
			return nil, ErrUnexpectedToken(lit)
		}

		// Get modifer value
		tok, lit = p.scanIgnoreWhitespace()
		if tok != tNUM {
			return nil, ErrUnexpectedToken(lit)
		}

		// Add to statement modifer
		mod, _ := strconv.Atoi(lit)
		stmt.Modifier += mod * mult
	}
}

func (p *Parser) parseDie(dieCode string) (Die, error) {
	if num, err := strconv.Atoi(dieCode); err == nil {
		return NormalDie(num), nil
	}

	// Is it a Fate/Fudge die roll?
	if strings.ToUpper(dieCode) == "F" {
		return FateDie(0), nil
	}

	return nil, ErrUnknownDie(dieCode)
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == tWS {
		tok, lit = p.scan()
	}
	return
}
