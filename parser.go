package roll

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrUnexpectedToken is raised on unexpected tokens.
type ErrUnexpectedToken string

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf("found unexpected token %q", string(e))
}

// ErrUnknownDie is raised on unrecognised die types.
type ErrUnknownDie string

func (e ErrUnknownDie) Error() string {
	return fmt.Sprintf("unrecognised die type %q", string(e))
}

// ErrEndOfRoll is raised when parsing a roll has reached a terminating token.
type ErrEndOfRoll string

func (e ErrEndOfRoll) Error() string {
	return fmt.Sprintf("roll parsing terminated on %q", string(e))
}

// ErrAmbiguousModifier is raised when a multiplier was misread as a modifier.
type ErrAmbiguousModifier int

func (e ErrAmbiguousModifier) Error() string {
	return fmt.Sprintf("misread %+d as modifier", e)
}

// Parser compiles dice notation into VM bytecode.
type Parser struct {
	s      *Scanner
	limits Limits
	buf    struct {
		tok Token
		lit string
		n   int
	}
}

// NewParser returns a compiler instance.
func NewParser(r io.Reader) *Parser {
	return NewParserWithLimits(r, DefaultLimits)
}

// NewParserWithLimits returns a compiler instance using explicit safety limits.
func NewParserWithLimits(r io.Reader, limits Limits) *Parser {
	return &Parser{s: NewScanner(r), limits: limits.normalized()}
}

