package roll

import (
	"fmt"
	"sort"
	"strings"
)

// Canonical returns the deterministic, normalized dice notation for the compiled program.
func (p *Program) Canonical() string {
	if p == nil || len(p.Code) == 0 {
		return ""
	}

	// Build AST string representation canonically from AST nodes
	// We can reconstruct the root node string representation using canonical rules.
	return canonicalizeProgram(p)
}

// Canonicalize parses a dice roll string and returns its canonical representation.
func Canonicalize(input string) (string, error) {
	program, err := CompileString(input)
	if err != nil {
		return "", err
	}
	return program.Canonical(), nil
}

type stackNode struct {
	rendered string
	sortKey  string
}

func canonicalizeProgram(p *Program) string {
	stack := make([]stackNode, 0, len(p.Code))

	for _, inst := range p.Code {
		switch inst.Op {
		case OpRollDice:
			if inst.Arg < 0 || inst.Arg >= len(p.DiceTerms) {
				return p.Rendered
			}
			term := p.DiceTerms[inst.Arg]
			rendered := canonicalizeDiceTerm(term)
			sortKey := diceSortKey(term, rendered)
			stack = append(stack, stackNode{rendered: rendered, sortKey: sortKey})

		case OpRollGroup:
			if inst.Arg < 0 || inst.Arg >= len(p.GroupTerms) {
				return p.Rendered
			}
			term := p.GroupTerms[inst.Arg]
			if term.ChildCount > len(stack) {
				return p.Rendered
			}

			children := stack[len(stack)-term.ChildCount:]
			stack = stack[:len(stack)-term.ChildCount]

			rendered := canonicalizeGroupTerm(term, children)
			sortKey := strings.TrimPrefix(strings.TrimPrefix(rendered, "-"), "+")
			stack = append(stack, stackNode{rendered: rendered, sortKey: sortKey})
		}
	}

	if len(stack) != 1 {
		return p.Rendered
	}

	return strings.TrimPrefix(stack[0].rendered, "+")
}

func diceSortKey(term DiceTerm, rendered string) string {
	cleanRendered := strings.TrimPrefix(strings.TrimPrefix(rendered, "+"), "-")
	switch d := term.Die.(type) {
	case NormalDie:
		return fmt.Sprintf("d%09d:%s", d, cleanRendered)
	case PercentileDie:
		return fmt.Sprintf("d%09d:%s", 100, cleanRendered)
	case FateDie:
		return fmt.Sprintf("dF:%s", cleanRendered)
	default:
		return fmt.Sprintf("%s:%s", term.Die.String(), cleanRendered)
	}
}

func canonicalizeDiceTerm(term DiceTerm) string {
	var output strings.Builder

	// Multiplier & Die
	if term.Multiplier > 1 || term.Multiplier < -1 {
		fmt.Fprintf(&output, "%+d", term.Multiplier)
	} else if term.Multiplier == -1 {
		output.WriteString("-")
	} else if term.Multiplier == 1 {
		output.WriteString("+")
	}

	// Die representation (canonical lowercase/uppercase symbols, omit count 1)
	output.WriteString(canonicalDieString(term.Die))

	// Term modifier
	if term.Modifier != 0 {
		fmt.Fprintf(&output, "%+d", term.Modifier)
	}

	// Rerolls (sorted lexicographically by string)
	if len(term.Rerolls) > 0 {
		rrStrs := make([]string, len(term.Rerolls))
		for i, rr := range term.Rerolls {
			rrStrs[i] = rr.String()
		}
		sort.Strings(rrStrs)
		for _, s := range rrStrs {
			output.WriteString(s)
		}
	}

	// Exploding
	if term.Exploding != nil {
		output.WriteString(term.Exploding.String())
	}

	// Limit (kh, kl, dh, dl)
	if term.Limit != nil {
		output.WriteString(term.Limit.String())
	}

	// Success
	if term.Success != nil {
		output.WriteString(term.Success.String())
	}

	// Failure
	if term.Failure != nil {
		output.WriteString("f" + term.Failure.String())
	}

	// Sort
	output.WriteString(term.Sort.String())

	return output.String()
}

func canonicalDieString(die Die) string {
	switch d := die.(type) {
	case NormalDie:
		return fmt.Sprintf("d%d", d)
	case PercentileDie:
		return "d%"
	case FateDie:
		return "dF"
	default:
		return die.String()
	}
}

func canonicalizeGroupTerm(term GroupTerm, children []stackNode) string {
	parts := make([]stackNode, len(children))
	copy(parts, children)

	if term.Combined {
		// Sort combined children to make commutative additions deterministic
		sort.SliceStable(parts, func(i, j int) bool {
			return parts[i].sortKey < parts[j].sortKey
		})
	}

	rawParts := make([]string, len(parts))
	for i, p := range parts {
		rawParts[i] = p.rendered
	}

	sep := ", "
	if term.Combined {
		sep = " + "
	}

	output := strings.Join(rawParts, sep)
	if term.Combined {
		output = strings.ReplaceAll(output, "+-", "-")
	} else if len(parts) == 1 {
		output += ","
	}

	output = "{" + output + "}"
	output = strings.ReplaceAll(output, "{+", "{")
	output = strings.ReplaceAll(output, "{-", "{")
	output = strings.ReplaceAll(output, ", +", ", ")
	output = strings.ReplaceAll(output, ", -", ", ")
	output = strings.ReplaceAll(output, "+ +", "+ ")
	output = strings.ReplaceAll(output, "+ -", "- ")

	if term.Limit != nil {
		output += term.Limit.String()
	}
	if term.Success != nil {
		output += term.Success.String()
	}
	if term.Failure != nil {
		output += "f" + term.Failure.String()
	}
	if term.Modifier != 0 {
		output += fmt.Sprintf("%+d", term.Modifier)
	}
	if term.Negative {
		output = "-" + output
	}

	return output
}
