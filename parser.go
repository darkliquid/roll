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

// ErrEndOfRoll is raised when parsing a roll has reached a terminating token
type ErrEndOfRoll string

func (e ErrEndOfRoll) Error() string {
	return fmt.Sprintf("roll parsing terminated on %q", string(e))
}

// ErrAmbiguousModifier is raised when a multiplier was misread as a modifier
type ErrAmbiguousModifier int

func (e ErrAmbiguousModifier) Error() string {
	return fmt.Sprintf("misread %+d as modifier", e)
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

// Parse parses a Roll statement.
func (p *Parser) Parse() (roll Roll, err error) {
	// First token should be a NUM or a DIE
	tok, lit := p.scanIgnoreWhitespace()

	roll, err = p.parseRoll(tok, lit, false)
	if e, ok := err.(ErrEndOfRoll); ok && e == "" {
		err = nil
	}

	return
}

// parseRoll gets a roll of any type
func (p *Parser) parseRoll(tok Token, lit string, grouped bool) (Roll, error) {
	switch tok {
	case tNUM, tDIE:
		return p.parseDiceRoll(grouped)
	case tGROUPSTART:
		return p.parseGroupedRoll(grouped)
	default:
		return nil, ErrUnexpectedToken(lit)
	}
}

// parseGrouped parses a GroupedRoll statement
func (p *Parser) parseGroupedRoll(grouped bool) (roll *GroupedRoll, err error) {
	roll = &GroupedRoll{Combined: true}

	var negative bool
	var multiplier int
	for err == nil {
		tok, lit := p.scanIgnoreWhitespace()

		var r Roll
		r, err = p.parseRoll(tok, lit, true)
		if r != nil && multiplier != 0 {
			switch t := r.(type) {
			case *GroupedRoll:
				if multiplier < 0 && t != nil {
					t.Negative = true
				}
				r = t
			case *DiceRoll:
				if t != nil {
					t.Multiplier = multiplier
				}
				r = t
			}
			multiplier = 0
		}

		if negative {
			switch t := r.(type) {
			case *GroupedRoll:
				t.Negative = true
				r = t
			case *DiceRoll:
				t.Multiplier *= -1
				r = t
			}
		}

		if err != nil {
			negative = false
			// If we got an error and we have no rolls, it's definitely broken
			if r == nil && len(roll.Rolls) == 0 {
				return
			}

			// We got an ambiguous modifier, which means the *next* roll needs
			// to use this modifier as it's multiplier, so store it for later
			if e, ok := err.(ErrAmbiguousModifier); ok {
				multiplier = int(e)
				p.unscan()
				err = nil
			} else {
				// Rollback
				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()

				// Handle separators of the group
				switch tok {
				case tPLUS:
				case tMINUS:
					negative = true
				case tGROUPSEP:
					// If we have multiple rolls and are in combined mode when we
					// get a separator, then this is an invalid grouped roll
					if len(roll.Rolls) > 1 && roll.Combined {
						return
					}

					// We aren't combining if grouping with the GROUPSEP delimiter
					roll.Combined = false
					// We've finished parsing a roll, so reset err for loop
					err = nil
				case tGROUPSTART:
					// We've finished parsing a roll, so reset err for loop and
					// unscan again to start us off on the new group
					p.unscan()
					err = nil
				case tGROUPEND:
					// We've exited the group, so leave loop by letting error fall
					// through
					err = ErrEndOfRoll(lit)
				default:
					// Otherwise it IS an error
					return
				}
			}
		}

		// If we've ended up with a dummy roll for some reason, don't add it
		if r != nil {
			roll.Rolls = append(roll.Rolls, r)
		}
	}

	// We now have the collection of rolls within the grouped roll, now we need
	// to apply the modifiers to it
	for {
		// Read a modifier
		tok, lit := p.scanIgnoreWhitespace()

		// Handle modifier or EOF
		switch tok {
		case tPLUS, tMINUS:
			var mod int
			mod, err = p.parseModifier(tok)
			if err == nil {
				roll.Modifier += mod
			} else {
				if tok == tMINUS {
					mod = -1
				}

				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()
				switch tok {
				case tEOF:
					err = ErrEndOfRoll(lit)
					return
				case tGROUPSTART:
					if grouped {
						// Technically this is an end of roll, but we want to
						// capture the multiplier to determine the sign of the
						// next term
						err = ErrAmbiguousModifier(mod)
						return
					}
					return nil, ErrUnexpectedToken(lit)
				case tGROUPEND, tGROUPSEP:
					if grouped {
						err = ErrEndOfRoll(lit)
						return
					}
					return nil, ErrUnexpectedToken(lit)
				default:
					return nil, err
				}
			}
		case tKEEPHIGH, tKEEPLOW, tDROPHIGH, tDROPLOW:
			roll.Limit, err = p.parseLimit(tok, lit)
		case tGREATER, tLESS, tEQUAL:
			p.unscan()
			roll.Success, err = p.parseComparison()
		case tFAILURES:
			roll.Failure, err = p.parseComparison()
		case tEOF:
			err = ErrEndOfRoll(lit)
			return
		case tGROUPSEP:
			if grouped {
				err = ErrEndOfRoll(lit)
				return
			}
			return nil, ErrUnexpectedToken(lit)
		case tGROUPEND:
			p.unscan()
			err = nil
			return
		default:
			return nil, ErrUnexpectedToken(lit)
		}

		// If there is an error, lets bail out
		if err != nil {
			return nil, err
		}
	}
}

// parseDiceRoll parses a DiceRoll statement
func (p *Parser) parseDiceRoll(grouped bool) (roll *DiceRoll, err error) {
	roll = &DiceRoll{}
	tok := p.buf.tok
	lit := p.buf.lit

	// If NUM, we store it as the multiplier, else we use 1
	if tok == tNUM {
		roll.Multiplier, _ = strconv.Atoi(lit)
		tok, lit = p.scanIgnoreWhitespace()
		if tok != tDIE {
			return nil, ErrUnexpectedToken(lit)
		}
	} else {
		roll.Multiplier = 1
	}

	// We will have a DIE token here, so parse it
	if roll.Die, err = p.parseDie(lit); err != nil {
		return nil, err
	}

	// Next we should loop over all our modifiers and total them up
	var mod int
	var lastTok Token
	for {
		// Read a modifier
		tok, lit := p.scanIgnoreWhitespace()

		// Handle modifier or EOF
		switch tok {
		case tPLUS, tMINUS:
			mod, err = p.parseModifier(tok)
			if err == nil {
				roll.Modifier += mod
			} else {
				if tok == tMINUS {
					mod = -1
				}

				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()
				switch tok {
				case tEOF:
					err = ErrEndOfRoll(lit)
					return
				case tGROUPSTART:
					if grouped {
						// Technically this is an end of roll, but we want to
						// capture the multiplier to determine the sign of the
						// next term
						err = ErrAmbiguousModifier(mod)
						return
					}
					return nil, ErrUnexpectedToken(lit)
				case tGROUPEND, tGROUPSEP:
					if grouped {
						err = ErrEndOfRoll(lit)
						return
					}
					return nil, ErrUnexpectedToken(lit)
				default:
					return nil, err
				}
			}
		case tEXPLODE, tCOMPOUND, tPENETRATE:
			roll.Exploding, err = p.parseExplosion(tok, lit)
		case tKEEPHIGH, tKEEPLOW, tDROPHIGH, tDROPLOW:
			roll.Limit, err = p.parseLimit(tok, lit)
		case tSORT:
			switch lit {
			case "s":
				roll.Sort = Ascending
			case "sd":
				roll.Sort = Descending
			}
		case tREROLL:
			var rr RerollOp
			rr, err = p.parseReroll(lit)
			roll.Rerolls = append(roll.Rerolls, rr)
		case tGREATER, tLESS, tEQUAL:
			p.unscan()
			roll.Success, err = p.parseComparison()
		case tFAILURES:
			roll.Failure, err = p.parseComparison()
		case tEOF:
			err = ErrEndOfRoll(lit)
			return
		case tGROUPEND, tGROUPSEP:
			if grouped {
				err = ErrEndOfRoll(lit)
				return
			}
			return nil, ErrUnexpectedToken(lit)
		case tDIE:
			// It's ambiguous whether or not a +/- number is a modifier or a
			// a combined die roll. If grouped and we get a die character AND
			// the last token processed was a modifier, then we rewind and then
			// raise a special error to indicate it needs attention.
			if grouped && (lastTok == tPLUS || lastTok == tMINUS) {
				p.unscan()
				roll.Modifier -= mod
				err = ErrAmbiguousModifier(mod)
				return
			}
			return nil, ErrUnexpectedToken(lit)
		default:
			return nil, ErrUnexpectedToken(lit)
		}

		// If there is an error, lets bail out
		if err != nil {
			return nil, err
		}

		lastTok = tok
	}
}

func (p *Parser) parseReroll(lit string) (rr RerollOp, err error) {
	if lit == "ro" {
		rr.Once = true
	}

	// determine the comparison operator for the reroll op
	compOp, err := p.parseComparison()
	if err != nil {
		return
	}

	rr.ComparisonOp = compOp
	return
}

func (p *Parser) parseModifier(tok Token) (int, error) {
	mult := 1
	if tok == tMINUS {
		mult = -1
	}
	// Get modifer value
	tok, lit := p.scanIgnoreWhitespace()
	if tok != tNUM {
		return 0, ErrUnexpectedToken(lit)
	}

	// Add to statement modifer
	mod, err := strconv.Atoi(lit)
	return mod * mult, err
}

func (p *Parser) parseDie(dieCode string) (Die, error) {
	trimmedDieCode := strings.TrimPrefix(strings.ToUpper(dieCode), "D")
	if num, err := strconv.Atoi(trimmedDieCode); err == nil {
		return NormalDie(num), nil
	}

	// Is it a Fate/Fudge die roll?
	if trimmedDieCode == "F" {
		return FateDie(0), nil
	}

	return nil, ErrUnknownDie(dieCode)
}

func (p *Parser) parseExplosion(tok Token, lit string) (*ExplodingOp, error) {
	exp := &ExplodingOp{}

	switch tok {
	case tEXPLODE:
		exp.Type = Exploding
	case tCOMPOUND:
		exp.Type = Compounded
	case tPENETRATE:
		exp.Type = Penetrating
	default:
		return nil, ErrUnexpectedToken(lit)
	}

	// determine the comparison operator for the explosion op
	compOp, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	exp.ComparisonOp = compOp

	return exp, nil
}

func (p *Parser) parseComparison() (cmp *ComparisonOp, err error) {
	tok, lit := p.scan()

	cmp = &ComparisonOp{}

	switch tok {
	case tNUM:
		cmp.Value, err = strconv.Atoi(lit)
		if err != nil {
			return
		}

		cmp.Type = Equals
		return
	case tEQUAL:
		cmp.Type = Equals
	case tGREATER:
		cmp.Type = GreaterThan
	case tLESS:
		cmp.Type = LessThan
	default:
		err = ErrUnexpectedToken(lit)
		return
	}

	tok, lit = p.scan()
	if tok != tNUM {
		err = ErrUnexpectedToken(lit)
		return
	}

	cmp.Value, err = strconv.Atoi(lit)
	if err != nil {
		return
	}

	return cmp, nil
}

func (p *Parser) parseLimit(tok Token, lit string) (lmt *LimitOp, err error) {
	lmt = &LimitOp{
		Amount: 1,
	}

	switch tok {
	case tKEEPHIGH:
		lmt.Type = KeepHighest
		lit = strings.TrimPrefix(lit, "kh")
	case tKEEPLOW:
		lmt.Type = KeepLowest
		lit = strings.TrimPrefix(lit, "kl")
	case tDROPHIGH:
		lmt.Type = DropHighest
		lit = strings.TrimPrefix(lit, "dh")
	case tDROPLOW:
		lmt.Type = DropLowest
		lit = strings.TrimPrefix(lit, "dl")
	}

	if lit != "" {
		lmt.Amount, err = strconv.Atoi(lit)
	}

	return
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
