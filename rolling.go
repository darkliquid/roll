package roll

import (
	"fmt"
	"io"
	"strings"
)

// CompileString compiles a dice expression string into a VM program.
func CompileString(rollStr string) (*Program, error) {
	return CompileStringWithLimits(rollStr, DefaultLimits)
}

// CompileStringWithLimits compiles a dice expression string using explicit limits.
func CompileStringWithLimits(rollStr string, limits Limits) (*Program, error) {
	return CompileWithLimits(strings.NewReader(rollStr), limits)
}

// Compile reads from an io.Reader and compiles a VM program.
func Compile(r io.Reader) (*Program, error) {
	return CompileWithLimits(r, DefaultLimits)
}

// CompileWithLimits reads from an io.Reader and compiles a VM program using explicit limits.
func CompileWithLimits(r io.Reader, limits Limits) (*Program, error) {
	parser := NewParserWithLimits(r, limits)
	return parser.Parse()
}

// ParseString takes a string and executes a dice roll.
func ParseString(rollStr string) (string, error) {
	return ParseStringWithLimits(rollStr, DefaultLimits)
}

// ParseStringWithLimits takes a string and executes a dice roll using explicit limits.
func ParseStringWithLimits(rollStr string, limits Limits) (string, error) {
	return ParseWithLimits(strings.NewReader(rollStr), limits)
}

// Parse reads from an io.Reader and generates a dice roll result string.
func Parse(r io.Reader) (string, error) {
	return ParseWithLimits(r, DefaultLimits)
}

// ParseWithLimits reads from an io.Reader and generates a dice roll result string using explicit limits.
func ParseWithLimits(r io.Reader, limits Limits) (string, error) {
	program, err := CompileWithLimits(r, limits)
	if err != nil {
		return "", err
	}

	results, err := EvaluateProgramWithLimits(program, limits)
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Rolled %q and got ", program.String())
	for _, result := range results.Results {
		output += result.Symbol + ", "
	}

	output = strings.TrimSuffix(output, ", ") + fmt.Sprintf(" for a total of %d", results.Total)
	return output, nil
}
