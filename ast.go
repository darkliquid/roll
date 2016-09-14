package roll

import "fmt"

// DiceRollStmt is a Dice Rolling Statement
type DiceRollStmt struct {
	Multiplier int
	Die        Die
	Modifier   int
}

// Roll gets the results of rolling the dice that make up a dice roll
func (stmt *DiceRollStmt) Roll() (results []DieRoll) {
	results = make([]DieRoll, stmt.Multiplier)
	for i := 0; i < stmt.Multiplier; i++ {
		results[i] = stmt.Die.Roll()
	}
	return results
}

func (stmt *DiceRollStmt) String() string {
	var output string
	if stmt.Multiplier > 1 {
		output += fmt.Sprintf("%d", stmt.Multiplier)
	}

	output += stmt.Die.String()

	if stmt.Modifier != 0 {
		output += fmt.Sprintf("%+d", stmt.Modifier)
	}
	return output
}
