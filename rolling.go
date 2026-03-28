package roll

import (
	"fmt"
	"io"
	"strings"
)

// ParseString takes a string and does a dice roll
func ParseString(rollStr string) (string, error) {
	return ParseStringWithLimits(rollStr, DefaultLimits)
}

// ParseStringWithLimits takes a string and does a dice roll using explicit limits.
func ParseStringWithLimits(rollStr string, limits Limits) (string, error) {
	return ParseWithLimits(strings.NewReader(rollStr), limits)
}

// Parse reads from an io.Reader and generates a dice roll result string
func Parse(r io.Reader) (string, error) {
	return ParseWithLimits(r, DefaultLimits)
}

// ParseWithLimits reads from an io.Reader and generates a dice roll result string using explicit limits.
func ParseWithLimits(r io.Reader, limits Limits) (string, error) {
	parser := NewParserWithLimits(r, limits)
	stmt, err := parser.Parse()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Rolled %q and got ", strings.TrimPrefix(stmt.String(), "+"))

	results, err := EvaluateRollWithLimits(stmt, limits)
	if err != nil {
		return "", err
	}
	for _, result := range results.Results {
		output += result.Symbol + ", "
	}

	output = strings.TrimSuffix(output, ", ") + fmt.Sprintf(" for a total of %d", results.Total)
	return output, nil
}
