package roll

import "fmt"

// Node is a node in the AST
type Node interface {
	node()
	String() string
}

// ComparisonType is the type of comparison
type ComparisonType int

const (
	// Equals matches only values that are equal to the comparison value
	Equals ComparisonType = iota
	// GreaterThan matches only values greater than the comparison value
	GreaterThan
	// LessThan matches only values less than the comparison value
	LessThan
)

// ComparisonOp is the operation that defines how you compare against a roll
// to determine whether the result counts
type ComparisonOp struct {
	Type  ComparisonType
	Value int
}

// ExplodingType is the type of exploding die
type ExplodingType int

const (
	// Exploding adds new dice for each roll satisfying the exploding condition
	Exploding ExplodingType = iota
	// Compounded adds to a single new result for each roll
	Compounded
	// Penetrating is like Exploding, except each die result has a -1 modifier
	Penetrating
)

// ExplodingOp is the operation that defines how a dice roll explodes
type ExplodingOp struct {
	Type      ExplodingType
	Condition ComparisonOp
}

// LimitType is the type of roll limitation
type LimitType int

const (
	// KeepHighest indicated we should keep the highest results
	KeepHighest LimitType = iota
	// KeepLowest indicated we should keep the lowest results
	KeepLowest
	// DropHighest indicated we should drop the highest results
	DropHighest
	// DropLowest indicated we should drop the lowest results
	DropLowest
)

// LimitOp is the operation that defines how dice roll results are limited
type LimitOp struct {
	Type      LimitType
	Condition ComparisonOp
}

// RerollOp is the operation that defines how dice are rerolled
type RerollOp struct {
	ComparisonOp
	Once bool
}

// SortType is the type of sorting to use for dice roll results
type SortType int

const (
	// Ascending sorts dice rolls from lowest to highest
	Ascending SortType = iota
	// Descending sorts dice rolls from highest to lowest
	Descending
)

// DiceRollStmt is a Dice Rolling Statement
type DiceRollStmt struct {
	Multiplier int
	Die        Die
	Modifier   int
	Exploding  ExplodingOp
	Limit      LimitOp
	Success    ComparisonOp
	Failure    ComparisonOp
	Rerolls    []RerollOp
	Sort       SortType
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
