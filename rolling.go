package roll

import (
	"fmt"
	"io"
	"strings"
)

// ParseString takes a string and does a dice roll
func ParseString(rollStr string) (string, error) {
	return Parse(strings.NewReader(rollStr))
}

// Parse reads from an io.Reader and generates a dice roll result string
func Parse(r io.Reader) (string, error) {
	parser := NewParser(r)
	stmt, err := parser.Parse()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Rolled %q and got ", strings.TrimPrefix(stmt.String(), "+"))

	results := stmt.Roll()
	for _, result := range results.Results {
		output += result.Symbol + ", "
	}

	output = strings.TrimSuffix(output, ", ") + " for a total of %d"
	return fmt.Sprintf(output, results.Total), nil
}
