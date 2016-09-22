package roll

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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

// Match returns true if the given value compares positively against the op val
func (op ComparisonOp) Match(val int) bool {
	switch op.Type {
	case Equals:
		return val == op.Value
	case GreaterThan:
		return val > op.Value
	case LessThan:
		return val < op.Value
	}
	return false
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
	ComparisonOp
	Type ExplodingType
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
	Amount int
	Type   LimitType
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

// Roll is any kind of roll
type Roll interface {
	Roll() Result
	String() string
}

// Result is a collection of die rolls and a count of successes
type Result struct {
	Results   []DieRoll
	Total     int
	Successes int
}

// Len is the number of results
func (r *Result) Len() int {
	return len(r.Results)
}

// Less return true if DieRoll at index i is less than the one at index j
func (r *Result) Less(i, j int) bool {
	return r.Results[i].Result < r.Results[j].Result
}

// Swap swaps the DieRoll at index i with the one at index j
func (r *Result) Swap(i, j int) {
	r.Results[i], r.Results[j] = r.Results[j], r.Results[i]
}

// DiceRoll is an individual Dice Roll
type DiceRoll struct {
	Multiplier int
	Die        Die
	Modifier   int
	Exploding  *ExplodingOp
	Limit      *LimitOp
	Success    *ComparisonOp
	Failure    *ComparisonOp
	Rerolls    []RerollOp
	Sort       *SortType
}

// Roll gets the results of rolling the dice that make up a dice roll
func (dr *DiceRoll) Roll() (result Result) {
	// 1. Do Multiplier rolls of Die
	if dr.Multiplier == 0 {
		return
	}

	totalMultiplier := 1
	if dr.Multiplier < 0 {
		totalMultiplier = -1
	}

	for i := 0; i < dr.Multiplier*totalMultiplier; i++ {
		result.Results = append(result.Results, dr.Die.Roll())
	}

	// 2. For each result, check reroll criteria and reroll if a match
	for _, reroll := range dr.Rerolls {
		for i, roll := range result.Results {
			for reroll.Match(roll.Result) {
				roll = dr.Die.Roll()
			}
			result.Results[i] = roll
		}
	}

	// 3. For each result, check exploding criteria and generate new rolls
	if dr.Exploding != nil {
		switch dr.Exploding.Type {
		case Exploding:
			for _, roll := range result.Results {
				if dr.Exploding.Match(roll.Result) {
					roll = dr.Die.Roll()
					result.Results = append(result.Results, roll)
				}
			}
		case Compounded:
			compound := 0
			for _, roll := range result.Results {
				for dr.Exploding.Match(roll.Result) {
					compound += roll.Result
					roll = dr.Die.Roll()
				}
			}
			result.Results = append(result.Results, DieRoll{compound, strconv.Itoa(compound)})
		case Penetrating:
			for _, roll := range result.Results {
				if dr.Exploding.Match(roll.Result) {
					roll = dr.Die.Roll()
					roll.Result--
					roll.Symbol = strconv.Itoa(roll.Result)
					result.Results = append(result.Results, roll)
				}
			}
		}
	}

	// 4. Check results and apply limit operation
	applyLimit(dr.Limit, &result)

	// 5. If success op set, add modifier to each result and add successes for each match
	applySuccess(dr.Success, dr.Modifier, &result)

	// 6. If failure op set, add modifier to each result and subtract successes for each match
	applyFailure(dr.Failure, dr.Modifier, &result)

	// 7. If sort op set, sort results
	applySort(dr.Sort, &result)

	// 8. If success and failure ops not set, add modifier to total result
	finaliseTotals(dr.Success, dr.Failure, dr.Modifier, totalMultiplier, &result)

	return
}

// String represents the dice roll as a string
func (dr *DiceRoll) String() string {
	var output string
	if dr.Multiplier > 1 {
		output += fmt.Sprintf("%d", dr.Multiplier)
	}

	output += dr.Die.String()

	if dr.Modifier != 0 {
		output += fmt.Sprintf("%+d", dr.Modifier)
	}
	return output
}

// GroupedRoll is a group of other rolls. You can have nested groups.
type GroupedRoll struct {
	Rolls    []Roll
	Modifier int
	Limit    *LimitOp
	Success  *ComparisonOp
	Failure  *ComparisonOp
	Combined bool
}

// Roll gets the results of rolling the dice that make up a dice roll
func (gr *GroupedRoll) Roll() (result Result) {
	// 1. Generate results for each roll
	for _, roll := range gr.Rolls {
		if gr.Combined {
			// 2. If combined, merge all roll results into one result set
			result.Results = append(result.Results, roll.Roll().Results...)
		} else {
			// 3. If not combined, make new result set out of the totals for each roll
			total := roll.Roll().Total
			result.Results = append(result.Results, DieRoll{total, strconv.Itoa(total)})
		}
	}

	// 4. If limit set, apply limit operation to results
	applyLimit(gr.Limit, &result)

	// 5. If Success set, apply success op to results
	applySuccess(gr.Success, gr.Modifier, &result)

	// 6. If Failure set, apply failure op to results
	applyFailure(gr.Failure, gr.Modifier, &result)

	// 7. Add modifier or tally successes
	finaliseTotals(gr.Success, gr.Failure, gr.Modifier, 1, &result)

	return result
}

// String represents the grouped roll as a string
func (gr *GroupedRoll) String() string {
	parts := []string{"{"}
	for _, roll := range gr.Rolls {
		parts = append(parts, roll.String())
	}
	parts = append(parts, "}")

	sep := ", "
	if gr.Combined {
		sep = "+"
	}

	output := strings.Join(parts, sep)
	if gr.Combined {
		output = strings.Replace(output, "+-", "-", -1)
	}

	return output
}

func applyLimit(limitOp *LimitOp, result *Result) {
	if limitOp != nil {
		var rolls Result
		rolls.Results = result.Results[:]

		// Sort our tmp result copy
		sort.Sort(&rolls)

		// Work out limit
		limit := limitOp.Amount
		if limit > len(rolls.Results) {
			limit = len(rolls.Results)
		}

		switch limitOp.Type {
		case KeepHighest:
			result.Results = rolls.Results[len(rolls.Results)-limit:]
		case KeepLowest:
			result.Results = rolls.Results[:limit]
		case DropHighest:
			result.Results = rolls.Results[:len(rolls.Results)-limit]
		case DropLowest:
			result.Results = rolls.Results[limit:]
		}
	}
}

func applySuccess(successOp *ComparisonOp, modifier int, result *Result) {
	if successOp != nil {
		for _, roll := range result.Results {
			if successOp.Match(roll.Result + modifier) {
				result.Successes++
			}
		}
	}
}

func applyFailure(failureOp *ComparisonOp, modifier int, result *Result) {
	if failureOp != nil {
		for _, roll := range result.Results {
			if failureOp.Match(roll.Result + modifier) {
				result.Successes--
			}
		}
	}
}

func applySort(sortType *SortType, result *Result) {
	if sortType != nil {
		switch *sortType {
		case Ascending:
			sort.Sort(result)
		case Descending:
			sort.Sort(sort.Reverse(result))
		}
	}
}

func finaliseTotals(successOp, failureOp *ComparisonOp, modifier, multiplier int, result *Result) {
	if successOp == nil && failureOp == nil {
		for _, roll := range result.Results {
			result.Total += roll.Result
		}
		result.Total += modifier
		result.Total *= multiplier
	} else {
		result.Total = result.Successes
	}
}
