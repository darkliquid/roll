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
	Type      ComparisonType
	Value     int
	Inclusive bool
}

// Match returns true if the given value compares positively against the op val
func (op *ComparisonOp) Match(val int) bool {
	switch op.Type {
	case Equals:
		return val == op.Value
	case GreaterThan:
		if op.Inclusive {
			return val >= op.Value
		}
		return val > op.Value
	case LessThan:
		if op.Inclusive {
			return val <= op.Value
		}
		return val < op.Value
	}
	return false
}

// String returns the string representation of the comparison operator
func (op *ComparisonOp) String() string {
	switch op.Type {
	case Equals:
		return fmt.Sprintf("=%d", op.Value)
	case GreaterThan:
		if op.Inclusive {
			return fmt.Sprintf(">=%d", op.Value)
		}
		return fmt.Sprintf(">%d", op.Value)
	case LessThan:
		if op.Inclusive {
			return fmt.Sprintf("<=%d", op.Value)
		}
		return fmt.Sprintf("<%d", op.Value)
	}

	return ""
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
	*ComparisonOp
	Type ExplodingType
}

// String returns the string representation of the exploding dice operation
func (e ExplodingOp) String() (output string) {
	switch e.Type {
	case Exploding:
		output = "!"
	case Compounded:
		output = "!!"
	case Penetrating:
		output = "!p"
	}

	return output + strings.TrimPrefix(e.ComparisonOp.String(), "=")
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

func (op LimitOp) String() (output string) {
	switch op.Type {
	case KeepHighest:
		output = "kh"
	case KeepLowest:
		output = "kl"
	case DropHighest:
		output = "dh"
	case DropLowest:
		output = "dl"
	}

	if op.Amount > 1 {
		output += strconv.Itoa(op.Amount)
	}

	return
}

// RerollOp is the operation that defines how dice are rerolled
type RerollOp struct {
	*ComparisonOp
	Once bool
}

// String returns the string representation of the exploding dice operation
func (e RerollOp) String() (output string) {
	output = "r"
	if e.Once {
		output = "ro"
	}

	return output + strings.TrimPrefix(e.ComparisonOp.String(), "=")
}

// SortType is the type of sorting to use for dice roll results
type SortType int

// String return the string representation of a SortType value
func (t SortType) String() string {
	switch t {
	case Unsorted:
		return ""
	case Ascending:
		return "s"
	case Descending:
		return "sd"
	}

	return ""
}

const (
	// Unsorted doesn't sort dice rolls
	Unsorted SortType = iota
	// Ascending sorts dice rolls from lowest to highest
	Ascending
	// Descending sorts dice rolls from highest to lowest
	Descending
)

// Roll is any kind of roll
type Roll interface {
	Roll() Result
	String() string
}

// Limits configures parser and evaluator safety limits.
type Limits struct {
	MaxDieSize     int
	MaxRollsPerDie int
	MaxRollsTotal  int
	MaxEvalDepth   int
}

// DefaultLimits are the package-level defaults used by Parse and ParseString.
var DefaultLimits = Limits{
	MaxDieSize:     1000000,
	MaxRollsPerDie: 1000000,
	MaxRollsTotal:  1000000,
	MaxEvalDepth:   256,
}

func (l Limits) normalized() Limits {
	if l.MaxDieSize <= 0 {
		l.MaxDieSize = DefaultLimits.MaxDieSize
	}
	if l.MaxRollsPerDie <= 0 {
		l.MaxRollsPerDie = DefaultLimits.MaxRollsPerDie
	}
	if l.MaxRollsTotal <= 0 {
		l.MaxRollsTotal = DefaultLimits.MaxRollsTotal
	}
	if l.MaxEvalDepth <= 0 {
		l.MaxEvalDepth = DefaultLimits.MaxEvalDepth
	}
	return l
}

// ErrUnsafeDie is raised when a die configuration is categorically unsafe.
type ErrUnsafeDie string

func (e ErrUnsafeDie) Error() string {
	return fmt.Sprintf("unsafe die type %q", string(e))
}

// ErrLimitExceeded is raised when parser or evaluator safety limits are exceeded.
type ErrLimitExceeded string

func (e ErrLimitExceeded) Error() string { return string(e) }

type rollContext struct {
	limits     Limits
	totalRolls int
	depth      int
}

func (ctx *rollContext) enter() error {
	ctx.depth++
	if ctx.depth > ctx.limits.MaxEvalDepth {
		return ErrLimitExceeded(fmt.Sprintf("roll exceeded maximum evaluation depth of %d", ctx.limits.MaxEvalDepth))
	}
	return nil
}

func (ctx *rollContext) leave() {
	if ctx.depth > 0 {
		ctx.depth--
	}
}

func (ctx *rollContext) recordRoll(perDie *int) error {
	(*perDie)++
	if *perDie > ctx.limits.MaxRollsPerDie {
		return ErrLimitExceeded(fmt.Sprintf("die term exceeded maximum roll count of %d", ctx.limits.MaxRollsPerDie))
	}

	ctx.totalRolls++
	if ctx.totalRolls > ctx.limits.MaxRollsTotal {
		return ErrLimitExceeded(fmt.Sprintf("roll exceeded maximum total roll count of %d", ctx.limits.MaxRollsTotal))
	}
	return nil
}

// EvaluateRoll executes a parsed roll using DefaultLimits.
func EvaluateRoll(roll Roll) (Result, error) {
	return EvaluateRollWithLimits(roll, DefaultLimits)
}

// EvaluateRollWithLimits executes a parsed roll using explicit safety limits.
func EvaluateRollWithLimits(roll Roll, limits Limits) (Result, error) {
	ctx := &rollContext{limits: limits.normalized()}
	return evalRollWithContext(ctx, roll)
}

func evalRollWithContext(ctx *rollContext, roll Roll) (Result, error) {
	switch r := roll.(type) {
	case *DiceRoll:
		return r.rollWithContext(ctx)
	case *GroupedRoll:
		return r.rollWithContext(ctx)
	default:
		return roll.Roll(), nil
	}
}

func validateDieLimits(die Die, limits Limits) error {
	limits = limits.normalized()

	var size int
	switch d := die.(type) {
	case NormalDie:
		size = int(d)
	case PercentileDie:
		size = 100
	case FateDie:
		size = 3
	default:
		return nil
	}

	if size < 2 {
		return ErrUnsafeDie(die.String())
	}
	if size > limits.MaxDieSize {
		return ErrLimitExceeded(fmt.Sprintf("die size %d exceeds maximum %d", size, limits.MaxDieSize))
	}
	return nil
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
	Sort       SortType
}

// Roll gets the results of rolling the dice that make up a dice roll
func (dr *DiceRoll) Roll() (result Result) {
	result, _ = dr.rollWithContext(nil)
	return
}

func (dr *DiceRoll) rollWithContext(ctx *rollContext) (result Result, err error) {
	if ctx != nil {
		if err = ctx.enter(); err != nil {
			return
		}
		defer ctx.leave()
		if err = validateDieLimits(dr.Die, ctx.limits); err != nil {
			return
		}
	}

	// 1. Do Multiplier rolls of Die
	if dr.Multiplier == 0 {
		return
	}

	totalMultiplier := 1
	if dr.Multiplier < 0 {
		totalMultiplier = -1
	}

	rollCount := dr.Multiplier * totalMultiplier
	if ctx != nil && rollCount > ctx.limits.MaxRollsPerDie {
		return Result{}, ErrLimitExceeded(fmt.Sprintf("die term exceeded maximum roll count of %d", ctx.limits.MaxRollsPerDie))
	}

	dieRolls := 0
	for i := 0; i < rollCount; i++ {
		if ctx != nil {
			if err = ctx.recordRoll(&dieRolls); err != nil {
				return Result{}, err
			}
		}
		result.Results = append(result.Results, dr.Die.Roll())
	}

	// 2. For each result, check reroll criteria and reroll if a match
	for i, roll := range result.Results {
	RerollOnce:
		for _, reroll := range dr.Rerolls {
			for reroll.Match(roll.Result) {
				if ctx != nil {
					if err = ctx.recordRoll(&dieRolls); err != nil {
						return Result{}, err
					}
				}
				roll = dr.Die.Roll()
				result.Results[i] = roll
				if reroll.Once {
					break RerollOnce
				}
			}
		}
	}

	// 3. For each result, check exploding criteria and generate new rolls
	if dr.Exploding != nil {
		switch dr.Exploding.Type {
		case Exploding:
			for _, roll := range result.Results {
				for dr.Exploding.Match(roll.Result) {
					if ctx != nil {
						if err = ctx.recordRoll(&dieRolls); err != nil {
							return Result{}, err
						}
					}
					roll = dr.Die.Roll()
					result.Results = append(result.Results, roll)
				}
			}
		case Compounded:
			compound := 0
			for _, roll := range result.Results {
				for dr.Exploding.Match(roll.Result) {
					compound += roll.Result
					if ctx != nil {
						if err = ctx.recordRoll(&dieRolls); err != nil {
							return Result{}, err
						}
					}
					roll = dr.Die.Roll()
				}
			}
			result.Results = append(result.Results, DieRoll{compound, strconv.Itoa(compound)})
		case Penetrating:
			for _, roll := range result.Results {
				for dr.Exploding.Match(roll.Result) {
					if ctx != nil {
						if err = ctx.recordRoll(&dieRolls); err != nil {
							return Result{}, err
						}
					}
					roll = dr.Die.Roll()
					newroll := roll
					newroll.Result--
					newroll.Symbol = strconv.Itoa(newroll.Result)
					result.Results = append(result.Results, newroll)
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
	var output strings.Builder
	if dr.Multiplier > 1 || dr.Multiplier < -1 {
		output.WriteString(fmt.Sprintf("%+d", dr.Multiplier))
	} else if dr.Multiplier == -1 {
		output.WriteString("-")
	} else if dr.Multiplier == 1 {
		output.WriteString("+")
	}

	output.WriteString(dr.Die.String())

	if dr.Modifier != 0 {
		output.WriteString(fmt.Sprintf("%+d", dr.Modifier))
	}

	for _, r := range dr.Rerolls {
		output.WriteString(r.String())
	}

	if dr.Exploding != nil {
		output.WriteString((*dr.Exploding).String())
	}

	if dr.Limit != nil {
		output.WriteString((*dr.Limit).String())
	}

	if dr.Success != nil {
		output.WriteString((*dr.Success).String())
	}

	if dr.Failure != nil {
		output.WriteString("f" + (*dr.Failure).String())
	}

	output.WriteString(dr.Sort.String())

	return output.String()
}

// GroupedRoll is a group of other rolls. You can have nested groups.
type GroupedRoll struct {
	Rolls    []Roll
	Modifier int
	Limit    *LimitOp
	Success  *ComparisonOp
	Failure  *ComparisonOp
	Combined bool
	Negative bool
}

// Roll gets the results of rolling the dice that make up a dice roll
func (gr *GroupedRoll) Roll() (result Result) {
	result, _ = gr.rollWithContext(nil)
	return
}

func (gr *GroupedRoll) rollWithContext(ctx *rollContext) (result Result, err error) {
	if ctx != nil {
		if err = ctx.enter(); err != nil {
			return
		}
		defer ctx.leave()
	}

	// 1. Generate results for each roll
	for _, roll := range gr.Rolls {
		childResult, childErr := evalRollWithContext(ctx, roll)
		if childErr != nil {
			return Result{}, childErr
		}
		if gr.Combined {
			// 2. If combined, merge all roll results into one result set
			// NOTE: in combined mode, the roll modifier is added to each result
			var mod int
			switch t := roll.(type) {
			case *GroupedRoll:
				mod = t.Modifier
			case *DiceRoll:
				mod = t.Modifier
			}

			for _, res := range childResult.Results {
				result.Results = append(result.Results, DieRoll{
					res.Result + mod,
					strconv.Itoa(res.Result + mod),
				})
			}
		} else {
			// 3. If not combined, make new result set out of the totals for each roll
			result.Results = append(result.Results, DieRoll{childResult.Total, strconv.Itoa(childResult.Total)})
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

	if gr.Negative {
		result.Total *= -1
		for i, r := range result.Results {
			r.Result *= -1
			result.Results[i] = r
		}
	}

	return result, nil
}

// String represents the grouped roll as a string
func (gr *GroupedRoll) String() (output string) {
	parts := []string{}

	for _, roll := range gr.Rolls {
		if roll != nil {
			parts = append(parts, roll.String())
		}
	}

	sep := ", "
	if gr.Combined {
		sep = " + "
	}

	output = strings.Join(parts, sep)
	if gr.Combined {
		output = strings.Replace(output, "+-", "-", -1)
	} else if len(gr.Rolls) == 1 {
		// This case should be impossible, but we want to be able to identify
		// it if it *does* somehow happen.
		output += ","
	}

	output = "{" + output + "}"
	output = strings.Replace(output, "{+", "{", -1)
	output = strings.Replace(output, "{-", "{", -1)
	output = strings.Replace(output, ", +", ", ", -1)
	output = strings.Replace(output, ", -", ", ", -1)
	output = strings.Replace(output, "+ +", "+ ", -1)
	output = strings.Replace(output, "+ -", "- ", -1)

	if gr.Limit != nil {
		output += (*gr.Limit).String()
	}

	if gr.Success != nil {
		output += (*gr.Success).String()
	}

	if gr.Failure != nil {
		output += "f" + (*gr.Failure).String()
	}

	if gr.Modifier != 0 {
		output += fmt.Sprintf("%+d", gr.Modifier)
	}

	if gr.Negative {
		output = "-" + output
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
		limit := min(limitOp.Amount, len(rolls.Results))

		switch limitOp.Type {
		case KeepHighest:
			rolls.Results = rolls.Results[len(rolls.Results)-limit:]
		case KeepLowest:
			rolls.Results = rolls.Results[:limit]
		case DropHighest:
			rolls.Results = rolls.Results[:len(rolls.Results)-limit]
		case DropLowest:
			rolls.Results = rolls.Results[limit:]
		}

		m := make(map[int]int, len(rolls.Results))
		for _, r := range rolls.Results {
			m[r.Result]++
		}

		newResults := make([]DieRoll, 0, len(rolls.Results))
		for _, a := range result.Results {
			if b, ok := m[a.Result]; ok {
				newResults = append([]DieRoll{a}, newResults...)
				b--
				if b == 0 {
					delete(m, a.Result)
				}
			}
		}

		result.Results = newResults
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

func applySort(sortType SortType, result *Result) {
	switch sortType {
	case Unsorted:
		return
	case Ascending:
		sort.Sort(result)
	case Descending:
		sort.Sort(sort.Reverse(result))
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
