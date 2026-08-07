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
			Category:    "<",
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
	}
}

// ExtractTargetToken extracts the prefix token at or before the cursor position in input.
func ExtractTargetToken(input string, cursor int) string {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(input) {
		cursor = len(input)
	}

	sub := input[:cursor]
	if len(sub) == 0 {
		return ""
	}

	runes := []rune(sub)
	lastIndex := len(runes) - 1

	if unicode.IsDigit(runes[lastIndex]) || unicode.IsSpace(runes[lastIndex]) {
		return ""
	}

	var tokenRunes []rune
	for i := lastIndex; i >= 0; i-- {
		r := runes[i]
		if unicode.IsDigit(r) || unicode.IsSpace(r) {
			break
		}
		tokenRunes = append([]rune{r}, tokenRunes...)
	}

	return string(tokenRunes)
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