// Parse compiles a roll expression into VM bytecode.
func (p *Parser) Parse() (program *Program, err error) {
	tok, lit := p.scanIgnoreWhitespace()

	var root compiledNode
	root, err = p.parseRoll(tok, lit, false)
	if e, ok := err.(ErrEndOfRoll); ok && e == "" {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	program = &Program{
		Rendered: strings.TrimPrefix(root.render(), "+"),
		MaxDepth: root.maxDepth(),
	}
	root.emit(program)
	return program, nil
}

type compiledNode interface {
	emit(*Program)
	render() string
	maxDepth() int
}

type diceNode struct {
	term DiceTerm
}

func (n *diceNode) emit(program *Program) {
	idx := len(program.DiceTerms)
	program.DiceTerms = append(program.DiceTerms, n.term)
	program.Code = append(program.Code, Instruction{Op: OpRollDice, Arg: idx})
}

func (n *diceNode) render() string {
	return renderDiceTerm(n.term)
}

func (n *diceNode) maxDepth() int {
	return 1
}

type groupNode struct {
	term     GroupTerm
	children []compiledNode
}

func (n *groupNode) emit(program *Program) {
	for _, child := range n.children {
		child.emit(program)
	}
	idx := len(program.GroupTerms)
	term := n.term
	term.ChildCount = len(n.children)
	program.GroupTerms = append(program.GroupTerms, term)
	program.Code = append(program.Code, Instruction{Op: OpRollGroup, Arg: idx})
}

func (n *groupNode) render() string {
	parts := make([]string, 0, len(n.children))
	for _, child := range n.children {
		if child != nil {
			parts = append(parts, child.render())
		}
	}

	sep := ", "
	if n.term.Combined {
		sep = " + "
	}

	output := strings.Join(parts, sep)
	if n.term.Combined {
		output = strings.ReplaceAll(output, "+-", "-")
	} else if len(n.children) == 1 {
		output += ","
	}

	output = "{" + output + "}"
	output = strings.ReplaceAll(output, "{+", "{")
	output = strings.ReplaceAll(output, "{-", "{")
	output = strings.ReplaceAll(output, ", +", ", ")
	output = strings.ReplaceAll(output, ", -", ", ")
	output = strings.ReplaceAll(output, "+ +", "+ ")
	output = strings.ReplaceAll(output, "+ -", "- ")

	if n.term.Limit != nil {
		output += n.term.Limit.String()
	}
	if n.term.Success != nil {
		output += n.term.Success.String()
	}
	if n.term.Failure != nil {
		output += "f" + n.term.Failure.String()
	}
	if n.term.Modifier != 0 {
		output += fmt.Sprintf("%+d", n.term.Modifier)
	}
	if n.term.Negative {
		output = "-" + output
	}

	return output
}

func (n *groupNode) maxDepth() int {
	depth := 1
	for _, child := range n.children {
		depth = max(depth, 1+child.maxDepth())
	}
	return depth
}

func renderDiceTerm(term DiceTerm) string {
	var output strings.Builder
	if term.Multiplier > 1 || term.Multiplier < -1 {
		output.WriteString(fmt.Sprintf("%+d", term.Multiplier))
	} else if term.Multiplier == -1 {
		output.WriteString("-")
	} else if term.Multiplier == 1 {
		output.WriteString("+")
	}

	output.WriteString(term.Die.String())

	if term.Modifier != 0 {
		output.WriteString(fmt.Sprintf("%+d", term.Modifier))
	}
	for _, reroll := range term.Rerolls {
		output.WriteString(reroll.String())
	}
	if term.Exploding != nil {
		output.WriteString(term.Exploding.String())
	}
	if term.Limit != nil {
		output.WriteString(term.Limit.String())
	}
	if term.Success != nil {
		output.WriteString(term.Success.String())
	}
	if term.Failure != nil {
		output.WriteString("f" + term.Failure.String())
	}
	output.WriteString(term.Sort.String())

	return output.String()
}

func (p *Parser) parseRoll(tok Token, lit string, grouped bool) (compiledNode, error) {
	switch tok {
	case tNUM, tDIE:
		return p.parseDiceRoll(grouped)
	case tGROUPSTART:
		return p.parseGroupedRoll(grouped)
	default:
		return nil, ErrUnexpectedToken(lit)
	}
}

func (p *Parser) parseGroupedRoll(grouped bool) (compiledNode, error) {
	node := &groupNode{term: GroupTerm{Combined: true}}

	var negative bool
	var multiplier int
	var err error
	for err == nil {
		tok, lit := p.scanIgnoreWhitespace()

		child, childErr := p.parseRoll(tok, lit, true)
		if child != nil && multiplier != 0 {
			switch n := child.(type) {
			case *groupNode:
				if multiplier < 0 {
					n.term.Negative = true
				}
			case *diceNode:
				n.term.Multiplier = multiplier
			}
			multiplier = 0
		}

		if negative {
			switch n := child.(type) {
			case *groupNode:
				n.term.Negative = true
			case *diceNode:
				n.term.Multiplier *= -1
			}
		}

		err = childErr
		if err != nil {
			negative = false
			if child == nil && len(node.children) == 0 {
				return nil, err
			}

			if e, ok := err.(ErrAmbiguousModifier); ok {
				multiplier = int(e)
				p.unscan()
				err = nil
			} else {
				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()
				switch tok {
				case tPLUS:
				case tMINUS:
					negative = true
				case tGROUPSEP:
					if len(node.children) > 1 && node.term.Combined {
						return nil, err
					}
					node.term.Combined = false
					err = nil
				case tGROUPSTART:
					p.unscan()
					err = nil
				case tGROUPEND:
					err = ErrEndOfRoll(lit)
				default:
					return nil, err
				}
			}
		}

		if child != nil {
			node.children = append(node.children, child)
		}
	}

	for {
		tok, lit := p.scanIgnoreWhitespace()
		switch tok {
		case tPLUS, tMINUS:
			var mod int
			mod, err = p.parseModifier(tok)
			if err == nil {
				node.term.Modifier += mod
			} else {
				if tok == tMINUS {
					mod = -1
				}

				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()
				switch tok {
				case tEOF:
					return node, ErrEndOfRoll(lit)
				case tGROUPSTART:
					if grouped {
						return node, ErrAmbiguousModifier(mod)
					}
					return nil, ErrUnexpectedToken(lit)
				case tGROUPEND, tGROUPSEP:
					if grouped {
						return node, ErrEndOfRoll(lit)
					}
					return nil, ErrUnexpectedToken(lit)
				default:
					return nil, err
				}
			}
		case tKEEPHIGH, tKEEPLOW, tDROPHIGH, tDROPLOW:
			node.term.Limit, err = p.parseLimit(tok, lit)
		case tGREATER, tLESS, tEQUAL:
			p.unscan()
			node.term.Success, err = p.parseComparison()
		case tFAILURES:
			node.term.Failure, err = p.parseComparison()
		case tEOF:
			return node, ErrEndOfRoll(lit)
		case tGROUPSEP:
			if grouped {
				return node, ErrEndOfRoll(lit)
			}
			return nil, ErrUnexpectedToken(lit)
		case tGROUPEND:
			p.unscan()
			return node, nil
		default:
			return nil, ErrUnexpectedToken(lit)
		}

		if err != nil {
			return nil, err
		}
	}
}

func (p *Parser) parseDiceRoll(grouped bool) (compiledNode, error) {
	node := &diceNode{term: DiceTerm{Multiplier: 1}}
	tok := p.buf.tok
	lit := p.buf.lit

	if tok == tNUM {
		node.term.Multiplier, _ = strconv.Atoi(lit)
		tok, lit = p.scanIgnoreWhitespace()
		if tok != tDIE {
			return nil, ErrUnexpectedToken(lit)
		}
	}

	die, err := p.parseDie(lit)
	if err != nil {
		return nil, err
	}
	node.term.Die = die

	var mod int
	var lastTok Token
	for {
		tok, lit = p.scanIgnoreWhitespace()
		switch tok {
		case tPLUS, tMINUS:
			mod, err = p.parseModifier(tok)
			if err == nil {
				node.term.Modifier += mod
			} else {
				if tok == tMINUS {
					mod = -1
				}

				p.unscan()
				tok, lit = p.scanIgnoreWhitespace()
				switch tok {
				case tEOF:
					return node, ErrEndOfRoll(lit)
				case tGROUPSTART:
					if grouped {
						return node, ErrAmbiguousModifier(mod)
					}
					return nil, ErrUnexpectedToken(lit)
				case tGROUPEND, tGROUPSEP:
					if grouped {
						return node, ErrEndOfRoll(lit)
					}
					return nil, ErrUnexpectedToken(lit)
				default:
					return nil, err
				}
			}
		case tEXPLODE, tCOMPOUND, tPENETRATE:
			node.term.Exploding, err = p.parseExplosion(tok, lit)
		case tKEEPHIGH, tKEEPLOW, tDROPHIGH, tDROPLOW:
			node.term.Limit, err = p.parseLimit(tok, lit)
		case tSORT:
			switch lit {
			case "s", "sa":
				node.term.Sort = Ascending
			case "sd":
				node.term.Sort = Descending
			}
		case tREROLL:
			var reroll RerollOp
			reroll, err = p.parseReroll(lit)
			node.term.Rerolls = append(node.term.Rerolls, reroll)
		case tGREATER, tLESS, tEQUAL:
			p.unscan()
			node.term.Success, err = p.parseComparison()
		case tFAILURES:
			node.term.Failure, err = p.parseComparison()
		case tEOF:
			return node, ErrEndOfRoll(lit)
		case tGROUPEND, tGROUPSEP:
			if grouped {
				return node, ErrEndOfRoll(lit)
			}
			return nil, ErrUnexpectedToken(lit)
		case tDIE:
			if grouped && (lastTok == tPLUS || lastTok == tMINUS) {
				p.unscan()
				node.term.Modifier -= mod
				return node, ErrAmbiguousModifier(mod)
			}
			return nil, ErrUnexpectedToken(lit)
		default:
			return nil, ErrUnexpectedToken(lit)
		}

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

	compOp, err := p.parseComparison()
	if err != nil {
		return rr, err
	}

	rr.ComparisonOp = compOp
	return rr, nil
}

func (p *Parser) parseModifier(tok Token) (int, error) {
	mult := 1
	if tok == tMINUS {
		mult = -1
	}
	tok, lit := p.scanIgnoreWhitespace()
	if tok != tNUM {
		return 0, ErrUnexpectedToken(lit)
	}

	mod, err := strconv.Atoi(lit)
	return mod * mult, err
}

func (p *Parser) parseDie(dieCode string) (Die, error) {
	trimmedDieCode := strings.TrimPrefix(strings.ToUpper(dieCode), "D")
	if num, err := strconv.Atoi(trimmedDieCode); err == nil {
		die := NormalDie(num)
		if err := validateDieLimits(die, p.limits); err != nil {
			return nil, err
		}
		return die, nil
	}

	if trimmedDieCode == "F" {
		die := FateDie(0)
		if err := validateDieLimits(die, p.limits); err != nil {
			return nil, err
		}
		return die, nil
	}

	if trimmedDieCode == "%" {
		die := PercentileDie(0)
		if err := validateDieLimits(die, p.limits); err != nil {
			return nil, err
		}
		return die, nil
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
			return nil, err
		}
		cmp.Type = Equals
		return cmp, nil
	case tEQUAL:
		cmp.Type = Equals
	case tGREATER:
		cmp.Type = GreaterThan
	case tLESS:
		cmp.Type = LessThan
	default:
		return nil, ErrUnexpectedToken(lit)
	}

	tok, lit = p.scan()
	if tok == tEQUAL && (cmp.Type == GreaterThan || cmp.Type == LessThan) {
		cmp.Inclusive = true
		tok, lit = p.scan()
	}
	if tok != tNUM {
		return nil, ErrUnexpectedToken(lit)
	}

	cmp.Value, err = strconv.Atoi(lit)
	if err != nil {
		return nil, err
	}

	return cmp, nil
}

func (p *Parser) parseLimit(tok Token, lit string) (lmt *LimitOp, err error) {
	lmt = &LimitOp{Amount: 1}

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

	return lmt, err
}

func (p *Parser) scan() (tok Token, lit string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit = p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit

	return tok, lit
}

func (p *Parser) unscan() { p.buf.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == tWS {
		tok, lit = p.scan()
	}
	return tok, lit
}
