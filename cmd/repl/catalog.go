package main

import (
	"strings"
	"unicode"
)

// Component represents a dice syntax component for autocompletion and help text.
type Component struct {
	Name        string
	Syntax      string
	Template    string
	Category    string
	Description string
	Example     string
}

// DefaultCatalog returns the list of supported dice syntax components.
func DefaultCatalog() []Component {
	return []Component{
		{
			Name:        "Standard Die",
			Syntax:      "d6",
			Template:    "d6",
			Category:    "Dice",
			Description: "Roll a 6-sided die (or any N sides)",
			Example:     "4d6",
		},
		{
			Name:        "Percentile Die",
			Syntax:      "d%",
			Template:    "d%",
			Category:    "Dice",
			Description: "Roll a 100-sided percentile die",
			Example:     "1d%",
		},
		{
			Name:        "Fudge/Fate Die",
			Syntax:      "dF",
			Template:    "dF",
			Category:    "Dice",
			Description: "Roll a Fudge/Fate die (-1, 0, +1)",
			Example:     "4dF",
		},
		{
			Name:        "Explode",
			Syntax:      "!",
			Template:    "!",
			Category:    "Exploding",
			Description: "Explode dice on max value",
			Example:     "4d6!",
		},
		{
			Name:        "Compound Explode",
			Syntax:      "!!",
			Template:    "!!",
			Category:    "Exploding",
			Description: "Compound exploding dice",
			Example:     "4d6!!",
		},
		{
			Name:        "Penetrating Explode",
			Syntax:      "!p",
			Template:    "!p",
			Category:    "Exploding",
			Description: "Penetrating exploding dice (-1 on exploded)",
			Example:     "4d6!p",
		},
		{
			Name:        "Keep Highest",
			Syntax:      "kh",
			Template:    "kh",
			Category:    "Keep/Drop",
			Description: "Keep highest N dice",
			Example:     "4d6kh3",
		},
		{
			Name:        "Keep Lowest",
			Syntax:      "kl",
			Template:    "kl",
			Category:    "Keep/Drop",
			Description: "Keep lowest N dice",
			Example:     "4d6kl1",
		},
		{
			Name:        "Drop Highest",
			Syntax:      "dh",
			Template:    "dh",
			Category:    "Keep/Drop",
			Description: "Drop highest N dice",
			Example:     "4d6dh1",
		},
		{
			Name:        "Drop Lowest",
			Syntax:      "dl",
			Template:    "dl",
			Category:    "Keep/Drop",
			Description: "Drop lowest N dice",
			Example:     "4d6dl1",
		},
		{
			Name:        "Reroll",
			Syntax:      "r",
			Template:    "r",
			Category:    "Reroll",
			Description: "Reroll dice matching condition",
			Example:     "4d6r1",
		},
		{
			Name:        "Reroll Once",
			Syntax:      "ro",
			Template:    "ro",
			Category:    "Reroll",
			Description: "Reroll dice matching condition once",
			Example:     "4d6ro1",
		},
		{
			Name:        "Target Greater Than",
			Syntax:      ">",
			Template:    ">",
			Category:    "Target/Success",
			Description: "Count successes greater than target",
			Example:     "4d6>4",
		},
		{
			Name:        "Target Less Than",
			Syntax:      "<",
			Template:    "<",
			Category:    "Target/Success",
			Description: "Count successes less than target",
			Example:     "4d6<3",
		},
		{
			Name:        "Target Equal",
			Syntax:      "=",
			Template:    "=",
			Category:    "Target/Success",
			Description: "Count successes equal to target",
			Example:     "4d6=6",
		},
		{
			Name:        "Failure Threshold",
			Syntax:      "f",
			Template:    "f",
			Category:    "Target/Success",
			Description: "Subtract failures matching condition",
			Example:     "4d6>4f1",
		},
		{
			Name:        "Sort Ascending",
			Syntax:      "s",
			Template:    "s",
			Category:    "Sorting",
			Description: "Sort rolled dice in ascending order",
			Example:     "4d6s",
		},
		{
			Name:        "Sort Descending",
			Syntax:      "sd",
			Template:    "sd",
			Category:    "Sorting",
			Description: "Sort rolled dice in descending order",
			Example:     "4d6sd",
		},
		{
			Name:        "Group Sub-expressions",
			Syntax:      "{}",
			Template:    "{}",
			Category:    "Grouping",
			Description: "Group multiple expressions together",
			Example:     "{4d6, 2d8}",
		},
		{
			Name:        "Multiplication",
			Syntax:      "*",
			Template:    "*",
			Category:    "Math",
			Description: "Multiply a roll or number",
			Example:     "2d6*3",
		},
		{
			Name:        "Division",
			Syntax:      "/",
			Template:    "/",
			Category:    "Math",
			Description: "Divide a roll or number",
			Example:     "2d6/2",
		},
		{
			Name:        "Modulus",
			Syntax:      "%",
			Template:    "%",
			Category:    "Math",
			Description: "Remainder after division",
			Example:     "2d6%3",
		},
		{
			Name:        "Exponentiation",
			Syntax:      "**",
			Template:    "**",
			Category:    "Math",
			Description: "Raise a value to a power",
			Example:     "2**3",
		},
		{
			Name:        "Floor",
			Syntax:      "floor()",
			Template:    "floor()",
			Category:    "Math",
			Description: "Round toward negative infinity",
			Example:     "floor(11/2)",
		},
		{
			Name:        "Round",
			Syntax:      "round()",
			Template:    "round()",
			Category:    "Math",
			Description: "Round to the nearest whole number",
			Example:     "round(11/2)",
		},
		{
			Name:        "Ceil",
			Syntax:      "ceil()",
			Template:    "ceil()",
			Category:    "Math",
			Description: "Round toward positive infinity",
			Example:     "ceil(11/2)",
		},
		{
			Name:        "Abs",
			Syntax:      "abs()",
			Template:    "abs()",
			Category:    "Math",
			Description: "Absolute value",
			Example:     "abs(-3)",
		},
		{
			Name:        "Dice Matching",
			Syntax:      "mt",
			Template:    "mt",
			Category:    "Matching",
			Description: "Count matches (Yahtzee style)",
			Example:     "20d6mt",
		},
	}
}

// ExtractTargetToken extracts the prefix token at or before the cursor position in input.
func ExtractTargetToken(input string, cursor int) string {
	if input == "" {
		return ""
	}
	runes := []rune(input)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	if cursor == 0 {
		return ""
	}

	if unicode.IsDigit(runes[cursor-1]) || unicode.IsSpace(runes[cursor-1]) {
		return ""
	}

	start := cursor - 1
	for start >= 0 {
		ch := runes[start]
		if unicode.IsDigit(ch) || unicode.IsSpace(ch) || ch == '+' || ch == '-' || ch == ',' {
			start++
			break
		}
		if start == 0 {
			break
		}
		start--
	}
	if start < 0 {
		start = 0
	}
	return string(runes[start:cursor])
}

// FilterComponents returns catalog components whose Syntax or Name starts with prefix token at cursor.
func FilterComponents(catalog []Component, input string, cursor int) []Component {
	token := ExtractTargetToken(input, cursor)
	if token == "" {
		return catalog
	}

	lowerToken := strings.ToLower(token)
	var filtered []Component
	for _, comp := range catalog {
		if strings.HasPrefix(strings.ToLower(comp.Syntax), lowerToken) || strings.HasPrefix(strings.ToLower(comp.Name), lowerToken) {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}
